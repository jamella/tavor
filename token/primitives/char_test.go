package primitives

import (
	"testing"

	. "github.com/zimmski/tavor/test/assert"

	"github.com/zimmski/tavor/test"
	"github.com/zimmski/tavor/token"
)

func TestCharTokensToBeTokens(t *testing.T) {
	var tok *token.Token

	Implements(t, tok, &CharacterClass{})
}

func TestCharacterClass(t *testing.T) {
	o := NewCharacterClass("abc")
	Equal(t, "a", o.String())

	r := test.NewRandTest(0)
	o.FuzzAll(r)
	Equal(t, "a", o.String())

	o.FuzzAll(r)
	Equal(t, "b", o.String())

	o.FuzzAll(r)
	Equal(t, "c", o.String())

	o.FuzzAll(r)
	Equal(t, "a", o.String())

	o2 := o.Clone()
	Equal(t, o.String(), o2.String())

	Equal(t, 3, o.Permutations())

	Nil(t, o.Permutation(1))
	Equal(t, "a", o.String())

	Nil(t, o.Permutation(2))
	Equal(t, "b", o.String())

	Nil(t, o.Permutation(3))
	Equal(t, "c", o.String())

	Equal(t, o.Permutation(4).(*token.PermutationError).Type, token.PermutationErrorIndexOutOfBound)

	o = NewCharacterClass(`\d`)
	Equal(t, "0", o.String())
	Equal(t, 10, o.Permutations())

	o = NewCharacterClass(`1-9`)
	Equal(t, "1", o.String())
	Equal(t, 9, o.Permutations())

	Nil(t, o.Permutation(2))
	Equal(t, "2", o.String())

	Nil(t, o.Permutation(9))
	Equal(t, "9", o.String())

	Equal(t, o.Permutation(10).(*token.PermutationError).Type, token.PermutationErrorIndexOutOfBound)

	o = NewCharacterClass(`a1-9`)
	Equal(t, "a", o.String())
	Equal(t, 10, o.Permutations())

	Nil(t, o.Permutation(2))
	Equal(t, "1", o.String())

	Nil(t, o.Permutation(10))
	Equal(t, "9", o.String())

	Equal(t, o.Permutation(11).(*token.PermutationError).Type, token.PermutationErrorIndexOutOfBound)
}
