package DCParser

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// item represents a token or text string returned from the scanner.
type item struct {
	typ itemType // The type of this item.
	pos int      // The starting position, in bytes, of this item in the input string.
	val string   // The value of this item.
}

func (i item) String() string {
	switch {
	case i.typ == itemEOF:
		return "EOF"
	case i.typ == itemError:
		return i.val
	case i.typ > itemKeyDelim:
		return fmt.Sprintf("<%s>", i.val)
	case len(i.val) > 10:
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

// itemType identifies the type of lex items.
type itemType int

const (
	// Lexer types
	itemError itemType = iota // error occurred; value is text of error
	itemEOF                   // end of file

	// Value types
	itemBool    // boolean constant
	itemNumber  // simple number
	itemRawchar // quoted character (quotes included)
	itemQuote   // quoted string (quotes included)

	// Parser Types
	itemIdentifier  // alphanumeric identifier
	itemOperator    // one of '+', '-', '*', '/', or '%'
	itemLeftParen   // a left paren '(', opens function arguments or arithmetic clause
	itemRightParen  // a left ')', closes function arguments or arithemtic clause
	itemLeftCurly   // a left curly '{', opens a block
	itemRightCurly  // a right curly '}', closes a block
	itemComposition // a colon ':', indicates a list of components following
	itemEndline     // a semicolon ';', indicates the end of a statement
	itemSeperator   // a comma ',' used to seperate arguments or components
	itemAssignment  // an equality sign '=', indicates assignment (for defaults)
	itemArray       // a pair of square brackets "[]", indicating an array type

	// Keyword types
	itemKeyDelim // used only to delimit the keywords
	itemKeyword  // 'keyword' keyword
	itemDClass   // 'dclass' keyword
	itemStruct   // 'struct' keyword

	// Variable-type keyword types
	itemInt8   // signed 8-bit int keyword
	itemInt16  // signed 16-bit int keyword
	itemInt32  // signed 32-bit int keyword
	itemInt64  // signed 64-bit int keyword
	itemUInt8  // unsigned 8-bit int keyword
	itemUInt16 // unsigned 16-bit int keyword
	itemUInt32 // unsigned 32-bit int keyword
	itemUInt64 // unsigned 64-bit int keyword
	itemFloat  // 64-bit floating point keyword
	itemString // string keyword
	itemBlob   // blob keyword
	itemChar   // char keyword
)

var key = map[string]itemType{
	// declarations
	"keyword": itemKeyword,
	"dclass":  itemDClass,
	"struct":  itemStruct,

	// variable types
	"int8":    itemInt8,
	"int16":   itemInt16,
	"int32":   itemInt32,
	"int64":   itemInt64,
	"uint8":   itemUInt8,
	"uint16":  itemUInt16,
	"uint32":  itemUInt32,
	"uint64":  itemUInt64,
	"float64": itemFloat,
	"string":  itemString,
	"blob":    itemBlob,
	"char":    itemChar,
}

const eof = -1

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	name       string    // the name of the input; used only for error reports
	input      string    // the string being scanned
	state      stateFn   // the next lexing function to enter
	pos        int       // current position in the input
	start      int       // start position of this item
	width      int       // width of last rune read from input
	lastPos    int       // position of most recent item returned by nextItem
	items      chan item // channel of scanned items
	parenDepth int       // nesting depth of ( ) exprs
	curlyDepth int       // nesting depth of { } blocks
	canBackup  bool      // if backup has been called for this rune
}

// next returns the next rune in the input.
func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w
	l.pos += l.width
	l.canBackup = true
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	if l.canBackup {
		l.pos -= l.width
		l.canBackup = false
	}
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.start, l.input[l.start:l.pos]}
	l.start = l.pos
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

// lineNumber reports which line we're on, based on the position of
// the previous item returned by nextItem. Doing it this way
// means we don't have to worry about peek double counting.
func (l *lexer) lineNumber() int {
	return 1 + strings.Count(l.input[:l.lastPos], "\n")
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, l.start, fmt.Sprintf(format, args...)}
	return nil
}

// nextItem returns the next item from the input.
func (l *lexer) nextItem() item {
	item := <-l.items
	l.lastPos = item.pos
	return item
}

// lex creates a new scanner for the input string.
func lex(name, input string) *lexer {
	l := &lexer{
		name:       name,
		input:      input,
		items:      make(chan item),
	}
	go l.run()
	return l
}

// run runs the state machine for the lexer.
func (l *lexer) run() {
	for l.state = lexAny; l.state != nil; {
		l.state = l.state(l)
	}
}

// state functions

const (
	endlineChars      = "\r\n"
	leftComment       = "//"
	leftBlockComment  = "/*"
	rightComment      = "\n" // Note, also accepts '\r'
	rightBlockComment = "*/"
)

