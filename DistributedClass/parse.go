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
	if len(p.errors) {
		fmt.Fprintf(os.Stderr, "Encountered %d errors while parsing dclass file...\n", len(p.errors))

		// Print max 10 errors...
		for i := 0; i < 10 && i < len(p.errors); i++ {
			fmt.Fprintf(os.Stderr, " - %s\n", p.errors[i].Error())
		}

		// ... and mention how many errors were not printed.
		if(len(p.errors) > 10) {
			fmt.Fprintf(os.Stderr, "... an extra %d errors were not printed.\n", len(p.errors-10))
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
			return p.parseDClass()
		}
		fallthrough
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
		return true
	}
}

// parses a struct declaration `struct foo {...};`, assumes the struct token has already been consumed
// returns false if upon reaching tokenEOF or tokenError
func (p parser) parseStruct() bool {
	t := p.lex.nextToken()
	switch t.typ {
	case tokenEOF:
		p.errors = append
	}
}

// parses a dclass declaration `dclass foo {...};`, assumes the struct token has already been consumed
// returns false if upon reaching tokenEOF or tokenError
// TODO: Implement
func (p parser) parseDClass() bool {
	return true
}

// parses a parameter
// returns false if upon reaching tokenEOF or tokenError
// TODO: Implement
func (p parser) parseParameter() bool {
	return true
}

func (p parser) expectEndline(startline int) bool {
	fail := false
	t := p.lex.nextToken()
	for t.typ != tokenEndline && t.typ != tokenEOF && t.typ != tokenError {
		fail = true
		t = p.lex.nextToken()
	} // consume all tokens till endline

	switch t.typ {
	case tokenEOF:
		fail = true
	case tokenError:
		p.errors = append(p.errors, lexError(t, p.lex.lineNumber()))
		fail = true
	}

	if fail {
		p.errors = append(p.errors, parseError("missing semicolon (;) at end of statement", startline))
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

func lexError(t token, line int) Error {
	return Error(fmt.Sprintf("lex error(line: %d): %s", line, t.String()))
}

func parseError(msg string, line int) Error {
	return Error(fmt.Sprintf("parse error(line: %d): %s", line, msg))
}

func definitionError(identifier string, typ tokenType, firstUsed int) Error {
	errStr := fmt.Sprintf("used %s '%s', but '%s' was never defined", tokenName[typ], keyword, keyword)
	errStr += fmt.Sprintf("\n\t first used on line: %d", firstLine)
	return Error("definition error: " + errStr)
}