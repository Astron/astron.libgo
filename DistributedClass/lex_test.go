package dclass

import (
	"testing"
)

type lexTest struct {
	name   string
	input  string
	tokens []token
}

var (
	tEOF   = token{tokenEOF, 0, ""}
	tLeft  = token{tokenLeftParen, 0, "("}
	tRight = token{tokenRightParen, 0, ")"}
	tOpen  = token{tokenLeftCurly, 0, "{"}
	tClose = token{tokenRightCurly, 0, "}"}

	caseDClass = `dclass DistributedShiny {
	                   int64 shininess required db;
	                   showListeners(int8 visible = true) broadcast;
	              }`
)

// The position is ignored on all these, so just use 0
var lexTests = []lexTest{
	{"empty", "", []token{tEOF}},
	{"spaces", " \t\n", []token{tEOF}},
	{"string", `"abc \n\t\" "`, []token{
		{tokenQuote, 0, `"abc \n\t\" "`},
		tEOF,
	}},
	{"comment (w/ strs)", `"Fig"/* this is a comment */"Newtons"`, []token{
		{tokenQuote, 0, `"Fig"`},
		{tokenQuote, 0, `"Newtons"`},
		tEOF,
	}},
	{"spacing (w/ strs)", `"Dinosaur" "Poppies"`, []token{
		{tokenQuote, 0, `"Dinosaur"`},
		{tokenQuote, 0, `"Poppies"`},
		tEOF,
	}},
	{"simple number", "3", []token{{tokenNumber, 0, "3"}, tEOF}},
	{"float", "3.1", []token{{tokenNumber, 0, "3.1"}, tEOF}},
	{"hex", "0x5", []token{{tokenNumber, 0, "0x5"}, tEOF}},
	{"oct", "0107", []token{{tokenNumber, 0, "0107"}, tEOF}},
	{"bin", "0b110", []token{{tokenNumber, 0, "0b110"}, tEOF}},
	{"characters", `'a' '\n' '\'' '\\' '\u00FF' '\xFF' '本'`, []token{
		{tokenRawchar, 0, `'a'`},
		{tokenRawchar, 0, `'\n'`},
		{tokenRawchar, 0, `'\''`},
		{tokenRawchar, 0, `'\\'`},
		{tokenRawchar, 0, `'\u00FF'`},
		{tokenRawchar, 0, `'\xFF'`},
		{tokenRawchar, 0, `'本'`},
		tEOF,
	}},
	{"bools", "true false", []token{
		{tokenBool, 0, "true"},
		{tokenBool, 0, "false"},
		tEOF,
	}},
	{"declarations", "dclass struct keyword", []token{
		{tokenDClass, 0, "dclass"},
		{tokenStruct, 0, "struct"},
		{tokenKeyword, 0, "keyword"},
		tEOF,
	}},
	{"variable types", "int8 uint32 uint8 int16 float64 blob string", []token{
		{tokenInt8, 0, "int8"},
		{tokenUint32, 0, "uint32"},
		{tokenUint8, 0, "uint8"},
		{tokenInt16, 0, "int16"},
		{tokenFloat, 0, "float64"},
		{tokenBlob, 0, "blob"},
		{tokenString, 0, "string"},
		tEOF,
	}},
	{"operators", `+ - = / * % ; :`, []token{
		{tokenOperator, 0, "+"},
		{tokenOperator, 0, "-"},
		{tokenAssignment, 0, "="},
		{tokenOperator, 0, "/"},
		{tokenOperator, 0, "*"},
		{tokenOperator, 0, `%`},
		{tokenEndline, 0, ";"},
		{tokenComposition, 0, ":"},
		tEOF,
	}},
	{"parens", "((3))", []token{
		tLeft, tLeft,
		{tokenNumber, 0, "3"},
		tRight, tRight,
		tEOF,
	}},
	{"empty block", "{}", []token{tOpen, tClose, tEOF}},
	{"simple paramater", "uint16 mask;", []token{
		{tokenUint16, 0, "uint16"},
		{tokenIdentifier, 0, "mask"},
		{tokenEndline, 0, `;`},
		tEOF,
	}},
	{"simple atomic", "interact(uint32) broadcast;", []token{
		{tokenIdentifier, 0, "interact"},
		tLeft, {tokenUint32, 0, "uint32"}, tRight,
		{tokenIdentifier, 0, "broadcast"},
		{tokenEndline, 0, `;`},
		tEOF,
	}},
	{"dclass declaration", caseDClass, []token{
		{tokenDClass, 0, "dclass"},
		{tokenIdentifier, 0, "DistributedShiny"},
		tOpen,
		{tokenInt64, 0, "int64"},
		{tokenIdentifier, 0, "shininess"},
		{tokenIdentifier, 0, "required"},
		{tokenIdentifier, 0, "db"},
		{tokenEndline, 0, ";"},
		{tokenIdentifier, 0, "showListeners"},
		tLeft,
		{tokenInt8, 0, "int8"},
		{tokenIdentifier, 0, "visible"},
		{tokenAssignment, 0, "="},
		{tokenBool, 0, "true"},
		tRight,
		{tokenIdentifier, 0, "broadcast"},
		{tokenEndline, 0, `;`},
		tClose,
		tEOF,
	}},
	{"parenthesized arithmatic", "(10 + 3) * 4", []token{
		tLeft,
		{tokenNumber, 0, "10"},
		{tokenOperator, 0, "+"},
		{tokenNumber, 0, "3"},
		tRight,
		{tokenOperator, 0, "*"},
		{tokenNumber, 0, "4"},
		tEOF,
	}},
	// errors
	{"unclosed string", "\"\n\"", []token{
		{tokenError, 0, "unterminated string"},
	}},
	{"bad number", "3k", []token{
		{tokenError, 0, `bad number syntax: "3k"`},
	}},
	{"unclosed paren", "(3", []token{
		tLeft,
		{tokenNumber, 0, "3"},
		{tokenError, 0, "unclosed left paren"},
	}},
	{"extra right paren", "3)", []token{
		{tokenNumber, 0, "3"},
		tRight,
		{tokenError, 0, `unexpected right paren U+0029 ')'`},
	}},
}