// lexAny scans for any item (keyword, struct, or dclass)
func lexAny(l *lexer) stateFn {
	switch r := l.next(); {
	case r == eof:
		if l.parenDepth > 0 {
			return l.errorf("Unclosed left paren")
		}
		if l.curlyDepth > 0 {
			return l.errorf("Unclosed right paren")
		}
		l.emit(itemEOF)
		return nil
	case isSpace(r) || isEndOfLine(r):
		return lexSpace

	// Must call before isAlphanumeric()
	case r == '.' || ('0' <= r && r <= '9'):
		l.backup()
		return lexNumber
	case isAlphaNumeric(r):
		l.backup()
		return lexIdentifier

	// Must call before isOperator()
	case r == '/':
		if l.peek() == '/' {
			l.next()
			return lexComment
		} else if l.peek() == '*' {
			l.next()
			return lexBlockComment
		} else {
			l.emit(itemOperator)
		}
	// Must call before isOperator()
	case r == '=':
		l.emit(itemAssignment)
	case isOperator(r):
		l.emit(itemOperator)
	case r == ':':
		l.emit(itemComposition)
	case r == ',':
		l.emit(itemSeperator)
	case r == '(':
		l.emit(itemLeftParen)
		l.parenDepth++
	case r == ')':
		l.emit(itemRightParen)
		l.parenDepth--
		if l.parenDepth < 0 {
			return l.errorf("Unexpected right paren %#U", r)
		}
	case r == '{':
		l.emit(itemLeftCurly)
		l.curlyDepth++
	case r == '}':
		l.emit(itemRightCurly)
		l.curlyDepth--
		if l.parenDepth < 0 {
			return l.errorf("Unexpected right curly %#U", r)
		}
	case r == '"':
		return lexQuote
	case r == '\'':
		return lexChar
	case r == '[':
		if l.peek() != ']' {
			return l.errorf("Found opening '[' without matching ']' for array.")
		}
		l.next()
		l.emit(itemArray)
	case r == ';':
		l.emit(itemEndline)
	case isOperator(r):
		l.emit(itemOperator)
	default:
		return l.errorf("Unexpected character: %#U", r)
	}
	return lexAny
}

// lexComment scans a single-line comment. The left delimiter is pre-consumed.
func lexComment(l *lexer) stateFn {
	i := strings.IndexAny(l.input[l.pos:], endlineChars)
	if i < 0 {
		return l.errorf("No new line after comment")
	}
	l.pos += i + len(rightComment)
	l.ignore()
	return lexAny
}

// lexBlockComment scans a block comment. The left delimiter is pre-consumed.
func lexBlockComment(l *lexer) stateFn {
	i := strings.Index(l.input[l.pos:], rightBlockComment)
	if i < 0 {
		return l.errorf("Unclosed block comment")
	}
	l.pos += i + len(rightBlockComment)
	l.ignore()
	return lexAny
}

// lexSpace scans a run of space and/or end-line characters.
// One space has already been seen.
func lexSpace(l *lexer) stateFn {
	for isSpace(l.peek()) || isEndOfLine(l.peek()) {
		l.next()
	}
	l.ignore()
	return lexAny
}

// lexIdentifier scans an alphanumeric.
func lexIdentifier(l *lexer) stateFn {
Loop:
	for {
		switch r := l.next(); {
		case isAlphaNumeric(r):
			// absorb.
		default:
			l.backup()
			word := l.input[l.start:l.pos]
			if !l.atTerminator() {
				return l.errorf("Bad character in identifier %#U", r)
			}
			switch {
			case key[word] > itemKeyDelim:
				l.emit(key[word])
			case word == "true", word == "false":
				l.emit(itemBool)
			default:
				l.emit(itemIdentifier)
			}
			break Loop
		}
	}
	return lexAny
}

// lexQuote scans a quoted string. The initial quote is already scanned.
func lexQuote(l *lexer) stateFn {
Loop:
	for {
		switch l.next() {
		case '\\':
			if r := l.next(); r != eof && r != '\n' {
				break
			}
			fallthrough
		case eof, '\n':
			return l.errorf("Unterminated string")
		case '"':
			break Loop
		}
	}
	l.emit(itemQuote)
	return lexAny
}

// lexChar scans a character constant. The initial quote is already scanned.
func lexChar(l *lexer) stateFn {
Loop:
	for {
		switch l.next() {
		case '\\':
			if r := l.next(); r != eof && r != '\n' {
				break
			}
			fallthrough
		case eof, '\n':
			return l.errorf("Unterminated character constant")
		case '\'':
			break Loop
		}
	}
	l.emit(itemRawchar)
	return lexAny
}

const (
	decimalDigits     string = "0123456789"
	hexadecimalDigits        = "0123456789abcdefABCDEF"
	octalDigits              = "01234567"
	binaryDigits             = "01"

	decimalEncode     = ""
	hexadecimalEncode = "0x"
	octalEncode       = "0"
	binaryEncode      = "0b"
)

// lexNumber scans a number: decimal, octal, binary, hex, or float.
func lexNumber(l *lexer) stateFn {
	// Check if it is hex, octal, or binary
	encode := decimalEncode
	digits := decimalDigits
	if l.accept("0") {
		if l.accept("xX") {
			encode = hexadecimalEncode
			digits = hexadecimalDigits
		} else if l.accept("bB") {
			encode = binaryEncode
			digits = binaryDigits
		} else {
			encode = octalEncode
			digits = octalDigits
		}
	}

	// Read integer part of number
	l.acceptRun(digits)

	// Read fractional part of number
	if l.accept(".") {
		if digits != decimalDigits {
			return l.errorf("Non-decimal number (starting with '" + encode +
				"') cannot contain a decimal point.")
		}
		l.acceptRun(digits)
	}

	// Next thing must not be alphanumeric.
	if isAlphaNumeric(l.peek()) {
		l.next()
		return l.errorf("Bad number syntax: %q", l.input[l.start:l.pos])
	}

	// Emit number
	l.emit(itemNumber)

	return lexAny
}

// atTerminator reports whether the input is at valid termination character to
// appear after an identifier.
func (l *lexer) atTerminator() bool {
	r := l.peek()
	if isSpace(r) || isEndOfLine(r) || isOperator(r) {
		return true
	}
	switch r {
	case eof, ',', ':', ';', ')', '(', '{', '}', '[':
		return true
	}
	return false
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

// isEndOfLine reports whether r is an end-of-line character.
func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

// isOperator reports whether r is an operator or assignment character
func isOperator(r rune) bool {
	return r == '+' || r == '-' || r == '*' || r == '/' || r == '%' || r == '='
}
