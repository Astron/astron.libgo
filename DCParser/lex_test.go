package DCParser

import (
	"fmt"
	"testing"
)

// Make the types prettyprint.
var itemName = map[itemType]string{
	itemError: "error",
	itemEOF:   "EOF",

	itemBool:    "bool",
	itemRawchar: "char-constant",
	itemNumber:  "number",
	itemQuote:   "quoted-string",

	itemIdentifier:  "identifier",
	itemOperator:    "<op>",
	itemLeftParen:   "(",
	itemRightParen:  ")",
	itemLeftCurly:   "{",
	itemRightCurly:  "}",
	itemComposition: ":",
	itemEndline:     ";",
	itemSeperator:   ",",
	itemAssignment:  "=",
	itemArray:       "[]",

	itemKeyword: "keyword",
	itemDClass:  "dclass",
	itemStruct:  "struct",

	itemInt8:   "int8",
	itemInt16:  "int16",
	itemInt32:  "int32",
	itemInt64:  "int64",
	itemUInt8:  "uint8",
	itemUInt16: "uint16",
	itemUInt32: "uint32",
	itemUInt64: "uint65",
	itemFloat:  "float64",
	itemString: "string",
	itemBlob:   "blob",
	itemChar:   "char",
}

func (i itemType) String() string {
	s := itemName[i]
	if s == "" {
		return fmt.Sprintf("item%d", int(i))
	}
	return s
}

type lexTest struct {
	name  string
	input string
	items []item
}

var (
	tEOF   = item{itemEOF, 0, ""}
	tLeft  = item{itemLeftParen, 0, "("}
	tRight = item{itemRightParen, 0, ")"}
	tOpen  = item{itemLeftCurly, 0, "{"}
	tClose = item{itemRightCurly, 0, "}"}

	caseDClass = `dclass DistributedShiny {
	                   int64 shininess required db;
	                   showListeners(int8 visible = true) broadcast;
	              }`
)

