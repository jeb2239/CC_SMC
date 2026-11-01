// Package semantic is a Go port of smc.semanticAnalyzer.
package semantic

import (
	"fmt"
	"sort"
	"strings"
)

// StateMachine is the semantic model of an FSM produced by Analyzer.
//
// Where the Java version uses null for a "-" event, this one uses the empty
// string. Events and Actions hold each name once, in the order it first
// appears in the source; the Java version keeps them in a HashSet.
type StateMachine struct {
	Errors       []AnalysisError
	Warnings     []AnalysisError
	States       map[string]*State
	Events       []string
	Actions      []string
	InitialState *State
	ActionClass  string
	FsmName      string
}

func newStateMachine() *StateMachine {
	return &StateMachine{States: map[string]*State{}}
}

func (sm *StateMachine) String() string {
	return fmt.Sprintf("Actions: %s\nFSM: %s\nInitial: %s%s",
		sm.ActionClass, sm.FsmName, sm.InitialState.Name, sm.StatesString())
}

func (sm *StateMachine) addError(err AnalysisError) {
	sm.Errors = append(sm.Errors, err)
}

// SortedStates returns the states ordered by name.
func (sm *StateMachine) SortedStates() []*State {
	states := make([]*State, 0, len(sm.States))
	for _, s := range sm.States {
		states = append(states, s)
	}
	sort.Slice(states, func(i, j int) bool { return states[i].Name < states[j].Name })
	return states
}

// StatesString formats every state, ordered by name.
func (sm *StateMachine) StatesString() string {
	var sb strings.Builder
	sb.WriteString("{")
	for _, s := range sm.SortedStates() {
		sb.WriteString(s.String())
	}
	sb.WriteString("}\n")
	return sb.String()
}

// State is a state of the semantic model. SuperStates is ordered by name and
// holds no duplicates.
type State struct {
	Name          string
	EntryActions  []string
	ExitActions   []string
	AbstractState bool
	SuperStates   []*State
	Transitions   []*Transition
}

func newState(name string) *State {
	return &State{Name: name}
}

func (s *State) addSuperState(superState *State) {
	i := sort.Search(len(s.SuperStates), func(i int) bool { return s.SuperStates[i].Name >= superState.Name })
	if i < len(s.SuperStates) && s.SuperStates[i].Name == superState.Name {
		return
	}
	s.SuperStates = append(s.SuperStates, nil)
	copy(s.SuperStates[i+1:], s.SuperStates[i:])
	s.SuperStates[i] = superState
}

func (s *State) String() string {
	return fmt.Sprintf("\n  %s {\n%s  }\n", s.nameWithAdornments(), s.transitionStrings())
}

func (s *State) transitionStrings() string {
	var sb strings.Builder
	for _, t := range s.Transitions {
		fmt.Fprintf(&sb, "    %s %s {%s}\n", nullable(t.Event), t.nextStateName(), strings.Join(t.Actions, " "))
	}
	return sb.String()
}

func (s *State) nameWithAdornments() string {
	stateName := s.Name
	if s.AbstractState {
		stateName = "(" + s.Name + ")"
	}
	for _, superState := range s.SuperStates {
		stateName += " :" + superState.Name
	}
	for _, entryAction := range s.EntryActions {
		stateName += " <" + entryAction
	}
	for _, exitAction := range s.ExitActions {
		stateName += " >" + exitAction
	}
	return stateName
}

// Transition is a transition out of a State.
type Transition struct {
	Event     string
	NextState *State
	Actions   []string
}

func (t *Transition) nextStateName() string {
	if t.NextState == nil {
		return "null"
	}
	return t.NextState.Name
}

// ErrorID identifies the kind of an AnalysisError.
type ErrorID int

const (
	NoFsm ErrorID = iota
	NoInitial
	InvalidHeader
	ExtraHeaderIgnored
	UndefinedState
	UndefinedSuperState
	UnusedState
	DuplicateTransition
	AbstractStateUsedAsNextState
	InconsistentAbstraction
	StateActionsMultiplyDefined
	ConflictingSuperstates
)

// The names match the Java enum so error messages are identical.
var errorIDNames = [...]string{
	"NO_FSM",
	"NO_INITIAL",
	"INVALID_HEADER",
	"EXTRA_HEADER_IGNORED",
	"UNDEFINED_STATE",
	"UNDEFINED_SUPER_STATE",
	"UNUSED_STATE",
	"DUPLICATE_TRANSITION",
	"ABSTRACT_STATE_USED_AS_NEXT_STATE",
	"INCONSISTENT_ABSTRACTION",
	"STATE_ACTIONS_MULTIPLY_DEFINED",
	"CONFLICTING_SUPERSTATES",
}

func (id ErrorID) String() string { return errorIDNames[id] }

// AnalysisError is a semantic error or warning. Extra is nil, a string, or a
// parser.Header; AnalysisErrors compare equal with ==.
type AnalysisError struct {
	ID    ErrorID
	Extra any
}

func (e AnalysisError) String() string {
	if e.Extra == nil {
		return fmt.Sprintf("Semantic Error: %s(null)", e.ID)
	}
	return fmt.Sprintf("Semantic Error: %s(%v)", e.ID, e.Extra)
}

func (e AnalysisError) Error() string { return e.String() }

// nullable prints an absent name the way Java prints null.
func nullable(name string) string {
	if name == "" {
		return "null"
	}
	return name
}
