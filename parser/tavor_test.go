package parser

import (
	"strings"
	"testing"

	. "github.com/stretchr/testify/assert"

	"github.com/zimmski/tavor/test"
	"github.com/zimmski/tavor/token"
	"github.com/zimmski/tavor/token/aggregates"
	"github.com/zimmski/tavor/token/constraints"
	"github.com/zimmski/tavor/token/lists"
	"github.com/zimmski/tavor/token/primitives"
)

func TestTavorParseErrors(t *testing.T) {
	var tok token.Token
	var err error

	// empty file
	tok, err = ParseTavor(strings.NewReader(""))
	Equal(t, ParseErrorNoStart, err.(*ParserError).Type)
	Nil(t, tok)

	// empty file
	tok, err = ParseTavor(strings.NewReader("START = 123"))
	Equal(t, ParseErrorNewLineNeeded, err.(*ParserError).Type)
	Nil(t, tok)

	// new line before =
	tok, err = ParseTavor(strings.NewReader("START \n= 123\n"))
	Equal(t, ParseErrorEarlyNewLine, err.(*ParserError).Type)
	Nil(t, tok)

	// expect =
	tok, err = ParseTavor(strings.NewReader("START 123\n"))
	Equal(t, ParseErrorExpectRune, err.(*ParserError).Type)
	Nil(t, tok)

	// new line after =
	tok, err = ParseTavor(strings.NewReader("START = \n123\n"))
	Equal(t, ParseErrorEmptyTokenDefinition, err.(*ParserError).Type)
	Nil(t, tok)

	// invalid token name. does not start with letter
	tok, err = ParseTavor(strings.NewReader("3TART = 123\n"))
	Equal(t, ParseErrorInvalidTokenName, err.(*ParserError).Type)
	Nil(t, tok)

	// unused token
	tok, err = ParseTavor(strings.NewReader("START = 123\nNumber = 123\n"))
	Equal(t, ParseErrorUnusedToken, err.(*ParserError).Type)
	Nil(t, tok)

	// non-terminated string
	tok, err = ParseTavor(strings.NewReader("START = \"non-terminated string\n"))
	Equal(t, ParseErrorNonTerminatedString, err.(*ParserError).Type)
	Nil(t, tok)

	// token already exists
	tok, err = ParseTavor(strings.NewReader("START = 123\nSTART = 456\n"))
	Equal(t, ParseErrorTokenAlreadyDefined, err.(*ParserError).Type)
	Nil(t, tok)

	// token does not exists
	tok, err = ParseTavor(strings.NewReader("START = Token\n"))
	Equal(t, ParseErrorTokenNotDefined, err.(*ParserError).Type)
	Nil(t, tok)

	// unexpected multi line token termination
	tok, err = ParseTavor(strings.NewReader("Hello = 1,\n\n"))
	Equal(t, ParseErrorUnexpectedTokenDefinitionTermination, err.(*ParserError).Type)
	Nil(t, tok)

	// unexpected continue of multi line token
	tok, err = ParseTavor(strings.NewReader("Hello = 1,2\n"))
	Equal(t, ParseErrorExpectRune, err.(*ParserError).Type)
	Nil(t, tok)

	// unknown token attribute
	tok, err = ParseTavor(strings.NewReader("Token = 123\nSTART = $Token.yeah\n"))
	Equal(t, ParseErrorUnknownTokenAttribute, err.(*ParserError).Type)
	Nil(t, tok)

	// unknown token attribute
	tok, err = ParseTavor(strings.NewReader("Token = 123\nSTART = Token $Token.Count\n"))
	Equal(t, ParseErrorUnknownTokenAttribute, err.(*ParserError).Type)
	Nil(t, tok)
}

