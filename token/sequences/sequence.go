package sequences

import (
	"fmt"
	"strconv"

	"github.com/zimmski/tavor/rand"
	"github.com/zimmski/tavor/token"
	"github.com/zimmski/tavor/token/lists"
)

// Sequence implements a general sequence token which can generate Item tokens to use the internal sequence
// The sequence starts its numeration at the given start value and increases with every new sequence numeration its current value by the given step value.
type Sequence struct {
	start int
	step  int
	value int
}

// NewSequence returns a new instance of a Sequence token with a start value and a step value
func NewSequence(start, step int) *Sequence {
	return &Sequence{
		start: start,
		step:  step,
		value: start,
	}
}

func (s *Sequence) existing(r rand.Rand, except []token.Token) int {
	n := s.value - s.start

	if n == 0 {
		panic(fmt.Sprintf("There is no sequence value to choose from")) // TODO
	}

	n /= s.step

	if len(except) == 0 {
		return r.Intn(n)*s.step + s.start
	}

	checked := make(map[int]struct{})
	exceptLookup := make(map[int]struct{})

	for i := 0; i < len(except); i++ {
		ex, err := strconv.Atoi(except[i].String())
		if err != nil {
			panic(err) // TODO
		}

		exceptLookup[ex] = struct{}{}
	}

	for n != len(checked) {
		i := r.Intn(n)*s.step + s.start

		if _, ok := checked[i]; ok {
			continue
		}

		checked[i] = struct{}{}

		if _, ok := exceptLookup[i]; !ok {
			return i
		}
	}

	panic(fmt.Sprintf("There is no sequence value to choose from")) // TODO
}

// ExistingItem returns a new instance of a SequenceExistingItem token referencing the sequence and holding the starting value of the sequence as its current value
func (s *Sequence) ExistingItem(except []token.Token) *SequenceExistingItem {
	v := -1 // TODO there should be some kind of real nil value

	if s.value != s.start {
		v = s.start
	}

	return &SequenceExistingItem{
		sequence: s,
		value:    v,
		except:   except,
	}
}

// Item returns a new instance of a SequenceItem token referencing the sequence and generating and holding a new sequence numeration
func (s *Sequence) Item() *SequenceItem {
	return &SequenceItem{
		sequence: s,
		value:    s.Next(),
	}
}

// Next generates a new sequence numeration
func (s *Sequence) Next() int {
	c := s.value

	s.value += s.step

	return c
}

// ResetToken interface methods

// Reset resets the (internal) state of this token and its dependences
func (s *Sequence) Reset() {
	s.value = s.start
}

// ResetItem returns a new intsance of a SequenceResetItem token referencing the sequence
func (s *Sequence) ResetItem() *SequenceResetItem {
	return &SequenceResetItem{
		sequence: s,
	}
}

// Sequence is an unusable token

// Clone returns a copy of the token and all its children
func (s *Sequence) Clone() token.Token { panic("unusable token") }

// Fuzz fuzzes this token using the random generator by choosing one of the possible permutations for this token
func (s *Sequence) Fuzz(r rand.Rand) { panic("unusable token") }

// FuzzAll calls Fuzz for this token and then FuzzAll for all children of this token
func (s *Sequence) FuzzAll(r rand.Rand) { panic("unusable token") }

// Parse tries to parse the token beginning from the current position in the parser data.
// If the parsing is successful the error argument is nil and the next current position after the token is returned.
func (s *Sequence) Parse(pars *token.InternalParser, cur int) (int, []error) {
	panic("unusable token")
}

// Permutation sets a specific permutation for this token
func (s *Sequence) Permutation(i uint) error { panic("unusable token") }

// Permutations returns the number of permutations for this token
func (s *Sequence) Permutations() uint { panic("unusable token") }

// PermutationsAll returns the number of all possible permutations for this token including its children
func (s *Sequence) PermutationsAll() uint { panic("unusable token") }

func (s *Sequence) String() string { panic("unusable token") }