// The position is ignored on all these, so just use 0
var lexTests = []lexTest{
	{"empty", "", []item{tEOF}},
	{"spaces", " \t\n", []item{tEOF}},
	{"string", `"abc \n\t\" "`, []item{
		{itemQuote, 0, `"abc \n\t\" "`},
		tEOF,
	}},
	{"comment (w/ strs)", `"Fig"/* this is a comment */"Newtons"`, []item{
		{itemQuote, 0, `"Fig"`},
		{itemQuote, 0, `"Newtons"`},
		tEOF,
	}},
	{"spacing (w/ strs)", `"Dinosaur" "Poppies"`, []item{
		{itemQuote, 0, `"Dinosaur"`},
		{itemQuote, 0, `"Poppies"`},
		tEOF,
	}},
	{"simple number", "3", []item{{itemNumber, 0, "3"}, tEOF}},
	{"float", "3.1", []item{{itemNumber, 0, "3.1"}, tEOF}},
	{"hex", "0x5", []item{{itemNumber, 0, "0x5"}, tEOF}},
	{"oct", "0107", []item{{itemNumber, 0, "0107"}, tEOF}},
	{"bin", "0b110", []item{{itemNumber, 0, "0b110"}, tEOF}},
	{"characters", `'a' '\n' '\'' '\\' '\u00FF' '\xFF' '本'`, []item{
		{itemRawchar, 0, `'a'`},
		{itemRawchar, 0, `'\n'`},
		{itemRawchar, 0, `'\''`},
		{itemRawchar, 0, `'\\'`},
		{itemRawchar, 0, `'\u00FF'`},
		{itemRawchar, 0, `'\xFF'`},
		{itemRawchar, 0, `'本'`},
		tEOF,
	}},
	{"bools", "true false", []item{
		{itemBool, 0, "true"},
		{itemBool, 0, "false"},
		tEOF,
	}},
	{"declarations", "dclass struct keyword", []item{
		{itemDClass, 0, "dclass"},
		{itemStruct, 0, "struct"},
		{itemKeyword, 0, "keyword"},
		tEOF,
	}},
	{"variable types", "int8 uint32 uint8 int16 float64 blob string", []item{
		{itemInt8, 0, "int8"},
		{itemUInt32, 0, "uint32"},
		{itemUInt8, 0, "uint8"},
		{itemInt16, 0, "int16"},
		{itemFloat, 0, "float64"},
		{itemBlob, 0, "blob"},
		{itemString, 0, "string"},
		tEOF,
	}},
	{"operators", `+ - = / * % ; :`, []item{
		{itemOperator, 0, "+"},
		{itemOperator, 0, "-"},
		{itemAssignment, 0, "="},
		{itemOperator, 0, "/"},
		{itemOperator, 0, "*"},
		{itemOperator, 0, `%`},
		{itemEndline, 0, ";"},
		{itemComposition, 0, ":"},
		tEOF,
	}},
	{"parens", "((3))", []item{
		tLeft, tLeft,
		{itemNumber, 0, "3"},
		tRight, tRight,
		tEOF,
	}},
	{"empty block", "{}", []item{tOpen, tClose, tEOF}},
	{"simple paramater", "uint16 mask;", []item{
		{itemUInt16, 0, "uint16"},
		{itemIdentifier, 0, "mask"},
		{itemEndline, 0, `;`},
		tEOF,
	}},
	{"simple atomic", "interact(uint32) broadcast;", []item{
		{itemIdentifier, 0, "interact"},
		tLeft, {itemUInt32, 0, "uint32"}, tRight,
		{itemIdentifier, 0, "broadcast"},
		{itemEndline, 0, `;`},
		tEOF,
	}},
	{"dclass declaration", caseDClass, []item{
		{itemDClass, 0, "dclass"},
		{itemIdentifier, 0, "DistributedShiny"},
		tOpen,
		{itemInt64, 0, "int64"},
		{itemIdentifier, 0, "shininess"},
		{itemIdentifier, 0, "required"},
		{itemIdentifier, 0, "db"},
		{itemEndline, 0, ";"},
		{itemIdentifier, 0, "showListeners"},
		tLeft,
		{itemInt8, 0, "int8"},
		{itemIdentifier, 0, "visible"},
		{itemAssignment, 0, "="},
		{itemBool, 0, "true"},
		tRight,
		{itemIdentifier, 0, "broadcast"},
		{itemEndline, 0, `;`},
		tClose,
		tEOF,
	}},
	{"parenthesized arithmatic", "(10 + 3) * 4", []item{
		tLeft,
		{itemNumber, 0, "10"},
		{itemOperator, 0, "+"},
		{itemNumber, 0, "3"},
		tRight,
		{itemOperator, 0, "*"},
		{itemNumber, 0, "4"},
		tEOF,
	}},
	// errors
	{"unclosed string", "\"\n\"", []item{
		{itemError, 0, "Unterminated string"},
	}},
	{"bad number", "3k", []item{
		{itemError, 0, `Bad number syntax: "3k"`},
	}},
	{"unclosed paren", "(3", []item{
		tLeft,
		{itemNumber, 0, "3"},
		{itemError, 0, "Unclosed left paren"},
	}},
	{"extra right paren", "3)", []item{
		{itemNumber, 0, "3"},
		tRight,
		{itemError, 0, `Unexpected right paren U+0029 ')'`},
	}},
}

// collect gathers the emitted items into a slice.
func collect(t *lexTest) (items []item) {
	l := lex(t.name, t.input)
	for {
		item := l.nextItem()
		items = append(items, item)
		if item.typ == itemEOF || item.typ == itemError {
			break
		}
	}
	return
}

func equal(i1, i2 []item, checkPos bool) bool {
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
		items := collect(&test)
		if !equal(items, test.items, false) {
			t.Errorf("%s: got\n\t%+v\nexpected\n\t%v", test.name, items, test.items)
		}
	}
}

var lexPosTests = []lexTest{
	{"empty", "", []item{tEOF}},
	{"sampler", `"0123" 3  Kalimarr;`, []item{
		{itemQuote, 0, `"0123"`},
		{itemNumber, len(`"0123" `), "3"},
		{itemIdentifier, len(`"0123" 3  `), "Kalimarr"},
		{itemEndline, len(`"0123" 3  Kalimarr`), ";"},
		{itemEOF, len(`"0123" 3  Kalimarr;`), ""},
	}},
}

// TestPos test the position functionality. (TestLex does not, to simplify adding new cases)
func TestPos(t *testing.T) {
	for _, test := range lexPosTests {
		items := collect(&test)
		if !equal(items, test.items, true) {
			t.Errorf("%s: got\n\t%v\nexpected\n\t%v", test.name, items, test.items)
			if len(items) == len(test.items) {
				// Detailed print; avoid item.String() to expose the position value.
				for i := range items {
					if !equal(items[i:i+1], test.items[i:i+1], true) {
						i1 := items[i]
						i2 := test.items[i]
						t.Errorf("\t#%d: got {%v %d %q} expected  {%v %d %q}", i, i1.typ, i1.pos, i1.val, i2.typ, i2.pos, i2.val)
					}
				}
			}
		}
	}
}
