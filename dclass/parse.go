package dclass

import (
	"bytes"
	"fmt"
	"io"
	"os"
	/*	"runtime"
		"strconv"
		"strings"
		"unicode"
	*/)

// Parse returns a pointer to a dclass File created by parsing the argument io.Reader.
// If one or more errors are encountered, a nil value is returned.
func Parse(r io.Reader) (dcf *File, err error) {
	// Load data from file
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)

	p := parser{
		dcf: new(File),
		lex: lex(buf.String()),
	}

	return p.parse(r), nil
}

// parser is a constructor for a single parsed dclass File
type parser struct {
	dcf *File  // dclass File being produced by parser
	lex *lexer // lexer to read tokens from

	// The expectedFoo fields are lists of identifiers that are expected for a declaration type, but
	// not yet declared. Each identifer in the lists maps to the line number where it was first used.
	expectedKeywords map[string]int
	expectedStructs  map[string]int
	expectedClasses  map[string]int

	// The expectingFoo fields are maps from an expected identifier to the list
	// of objects that are expecting it.
	expectingKeyword map[string][]Field  // field is qualified by the missing keyword
	expectingStruct  map[string][]Field  // field's datatype is defined as the missing struct
	expectingClass   map[string][]*Class // class inherits from the missing class or struct

	errors   []Error // errors encountered while parsing (including lexer errors)
	foundEOF bool    // whether next() has encountered an eof token
}

func (p parser) parse(r io.Reader) *File {
	// Parse declarations until EOF or lexer error
	for p.parseDeclaration() {
	}

	// Create errors if there are any expected identifiers remaining that have not been defined
	for keyword, firstLine := range p.expectedKeywords {
		p.errors = append(p.errors, definitionError(keyword, tokenKeyword, firstLine))
	}
	for structName, firstLine := range p.expectedStructs {
		p.errors = append(p.errors, definitionError(structName, tokenStruct, firstLine))
	}
	for className, firstLine := range p.expectedClasses {
		p.errors = append(p.errors, definitionError(className, tokenDClass, firstLine))
	}

	// If errors exist, print them and return nil
	if len(p.errors) > 0 {
		fmt.Fprintf(os.Stderr, "Encountered %d errors while parsing dclass file...\n", len(p.errors))

		// Print max 10 errors...
		for i := 0; i < 10 && i < len(p.errors); i++ {
			fmt.Fprintf(os.Stderr, " - %s\n", p.errors[i].Error())
		}

		// ... and mention how many errors were not printed.
		if len(p.errors) > 10 {
			fmt.Fprintf(os.Stderr, "... an extra %d errors were not printed.\n", len(p.errors)-10)
		}
	}

	return p.dcf
}

// returns false if upon reaching tokenEOF or tokenError
func (p parser) parseDeclaration() bool {
	t := p.lex.nextToken()
	switch t.typ {
	case tokenEOF:
		return false
	case tokenError:
		p.errors = append(p.errors, lexError(t, p.lex.lineNumber()))
		return false
	case tokenIdentifier:
		if t.val == tokenName[tokenKeyword] {
			return p.parseKeyword()
		} else if t.val == tokenName[tokenStruct] {
			return p.parseStruct()
		} else if t.val == tokenName[tokenDClass] {
			return p.parseClass()
		}
		fallthrough
	case tokenLeftCurly:
		p.errors = append(p.errors, parseError("expected a declaration but got '"+t.String()+"'",
			p.lex.lineNumber()))
		return p.expectRightCurly(p.lex.lineNumber())
	default:
		p.errors = append(p.errors, parseError("expected a declaration but got '"+t.String()+"'",
			p.lex.lineNumber()))
		return true
	}
}

// parses a keyword declaration `keyword foo;`, assumes the keyword token has already been consumed
// returns false if upon reaching tokenEOF or tokenError
func (p parser) parseKeyword() bool {
	t := p.lex.nextToken()
	switch t.typ {
	case tokenEOF:
		p.errors = append(p.errors, parseError("incomplete 'keyword' declaration, found EOF",
			p.lex.lineNumber()))
		return false
	case tokenError:
		p.errors = append(p.errors, lexError(t, p.lex.lineNumber()))
		return false
	case tokenIdentifier:
		p.dcf.AddKeyword(t.val)

		return p.expectEndline(p.lex.lineNumber())
	default:
		p.errors = append(p.errors, parseError("unexpected '"+t.String()+"' in 'keyword' declaration",
			p.lex.lineNumber()))
		return p.expectEndline(p.lex.lineNumber())
	}
}