// collect gathers the emitted tokens into a slice.
func collect(t *lexTest) (tokens []token) {
	l := lex(t.input)
	for {
		token := l.nextToken()
		tokens = append(tokens, token)
		if token.typ == tokenEOF || token.typ == tokenError {
			break
		}
	}
	return
}

func equal(i1, i2 []token, checkPos bool) bool {
	if len(i1) != len(i2) {
		return false
	}
	for k := range i1 {
		if i1[k].typ != i2[k].typ {
			return false
		}
		if i1[k].val != i2[k].val {
			return false
		}
		if checkPos && i1[k].pos != i2[k].pos {
			return false
		}
	}
	return true
}

func TestLex(t *testing.T) {
	for _, test := range lexTests {
		tokens := collect(&test)
		if !equal(tokens, test.tokens, false) {
			t.Errorf("%s: got\n\t%+v\nexpected\n\t%v", test.name, tokens, test.tokens)
		}
	}
}

var lexPosTests = []lexTest{
	{"empty", "", []token{tEOF}},
	{"sampler", `"0123" 3  Kalimarr;`, []token{
		{tokenQuote, 0, `"0123"`},
		{tokenNumber, len(`"0123" `), "3"},
		{tokenIdentifier, len(`"0123" 3  `), "Kalimarr"},
		{tokenEndline, len(`"0123" 3  Kalimarr`), ";"},
		{tokenEOF, len(`"0123" 3  Kalimarr;`), ""},
	}},
}

// TestPos test the position functionality. (TestLex does not, to simplify adding new cases)
func TestPos(t *testing.T) {
	for _, test := range lexPosTests {
		tokens := collect(&test)
		if !equal(tokens, test.tokens, true) {
			t.Errorf("%s: got\n\t%v\nexpected\n\t%v", test.name, tokens, test.tokens)
			if len(tokens) == len(test.tokens) {
				// Detailed print; avoid token.String() to expose the position value.
				for i := range tokens {
					if !equal(tokens[i:i+1], test.tokens[i:i+1], true) {
						i1 := tokens[i]
						i2 := test.tokens[i]
						t.Errorf("\t#%d: got {%v %d %q} expected  {%v %d %q}", i, i1.typ, i1.pos, i1.val, i2.typ, i2.pos, i2.val)
					}
				}
			}
		}
	}
}
