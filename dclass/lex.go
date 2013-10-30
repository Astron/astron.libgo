package dclass

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// token represents a token or text string returned from the scanner.
type token struct {
	typ tokenType // The type of this token.
	pos int       // The starting position, in bytes, of this token in the input string.
	val string    // The value of this token.
}

// implements stringer
func (t token) String() string {
	switch {
	case t.typ == tokenEOF:
		return "EOF"
	case t.typ == tokenError:
		return t.val
	case t.typ > tokenKeyDelim:
		return fmt.Sprintf("<%s>", t.val)
	case len(t.val) > 10:
		return fmt.Sprintf("%.10q...", t.val)
	}
	return fmt.Sprintf("%q", t.val)
}

// tokenType identifies the type of lex tokens.
type tokenType int

const (
	// Lexer types
	tokenError tokenType = iota // error occurred; value is text of error
	tokenEOF                    // end of file

	// Value types
	tokenBool    // boolean constant
	tokenNumber  // simple number
	tokenRawchar // quoted character (quotes included)
	tokenQuote   // quoted string (quotes included)

	// Parser Types
	tokenIdentifier  // alphanumeric identifier
	tokenOperator    // one of '+', '-', '*', '/', or '%'
	tokenLeftParen   // a left paren '(', opens function arguments or arithmetic clause
	tokenRightParen  // a right paren ')', closes function arguments or arithemtic clause
	tokenLeftCurly   // a left curly '{', opens a block
	tokenRightCurly  // a right curly '}', closes a block
	tokenLeftSquare  // a left square '[', opens a sized array type defintion
	tokenRightSquare // a right square ']', closes a sized array type definition
	tokenComposition // a colon ':', indicates a list of components following
	tokenEndline     // a semicolon ';', indicates the end of a statement
	tokenSeperator   // a comma ',' used to seperate arguments or components
	tokenAssignment  // an equality sign '=', indicates assignment (for defaults)
	tokenVarArray    // a pair of square brackets "[]", indicating an unsized array type

	// Keyword types
	tokenKeyDelim // used only to delimit the keywords
	tokenKeyword  // 'keyword' keyword
	tokenDClass   // 'dclass' keyword
	tokenStruct   // 'struct' keyword

	// Variable-type keyword types
	tokenTypeDelim // used only to delimit the data type keywords
	tokenInt8      // signed 8-bit int keyword
	tokenInt16     // signed 16-bit int keyword
	tokenInt32     // signed 32-bit int keyword
	tokenInt64     // signed 64-bit int keyword
	tokenUint8     // unsigned 8-bit int keyword
	tokenUint16    // unsigned 16-bit int keyword
	tokenUint32    // unsigned 32-bit int keyword
	tokenUint64    // unsigned 64-bit int keyword
	tokenFloat     // 64-bit floating point keyword
	tokenString    // string keyword
	tokenBlob      // blob keyword
	tokenChar      // char keyword
)

var key = map[string]tokenType{
	// declarations
	"keyword": tokenKeyword,
	"dclass":  tokenDClass,
	"struct":  tokenStruct,

	// variable types
	"int8":    tokenInt8,
	"int16":   tokenInt16,
	"int32":   tokenInt32,
	"int64":   tokenInt64,
	"uint8":   tokenUint8,
	"uint16":  tokenUint16,
	"uint32":  tokenUint32,
	"uint64":  tokenUint64,
	"float64": tokenFloat,
	"string":  tokenString,
	"blob":    tokenBlob,
	"char":    tokenChar,
}