// parses a struct declaration `struct foo {...};`, assumes the struct token has already been consumed
// returns false if upon reaching tokenEOF or tokenError
func (p parser) parseStruct() bool {
	t := p.lex.nextToken()
	switch t.typ {
	case tokenEOF:
		p.errors = append(p.errors, parseError("incomplete 'struct' declaration, found EOF",
			p.lex.lineNumber()))
		return false
	case tokenError:
		p.errors = append(p.errors, lexError(t, p.lex.lineNumber()))
		return false
	case tokenLeftCurly:
		errStr := "incomplete 'struct' declaration, missing identifier before definition start '{'"
		p.errors = append(p.errors, parseError(errStr, p.lex.lineNumber()))
		return p.expectRightCurly(p.lex.lineNumber())
	case tokenIdentifier:
		if p.dcf.Structs[t.val] != nil {
			p.errors = append(p.errors, parseError("struct "+t.val+" already defined above", p.lex.lineNumber()))
			return p.expectRightCurly(p.lex.lineNumber())
		}

		return p.parseStructInner(p.dcf.AddStruct(t.val))
	default:
		p.errors = append(p.errors, parseError("unexpected '"+t.String()+"' in 'struct' declaration",
			p.lex.lineNumber()))
		return true
	}
}

// parses the inner struct definition given within a block '{...}'
func (p parser) parseStructInner(s *Struct) bool {
	// expect a left curly to open the definition block
	t := p.lex.nextToken()
	switch t.typ {
	case tokenEOF:
		p.errors = append(p.errors, parseError("incomplete 'struct' declaration, found EOF",
			p.lex.lineNumber()))
		return false
	case tokenError:
		p.errors = append(p.errors, lexError(t, p.lex.lineNumber()))
		return false
	case tokenLeftCurly:
		break
	default:
		p.errors = append(p.errors, parseError("missing '{' after 'struct' declaration, found '"+t.String()+"'",
			p.lex.lineNumber()))
		return true
	}

	// parse for parameters till we find a RightCurly
	t = p.lex.nextToken()
	for t.typ != tokenRightCurly && t.typ != tokenEOF && t.typ != tokenError {
		if !p.parseField(s, t) {
			return false
		}
		t = p.lex.nextToken()
	}

	// finished struct definition, handle any errors then expect endline
	switch t.typ {
	case tokenEOF:
		p.errors = append(p.errors, parseError("incomplete 'struct' definition, found EOF",
			p.lex.lineNumber()))
		return false
	case tokenError:
		p.errors = append(p.errors, lexError(t, p.lex.lineNumber()))
		return false
	}

	return p.expectEndline(p.lex.lineNumber())
}

// parses a dclass declaration `dclass foo {...};`, assumes the struct token has already been consumed
// returns false if upon reaching tokenEOF or tokenError
// TODO: Implement
func (p parser) parseClass() bool {
	return true
}

// the fieldAdder interface is used by parseField() to accept any object that
// that can be composed of fields.
type fieldAdder interface {
	// AddField creates a new field and adds it to the object. The typ argument
	// can be one of "parameter", "atomic", or "molecular".  Will return nil if
	// the specified field type cannot be added to the object.
	AddField(name, typ string) Field
}

// parses a field, the first token should have been consumed and passed as an argument
// returns false upon reaching tokenEOF or tokenError
func (p parser) parseField(obj fieldAdder, first token) bool {
	switch {
	case first.typ == tokenIdentifier:
		switch p.lex.peekToken().typ {
		case tokenLeftParen:
			return p.parseAtomic(first.val, obj)
		case tokenComposition:
			return p.parseMolecular(first.val, obj)
		default:
			return p.parseParameter(first, obj, false)
		}
	case isDataTypeToken(first):
		return p.parseParameter(first, obj, false)
	default:
		p.errors = append(p.errors, parseError("expecting a field, found "+first.String(),
			p.lex.lineNumber()))
		return p.expectEndline(p.lex.lineNumber())
	}
}