func TestTavorParserSimple(t *testing.T) {
	var tok token.Token
	var err error

	// constant integer
	tok, err = ParseTavor(strings.NewReader("START = 123\n"))
	Nil(t, err)
	Equal(t, tok, primitives.NewConstantInt(123))

	// single line comment
	tok, err = ParseTavor(strings.NewReader("// hello\nSTART = 123\n"))
	Nil(t, err)
	Equal(t, tok, primitives.NewConstantInt(123))

	// single line multi line comment
	tok, err = ParseTavor(strings.NewReader("/* hello */\nSTART = 123\n"))
	Nil(t, err)
	Equal(t, tok, primitives.NewConstantInt(123))

	// multi line multi line comment
	tok, err = ParseTavor(strings.NewReader("/*\nh\ne\nl\nl\no\n*/\nSTART = 123\n"))
	Nil(t, err)
	Equal(t, tok, primitives.NewConstantInt(123))

	// inline comment
	tok, err = ParseTavor(strings.NewReader("START /* ok */= /* or so */ 123\n"))
	Nil(t, err)
	Equal(t, tok, primitives.NewConstantInt(123))

	// constant string
	tok, err = ParseTavor(strings.NewReader("START = \"abc\"\n"))
	Nil(t, err)
	Equal(t, tok, primitives.NewConstantString("abc"))

	// constant string with whitespaces and epic chars
	tok, err = ParseTavor(strings.NewReader("START = \"a b c !\\\"$%&/\"\n"))
	Nil(t, err)
	Equal(t, tok, primitives.NewConstantString("a b c !\\\"$%&/"))

	// concatination
	tok, err = ParseTavor(strings.NewReader("START = \"I am a constant string\" 123\n"))
	Nil(t, err)
	Equal(t, tok, lists.NewAll(
		primitives.NewConstantString("I am a constant string"),
		primitives.NewConstantInt(123),
	))

	// embed token
	tok, err = ParseTavor(strings.NewReader("Token=123\nSTART = Token\n"))
	Nil(t, err)
	Equal(t, tok, primitives.NewConstantInt(123))

	// embed over token
	tok, err = ParseTavor(strings.NewReader("Token=123\nAnotherToken = Token\nSTART = AnotherToken\n"))
	Nil(t, err)
	Equal(t, tok, primitives.NewConstantInt(123))

	// multi line token
	tok, err = ParseTavor(strings.NewReader("START = 1,\n2,\n3\n"))
	Nil(t, err)
	Equal(t, tok, lists.NewAll(
		primitives.NewConstantInt(1),
		primitives.NewConstantInt(2),
		primitives.NewConstantInt(3),
	))

	// Umläüt
	tok, err = ParseTavor(strings.NewReader("Umläüt=123\nSTART = Umläüt\n"))
	Nil(t, err)
	Equal(t, tok, primitives.NewConstantInt(123))
}