// Make the types prettyprint.
var tokenName = map[tokenType]string{
	tokenError: "error",
	tokenEOF:   "EOF",

	tokenBool:    "bool",
	tokenRawchar: "char-constant",
	tokenNumber:  "number",
	tokenQuote:   "quoted-string",

	tokenIdentifier:  "identifier",
	tokenOperator:    "<op>",
	tokenLeftParen:   "(",
	tokenRightParen:  ")",
	tokenLeftCurly:   "{",
	tokenRightCurly:  "}",
	tokenComposition: ":",
	tokenEndline:     ";",
	tokenSeperator:   ",",
	tokenAssignment:  "=",
	tokenVarArray:    "[]",

	tokenKeyword: "keyword",
	tokenDClass:  "dclass",
	tokenStruct:  "struct",

	tokenInt8:   "int8",
	tokenInt16:  "int16",
	tokenInt32:  "int32",
	tokenInt64:  "int64",
	tokenUint8:  "uint8",
	tokenUint16: "uint16",
	tokenUint32: "uint32",
	tokenUint64: "uint65",
	tokenFloat:  "float64",
	tokenString: "string",
	tokenBlob:   "blob",
	tokenChar:   "char",
}

func (t tokenType) String() string {
	s := tokenName[t]
	if s == "" {
		return fmt.Sprintf("token%d", int(t))
	}
	return s
}

const eof = -1

// lexerFn represents the state of the scanner as a function that returns the next state.
type lexerFn func(*lexer) lexerFn

// lexer holds the state of the scanner.
type lexer struct {
	input       string  // the string being scanned
	state       lexerFn // the next lexing function to enter
	pos         int     // current position in the input
	start       int     // start position of this token
	width       int     // width of last rune read from input
	parenDepth  int     // nesting depth of ( ) exprs
	curlyDepth  int     // nesting depth of { } blocks
	squareDepth int     // nesting depth of [ ] arrays -- should never be > 1
	canBackup   bool    // if backup has been called for this rune

	// variables for lexer token output
	lastPos     int        // position of most recent token returned by nextToken
	tokens      chan token // channel of scanned tokens
	peekedToken token      // the previous token to come out (for peek)
	hasPeeked   bool       // if true peekedToken is the next token
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

// emit passes an token back to the client.
func (l *lexer) emit(t tokenType) {
	l.tokens <- token{t, l.start, l.input[l.start:l.pos]}
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
	} // accept runes in the set of 'valid' until false

	l.backup()
}

