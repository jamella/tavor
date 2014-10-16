package strategy

import (
	"github.com/zimmski/container/list/linkedlist"

	"github.com/zimmski/tavor"
	"github.com/zimmski/tavor/log"
	"github.com/zimmski/tavor/rand"
	"github.com/zimmski/tavor/token"
	"github.com/zimmski/tavor/token/lists"
	"github.com/zimmski/tavor/token/primitives"
	"github.com/zimmski/tavor/token/sequences"
)

// RandomStrategy implements a fuzzing strategy that generates a random permutation of a token graph.
// The strategy does exactly one iteration which permutates at random all reachable tokens in the graph. The determinism is dependent on the random generator and is therefore for example deterministic if a seed for the random generator produces always the same outputs.
type RandomStrategy struct {
	root token.Token
}

// NewRandomStrategy returns a new instance of the random fuzzing strategy
func NewRandomStrategy(tok token.Token) *RandomStrategy {
	return &RandomStrategy{
		root: tok,
	}
}

func init() {
	Register("random", func(tok token.Token) Strategy {
		return NewRandomStrategy(tok)
	})
}

// Fuzz starts the first iteration of the fuzzing strategy returning a channel which controls the iteration flow.
// The channel returns a value if the iteration is complete and waits with calculating the next iteration until a value is put in. The channel is automatically closed when there are no more iterations. The error return argument is not nil if an error occurs during the setup of the fuzzing strategy.
func (s *RandomStrategy) Fuzz(r rand.Rand) (chan struct{}, error) {
	if tavor.LoopExists(s.root) {
		return nil, &Error{
			Message: "found endless loop in graph. Cannot proceed.",
			Type:    ErrorEndlessLoopDetected,
		}
	}

	continueFuzzing := make(chan struct{})

	go func() {
		log.Debug("start random fuzzing routine")

		s.fuzz(s.root, r)

		s.fuzzYADDA(s.root, r)

		log.Debug("done with fuzzing step")

		// done with the last fuzzing step
		continueFuzzing <- struct{}{}

		log.Debug("finished fuzzing. Wait till the outside is ready to close.")

		if _, ok := <-continueFuzzing; ok {
			log.Debug("close fuzzing channel")

			close(continueFuzzing)
		}
	}()

	return continueFuzzing, nil
}

func (s *RandomStrategy) fuzz(tok token.Token, r rand.Rand) {
	tok.Fuzz(r)

	switch t := tok.(type) {
	case token.ForwardToken:
		if v := t.Get(); v != nil {
			s.fuzz(v, r)
		}
	case token.List:
		l := t.Len()

		for i := 0; i < l; i++ {
			c, _ := t.Get(i)
			s.fuzz(c, r)
		}
	}
}

func (s *RandomStrategy) fuzzYADDA(root token.Token, r rand.Rand) {

	// TODO FIXME AND FIXME FIXME FIXME this should be done automatically somehow
	// since this doesn't work in other heuristics...
	// especially the fuzz again part is tricky. the whole reason is because of dynamic repeats that clone during a reset. so the "reset" or regenerating of new child tokens has to be done better

	scope := make(map[string]token.Token)
	queue := linkedlist.New()

	type set struct {
		token token.Token
		scope map[string]token.Token
	}

	queue.Push(set{
		token: root,
		scope: scope,
	})

	var fuzzAgain []token.Token

	for !queue.Empty() {
		v, _ := queue.Shift()
		s := v.(set)

		if tok, ok := s.token.(token.ResetToken); ok {
			log.Debugf("reset %#v(%p)", tok, tok)

			tok.Reset()

			fuzzAgain = append(fuzzAgain, tok)
		}

		if tok, ok := s.token.(token.ScopeToken); ok {
			log.Debugf("setScope %#v(%p)", tok, tok)

			tok.SetScope(s.scope)

			fuzzAgain = append(fuzzAgain, tok)
		}

		nScope := make(map[string]token.Token, len(s.scope))
		for k, v := range s.scope {
			nScope[k] = v
		}

		switch t := s.token.(type) {
		case token.ForwardToken:
			if v := t.Get(); v != nil {
				queue.Push(set{
					token: v,
					scope: nScope,
				})
			}
		case token.List:
			for i := 0; i < t.Len(); i++ {
				c, _ := t.Get(i)

				queue.Push(set{
					token: c,
					scope: nScope,
				})
			}
		}
	}

	alreadyFuzzed := make(map[token.Token]struct{})

	for _, tok := range fuzzAgain {
		queue.Push(tok)
	}

	for !queue.Empty() {
		v, _ := queue.Shift()
		tok := v.(token.Token)

		if _, ok := alreadyFuzzed[tok]; ok {
			continue
		}

		alreadyFuzzed[tok] = struct{}{}

		switch tok.(type) {
		case *sequences.SequenceExistingItem, *lists.UniqueItem, *primitives.CharacterClass:
			log.Debugf("Fuzz again %p(%#v)", tok, tok)

			tok.Fuzz(r)
		}

		switch t := tok.(type) {
		case token.ForwardToken:
			if v := t.Get(); v != nil {
				queue.Push(v)
			}
		case token.List:
			for i := 0; i < t.Len(); i++ {
				c, _ := t.Get(i)

				queue.Push(c)
			}
		}
	}
}