func TestTavorParserAlternationsAndGroupings(t *testing.T) {
	var tok token.Token
	var err error

	// simple alternation
	tok, err = ParseTavor(strings.NewReader("START = 1 | 2 | 3\n"))
	Nil(t, err)
	Equal(t, tok, lists.NewOne(
		primitives.NewConstantInt(1),
		primitives.NewConstantInt(2),
		primitives.NewConstantInt(3),
	))

	// concatinated alternation
	tok, err = ParseTavor(strings.NewReader("START = 1 | 2 3 | 4\n"))
	Nil(t, err)
	Equal(t, tok, lists.NewOne(
		primitives.NewConstantInt(1),
		lists.NewAll(
			primitives.NewConstantInt(2),
			primitives.NewConstantInt(3),
		),
		primitives.NewConstantInt(4),
	))

	// optional alternation
	tok, err = ParseTavor(strings.NewReader("START = | 2 | 3\n"))
	Nil(t, err)
	Equal(t, tok, constraints.NewOptional(lists.NewOne(
		primitives.NewConstantInt(2),
		primitives.NewConstantInt(3),
	)))

	tok, err = ParseTavor(strings.NewReader("START = 1 | | 3\n"))
	Nil(t, err)
	Equal(t, tok, constraints.NewOptional(lists.NewOne(
		primitives.NewConstantInt(1),
		primitives.NewConstantInt(3),
	)))

	tok, err = ParseTavor(strings.NewReader("START = 1 | 2 |\n"))
	Nil(t, err)
	Equal(t, tok, constraints.NewOptional(lists.NewOne(
		primitives.NewConstantInt(1),
		primitives.NewConstantInt(2),
	)))

	// alternation with embedded token
	tok, err = ParseTavor(strings.NewReader("Token = 2\nSTART = 1 | Token\n"))
	Nil(t, err)
	Equal(t, tok, lists.NewOne(
		primitives.NewConstantInt(1),
		primitives.NewConstantInt(2),
	))

	// simple group
	tok, err = ParseTavor(strings.NewReader("START = (1 2 3)\n"))
	Nil(t, err)
	Equal(t, tok, lists.NewAll(
		primitives.NewConstantInt(1),
		primitives.NewConstantInt(2),
		primitives.NewConstantInt(3),
	))

	// simple embedded group
	tok, err = ParseTavor(strings.NewReader("START = 0 (1 2 3) 4\n"))
	Nil(t, err)
	Equal(t, tok, lists.NewAll(
		primitives.NewConstantInt(0),
		lists.NewAll(
			primitives.NewConstantInt(1),
			primitives.NewConstantInt(2),
			primitives.NewConstantInt(3),
		),
		primitives.NewConstantInt(4),
	))

	// simple embedded or group
	tok, err = ParseTavor(strings.NewReader("START = 0 (1 | 2 | 3) 4\n"))
	Nil(t, err)
	Equal(t, tok, lists.NewAll(
		primitives.NewConstantInt(0),
		lists.NewOne(
			primitives.NewConstantInt(1),
			primitives.NewConstantInt(2),
			primitives.NewConstantInt(3),
		),
		primitives.NewConstantInt(4),
	))

	// Yo dog, I heard you like groups? so here is a group in a group
	tok, err = ParseTavor(strings.NewReader("START = (1 | (2 | 3)) | 4\n"))
	Nil(t, err)
	Equal(t, tok, lists.NewOne(
		lists.NewOne(
			primitives.NewConstantInt(1),
			lists.NewOne(
				primitives.NewConstantInt(2),
				primitives.NewConstantInt(3),
			),
		),
		primitives.NewConstantInt(4),
	))

	// simple optional
	tok, err = ParseTavor(strings.NewReader("START = 1 ?(2)\n"))
	Nil(t, err)
	Equal(t, tok, lists.NewAll(
		primitives.NewConstantInt(1),
		constraints.NewOptional(primitives.NewConstantInt(2)),
	))

	// or optional
	tok, err = ParseTavor(strings.NewReader("START = 1 ?(2 | 3) 4\n"))
	Nil(t, err)
	Equal(t, tok, lists.NewAll(
		primitives.NewConstantInt(1),
		constraints.NewOptional(lists.NewOne(
			primitives.NewConstantInt(2),
			primitives.NewConstantInt(3),
		)),
		primitives.NewConstantInt(4),
	))

	// simple repeat
	tok, err = ParseTavor(strings.NewReader("START = 1 +(2)\n"))
	Nil(t, err)
	Equal(t, tok, lists.NewAll(
		primitives.NewConstantInt(1),
		lists.NewRepeat(primitives.NewConstantInt(2), 1, MaxRepeat),
	))

	// or repeat
	tok, err = ParseTavor(strings.NewReader("START = 1 +(2 | 3) 4\n"))
	Nil(t, err)
	Equal(t, tok, lists.NewAll(
		primitives.NewConstantInt(1),
		lists.NewRepeat(lists.NewOne(
			primitives.NewConstantInt(2),
			primitives.NewConstantInt(3),
		), 1, MaxRepeat),
		primitives.NewConstantInt(4),
	))

	// simple optional repeat
	tok, err = ParseTavor(strings.NewReader("START = 1 *(2)\n"))
	Nil(t, err)
	Equal(t, tok, lists.NewAll(
		primitives.NewConstantInt(1),
		lists.NewRepeat(primitives.NewConstantInt(2), 0, MaxRepeat),
	))

	// or optional repeat
	tok, err = ParseTavor(strings.NewReader("START = 1 *(2 | 3) 4\n"))
	Nil(t, err)
	Equal(t, tok, lists.NewAll(
		primitives.NewConstantInt(1),
		lists.NewRepeat(lists.NewOne(
			primitives.NewConstantInt(2),
			primitives.NewConstantInt(3),
		), 0, MaxRepeat),
		primitives.NewConstantInt(4),
	))

	// simple optional repeat
	tok, err = ParseTavor(strings.NewReader("START = 1 *(2)\n"))
	Nil(t, err)
	Equal(t, tok, lists.NewAll(
		primitives.NewConstantInt(1),
		lists.NewRepeat(primitives.NewConstantInt(2), 0, MaxRepeat),
	))

	// exact repeat
	tok, err = ParseTavor(strings.NewReader("START = 1 +3(2)\n"))
	Nil(t, err)
	Equal(t, tok, lists.NewAll(
		primitives.NewConstantInt(1),
		lists.NewRepeat(primitives.NewConstantInt(2), 3, 3),
	))

	// at least repeat
	tok, err = ParseTavor(strings.NewReader("START = 1 +3,(2)\n"))
	Nil(t, err)
	Equal(t, tok, lists.NewAll(
		primitives.NewConstantInt(1),
		lists.NewRepeat(primitives.NewConstantInt(2), 3, MaxRepeat),
	))

	// at most repeat
	tok, err = ParseTavor(strings.NewReader("START = 1 +,3(2)\n"))
	Nil(t, err)
	Equal(t, tok, lists.NewAll(
		primitives.NewConstantInt(1),
		lists.NewRepeat(primitives.NewConstantInt(2), 1, 3),
	))

	// range repeat
	tok, err = ParseTavor(strings.NewReader("START = 1 +2,3(2)\n"))
	Nil(t, err)
	Equal(t, tok, lists.NewAll(
		primitives.NewConstantInt(1),
		lists.NewRepeat(primitives.NewConstantInt(2), 2, 3),
	))
}

func TestTavorParserTokenAttributes(t *testing.T) {
	var tok token.Token
	var err error

	// token attribute List.Count
	tok, err = ParseTavor(strings.NewReader(
		"Digit = 1 | 2 | 3\n" +
			"Digits = *(Digit)\n" +
			"START = Digits \"->\" $Digits.Count\n",
	))
	Nil(t, err)
	{
		v, _ := tok.(*lists.All).Get(0)
		list := v.(*lists.Repeat)

		Equal(t, tok, lists.NewAll(
			lists.NewRepeat(lists.NewOne(
				primitives.NewConstantInt(1),
				primitives.NewConstantInt(2),
				primitives.NewConstantInt(3),
			), 0, MaxRepeat),
			primitives.NewConstantString("->"),
			aggregates.NewLen(list),
		))

		r := test.NewRandTest(1)
		tok.Fuzz(r)
		Equal(t, "12->2", tok.String())
	}
}