// lineNumber reports which line we're on, based on the position of
// the previous token returned by nextToken. Doing it this way
// means we don't have to worry about peek double counting.
func (l *lexer) lineNumber() int {
	return 1 + strings.Count(l.input[:l.lastPos], "\n")
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextToken.
func (l *lexer) errorf(format string, args ...interface{}) lexerFn {
	l.tokens <- token{tokenError, l.start, fmt.Sprintf(format, args...)}
	return nil
}

// nextToken returns the next token from the input.
func (l *lexer) nextToken() token {
	var t token
	if l.hasPeeked {
		l.hasPeeked = false
		t = l.peekedToken
	} else {
		t = <-l.tokens
	}

	l.lastPos = t.pos
	return t
}

// peekToken returns but does not consume the next token from the input.
func (l *lexer) peekToken() token {
	if !l.hasPeeked {
		l.hasPeeked = true
		l.peekedToken = <-l.tokens
	}
	return l.peekedToken
}

// lex creates a new scanner for the input string.
func lex(input string) *lexer {
	l := &lexer{
		input:  input,
		tokens: make(chan token),
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

// lexAny scans for any token (keyword, struct, or dclass)
func lexAny(l *lexer) lexerFn {
	if l.squareDepth > 0 {
		r := l.next()
		if r == ']' {
			l.emit(tokenRightSquare)
		} else {
			return l.errorf("found opening '[' without matching ']' for array type definition.")
		}
	}
	switch r := l.next(); {
	case r == eof:
		if l.parenDepth > 0 {
			return l.errorf("unclosed left paren")
		}
		if l.curlyDepth > 0 {
			return l.errorf("unclosed right paren")
		}
		l.emit(tokenEOF)
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
			l.emit(tokenOperator)
		}
	// Must call before isOperator()
	case r == '=':
		l.emit(tokenAssignment)
	case isOperator(r):
		l.emit(tokenOperator)
	case r == ':':
		l.emit(tokenComposition)
	case r == ',':
		l.emit(tokenSeperator)
	case r == '(':
		l.emit(tokenLeftParen)
		l.parenDepth++
	case r == ')':
		l.emit(tokenRightParen)
		l.parenDepth--
		if l.parenDepth < 0 {
			return l.errorf("unexpected right paren %#U", r)
		}
	case r == '{':
		l.emit(tokenLeftCurly)
		l.curlyDepth++
	case r == '}':
		l.emit(tokenRightCurly)
		l.curlyDepth--
		if l.parenDepth < 0 {
			return l.errorf("unexpected right curly %#U", r)
		}
	case r == '"':
		return lexQuote
	case r == '\'':
		return lexChar
	case r == '[':
		rn := l.peek()
		if rn == '.' || ('0' <= rn && rn <= '9') {
			l.emit(tokenLeftSquare)
			l.squareDepth++
			return lexNumber
		} else if rn != ']' {
			return l.errorf("unexpected character: %#U, following '['", rn)
		}

		l.next() // consume "[]"
		l.emit(tokenVarArray)
	case r == ';':
		l.emit(tokenEndline)
	case isOperator(r):
		l.emit(tokenOperator)
	default:
		return l.errorf("unexpected character: %#U", r)
	}
	return lexAny
}

// lexComment scans a single-line comment. The left delimiter is pre-consumed.
func lexComment(l *lexer) lexerFn {
	i := strings.IndexAny(l.input[l.pos:], endlineChars)
	if i < 0 {
		return l.errorf("no new line after comment")
	}
	l.pos += i + len(rightComment)
	l.ignore()
	return lexAny
}

// lexBlockComment scans a block comment. The left delimiter is pre-consumed.
func lexBlockComment(l *lexer) lexerFn {
	i := strings.Index(l.input[l.pos:], rightBlockComment)
	if i < 0 {
		return l.errorf("unclosed block comment")
	}
	l.pos += i + len(rightBlockComment)
	l.ignore()
	return lexAny
}

// lexSpace scans a run of space and/or end-line characters.
// One space has already been seen.
func lexSpace(l *lexer) lexerFn {
	for isSpace(l.peek()) || isEndOfLine(l.peek()) {
		l.next()
	}
	l.ignore()
	return lexAny
}

// lexIdentifier scans an alphanumeric.
func lexIdentifier(l *lexer) lexerFn {
Loop:
	for {
		switch r := l.next(); {
		case isAlphaNumeric(r):
			// absorb.
		default:
			l.backup()
			word := l.input[l.start:l.pos]
			if !l.atTerminator() {
				return l.errorf("bad character in identifier %#U", r)
			}
			switch {
			case key[word] > tokenKeyDelim:
				l.emit(key[word])
			case word == "true", word == "false":
				l.emit(tokenBool)
			default:
				l.emit(tokenIdentifier)
			}
			break Loop
		}
	}
	return lexAny
}

// lexQuote scans a quoted string. The initial quote is already scanned.
func lexQuote(l *lexer) lexerFn {
Loop:
	for {
		switch l.next() {
		case '\\':
			if r := l.next(); r != eof && r != '\n' {
				break
			}
			fallthrough
		case eof, '\n':
			return l.errorf("unterminated string")
		case '"':
			break Loop
		}
	}
	l.emit(tokenQuote)
	return lexAny
}

// lexChar scans a character constant. The initial quote is already scanned.
func lexChar(l *lexer) lexerFn {
Loop:
	for {
		switch l.next() {
		case '\\':
			if r := l.next(); r != eof && r != '\n' {
				break
			}
			fallthrough
		case eof, '\n':
			return l.errorf("unterminated character constant")
		case '\'':
			break Loop
		}
	}
	l.emit(tokenRawchar)
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
func lexNumber(l *lexer) lexerFn {
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
			return l.errorf("non-decimal number (starting with '" + encode +
				"') cannot contain a decimal point.")
		}
		l.acceptRun(digits)
	}

	// Next thing must not be alphanumeric.
	if isAlphaNumeric(l.peek()) {
		l.next()
		return l.errorf("bad number syntax: %q", l.input[l.start:l.pos])
	}

	// Emit number
	l.emit(tokenNumber)

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

func isDataTypeToken(t token) bool {
	return t.typ > tokenTypeDelim
}