// parses an atomic field `foo(...) ...;`, assumes the identifier has been consumed.
// Returns false upon reaching tokenEOF or tokenError.
// TODO: Implement
func (p parser) parseAtomic(ident string, obj fieldAdder) bool {
	return true
}

// parses a molecular field `foo: baz, bar;`, assumes the identifier has been consumed.
// Returns false upon reaching tokenEOF or tokenError.
// TODO: Implement
func (p parser) parseMolecular(ident string, obj fieldAdder) bool {
	return true
}

// parses a parameter `type foo ...;` as either a struct/class member variable
// or as an atomic field argument, assumes the type has been consumed.
// Returns false upon reaching tokenEOF or tokenError.
//
// isArgument should be true if the parameter is an argument of an atomic field.
// TODO: Implement
func (p parser) parseParameter(typTok token, obj fieldAdder, isArgument bool) bool {
	dataType := typeFromToken(typTok)
	if dataType == InvalidType {
		// TODO
	}
	return true
}

func (p parser) expectEndline(startline int) bool {
	var fail, next bool

	next = true
	t := p.lex.nextToken()
	for t.typ != tokenEndline && t.typ != tokenEOF && t.typ != tokenError {
		t = p.lex.nextToken()
		next = false
	} // consume all tokens till endline

	switch t.typ {
	case tokenEOF:
		fail = true
	case tokenError:
		p.errors = append(p.errors, lexError(t, p.lex.lineNumber()))
		fail = true
	}

	if next || fail {
		p.errors = append(p.errors, parseError("missing semicolon (;) at end of statement", startline))
	}

	return !fail
}

func (p parser) expectRightCurly(leftline int) bool {
	t := p.lex.nextToken()
	for t.typ != tokenRightCurly && t.typ != tokenEOF && t.typ != tokenError {
		t = p.lex.nextToken()
	} // consume all tokens till RightCurly

	fail := false
	switch t.typ {
	case tokenEOF:
		fail = true
	case tokenError:
		p.errors = append(p.errors, lexError(t, p.lex.lineNumber()))
		fail = true
	}

	if fail {
		errStr := fmt.Sprintf("missing closing curly brace (}) at end of block starting on line %d", leftline)
		p.errors = append(p.errors, parseError(errStr, p.lex.lineNumber()))
	}

	return !fail
}

func (p parser) next() token {
	// The dcparser is not performance critical, so we can spend some extra time while parsing
	// each token to make sure we're not trying to read past an EOF.
	if p.foundEOF {
		panic(runtimeError("eof not handled by parser, this is a bug in dcparser"))
	} else {
		t := p.lex.nextToken()
		if t.typ == tokenEOF {
			p.foundEOF = true
		}

		return t
	}
}

func typeFromToken(t token) DataType {
	switch t.typ {
	case tokenInt8:
		return Int8Type
	case tokenInt16:
		return Int16Type
	case tokenInt32:
		return Int32Type
	case tokenInt64:
		return Int64Type
	case tokenUint8:
		return Uint8Type
	case tokenUint16:
		return Uint16Type
	case tokenUint32:
		return Uint32Type
	case tokenUint64:
		return Uint64Type
	case tokenFloat:
		return FloatType
	case tokenString:
		return StringType
	case tokenBlob:
		return BlobType
	case tokenChar:
		return CharType
	case tokenIdentifier:
		return StructType
	default:
		return InvalidType
	}
}

func lexError(t token, line int) Error {
	return Error(fmt.Sprintf("lex error(line: %d): %s", line, t.String()))
}

func parseError(msg string, line int) Error {
	return Error(fmt.Sprintf("parse error(line: %d): %s", line, msg))
}

func definitionError(identifier string, typ tokenType, firstUsed int) Error {
	errStr := fmt.Sprintf("used %s '%s', but '%s' was never defined", tokenName[typ], identifier, identifier)
	errStr += fmt.Sprintf("\n\t first used on line: %d", firstUsed)
	return Error("definition error: " + errStr)
}