// SequenceItem implements a sequence item token which holds one distinct value of the sequence
// A new sequence value is generated on every token permutation.
type SequenceItem struct {
	sequence *Sequence
	value    int
}

// Clone returns a copy of the token and all its children
func (s *SequenceItem) Clone() token.Token {
	return &SequenceItem{
		sequence: s.sequence,
		value:    s.value,
	}
}

// Fuzz fuzzes this token using the random generator by choosing one of the possible permutations for this token
func (s *SequenceItem) Fuzz(r rand.Rand) {
	s.permutation(0)
}

// FuzzAll calls Fuzz for this token and then FuzzAll for all children of this token
func (s *SequenceItem) FuzzAll(r rand.Rand) {
	s.Fuzz(r)
}

// Parse tries to parse the token beginning from the current position in the parser data.
// If the parsing is successful the error argument is nil and the next current position after the token is returned.
func (s *SequenceItem) Parse(pars *token.InternalParser, cur int) (int, []error) {
	panic("TODO implement")
}

func (s *SequenceItem) permutation(i uint) {
	s.value = s.sequence.Next()
}

// Permutation sets a specific permutation for this token
func (s *SequenceItem) Permutation(i uint) error {
	permutations := s.Permutations()

	if i < 1 || i > permutations {
		return &token.PermutationError{
			Type: token.PermutationErrorIndexOutOfBound,
		}
	}

	s.permutation(i - 1)

	return nil
}

// Permutations returns the number of permutations for this token
func (s *SequenceItem) Permutations() uint {
	return 1
}

// PermutationsAll returns the number of all possible permutations for this token including its children
func (s *SequenceItem) PermutationsAll() uint {
	return s.Permutations()
}

func (s *SequenceItem) String() string {
	return strconv.Itoa(s.value)
}

// ResetToken interface methods

// Reset resets the (internal) state of this token and its dependences
func (s *SequenceItem) Reset() {
	s.permutation(0)
}

// SequenceExistingItem implements a sequence item token which holds one existing value of the sequence
// A new existing sequence value is choosen on every token permutation.
type SequenceExistingItem struct {
	sequence *Sequence
	value    int
	except   []token.Token
}

// Clone returns a copy of the token and all its children
func (s *SequenceExistingItem) Clone() token.Token {
	c := SequenceExistingItem{
		sequence: s.sequence,
		value:    s.value,
		except:   make([]token.Token, len(s.except)),
	}

	for i, tok := range s.except {
		c.except[i] = tok.Clone()
	}

	return &c
}

// Fuzz fuzzes this token using the random generator by choosing one of the possible permutations for this token
func (s *SequenceExistingItem) Fuzz(r rand.Rand) {
	s.permutation(r)
}

// FuzzAll calls Fuzz for this token and then FuzzAll for all children of this token
func (s *SequenceExistingItem) FuzzAll(r rand.Rand) {
	s.Fuzz(r)
}

// Parse tries to parse the token beginning from the current position in the parser data.
// If the parsing is successful the error argument is nil and the next current position after the token is returned.
func (s *SequenceExistingItem) Parse(pars *token.InternalParser, cur int) (int, []error) {
	panic("TODO implement")
}

func (s *SequenceExistingItem) permutation(r rand.Rand) {
	s.value = s.sequence.existing(r, s.except)
}

// Permutation sets a specific permutation for this token
func (s *SequenceExistingItem) Permutation(i uint) error {
	permutations := s.Permutations()

	if i < 1 || i > permutations {
		return &token.PermutationError{
			Type: token.PermutationErrorIndexOutOfBound,
		}
	}

	s.permutation(rand.NewIncrementRand(0))

	return nil
}

// Permutations returns the number of permutations for this token
func (s *SequenceExistingItem) Permutations() uint {
	return 1
}

// PermutationsAll returns the number of all possible permutations for this token including its children
func (s *SequenceExistingItem) PermutationsAll() uint {
	return s.Permutations()
}

func (s *SequenceExistingItem) String() string {
	return strconv.Itoa(s.value)
}

// ForwardToken interface methods

// Get returns the current referenced token at the given index. The error return argument is not nil, if the index is out of bound.
func (s *SequenceExistingItem) Get(i int) (token.Token, error) {
	return nil, &lists.ListError{
		Type: lists.ListErrorOutOfBound,
	}
}

// Len returns the number of the current referenced tokens
func (s *SequenceExistingItem) Len() int {
	return 0
}

// InternalGet returns the current referenced internal token at the given index. The error return argument is not nil, if the index is out of bound.
func (s *SequenceExistingItem) InternalGet(i int) (token.Token, error) {
	if i < 0 || i >= len(s.except) {
		return nil, &lists.ListError{
			Type: lists.ListErrorOutOfBound,
		}
	}

	return s.except[i], nil
}

// InternalLen returns the number of referenced internal tokens
func (s *SequenceExistingItem) InternalLen() int {
	return len(s.except)
}

// InternalLogicalRemove removes the referenced internal token and returns the replacement for the current token or nil if the current token should be removed.
func (s *SequenceExistingItem) InternalLogicalRemove(tok token.Token) token.Token {
	for i := 0; i < len(s.except); i++ {
		if s.except[i] == tok {
			if i == len(s.except)-1 {
				s.except = s.except[:i]
			} else {
				s.except = append(s.except[:i], s.except[i+1:]...)
			}

			i--
		}
	}

	return s
}

// InternalReplace replaces an old with a new internal token if it is referenced by this token
func (s *SequenceExistingItem) InternalReplace(oldToken, newToken token.Token) {
	for i := 0; i < len(s.except); i++ {
		if s.except[i] == oldToken {
			s.except[i] = newToken
		}
	}
}

// ResetToken interface methods

// Reset resets the (internal) state of this token and its dependences
func (s *SequenceExistingItem) Reset() {
	s.permutation(rand.NewIncrementRand(0))
}

// ScopeToken interface methods

// SetScope sets the scope of the token
func (s *SequenceExistingItem) SetScope(variableScope map[string]token.Token) {
	if len(s.except) != 0 {
		for i := 0; i < len(s.except); i++ {
			if tok, ok := s.except[i].(token.ScopeToken); ok {
				tok.SetScope(variableScope)
			}
		}
	}
}

// SequenceResetItem implements a sequence token item which resets its referencing sequence on every permutation
type SequenceResetItem struct {
	sequence *Sequence
}

// Clone returns a copy of the token and all its children
func (s *SequenceResetItem) Clone() token.Token {
	return &SequenceResetItem{
		sequence: s.sequence,
	}
}

// Fuzz fuzzes this token using the random generator by choosing one of the possible permutations for this token
func (s *SequenceResetItem) Fuzz(r rand.Rand) {
	s.permutation(0)
}

// FuzzAll calls Fuzz for this token and then FuzzAll for all children of this token
func (s *SequenceResetItem) FuzzAll(r rand.Rand) {
	s.Fuzz(r)
}

// Parse tries to parse the token beginning from the current position in the parser data.
// If the parsing is successful the error argument is nil and the next current position after the token is returned.
func (s *SequenceResetItem) Parse(pars *token.InternalParser, cur int) (int, []error) {
	panic("TODO implement")
}

func (s *SequenceResetItem) permutation(i uint) {
	s.sequence.Reset()
}

// Permutation sets a specific permutation for this token
func (s *SequenceResetItem) Permutation(i uint) error {
	permutations := s.Permutations()

	if i < 1 || i > permutations {
		return &token.PermutationError{
			Type: token.PermutationErrorIndexOutOfBound,
		}
	}

	s.permutation(i - 1)

	return nil
}

// Permutations returns the number of permutations for this token
func (s *SequenceResetItem) Permutations() uint {
	return 1
}

// PermutationsAll returns the number of all possible permutations for this token including its children
func (s *SequenceResetItem) PermutationsAll() uint {
	return s.Permutations()
}

func (s *SequenceResetItem) String() string {
	return ""
}

// ResetToken interface methods

// Reset resets the (internal) state of this token and its dependences
func (s *SequenceResetItem) Reset() {
	s.permutation(0)
}
