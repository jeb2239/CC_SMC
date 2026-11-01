package semantic

import (
	"fmt"
	"slices"
	"strings"

	"github.com/jeb2239/CC_SMC/go/parser"
)

// Analyzer checks an FsmSyntax for semantic errors and, if there are none,
// builds the corresponding StateMachine.
type Analyzer struct {
	sm            *StateMachine
	fsmHeader     parser.Header
	actionsHeader parser.Header
	initialHeader parser.Header
}

// NewAnalyzer returns an Analyzer.
func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

// Analyze checks fsm and returns its semantic model. If the model has
// Errors, only the state, event and action lists are filled in.
func (a *Analyzer) Analyze(fsm *parser.FsmSyntax) *StateMachine {
	a.sm = newStateMachine()
	a.fsmHeader, a.actionsHeader, a.initialHeader = parser.Header{}, parser.Header{}, parser.Header{}
	a.analyzeHeaders(fsm)
	a.checkSemanticValidity(fsm)
	a.produceSemanticStateMachine(fsm)
	return a.sm
}

func (a *Analyzer) analyzeHeaders(fsm *parser.FsmSyntax) {
	a.setHeaders(fsm)
	a.checkMissingHeaders()
}

func (a *Analyzer) setHeaders(fsm *parser.FsmSyntax) {
	for _, header := range fsm.Headers {
		switch {
		case isNamed(header, "fsm"):
			a.setHeader(&a.fsmHeader, header)
		case isNamed(header, "actions"):
			a.setHeader(&a.actionsHeader, header)
		case isNamed(header, "initial"):
			a.setHeader(&a.initialHeader, header)
		default:
			a.sm.addError(AnalysisError{InvalidHeader, header})
		}
	}
}

func isNamed(header parser.Header, headerName string) bool {
	return strings.EqualFold(header.Name, headerName)
}

func (a *Analyzer) setHeader(targetHeader *parser.Header, header parser.Header) {
	if isNullHeader(*targetHeader) {
		*targetHeader = header
	} else {
		a.sm.addError(AnalysisError{ExtraHeaderIgnored, header})
	}
}

func (a *Analyzer) checkMissingHeaders() {
	if isNullHeader(a.fsmHeader) {
		a.sm.addError(AnalysisError{ID: NoFsm})
	}
	if isNullHeader(a.initialHeader) {
		a.sm.addError(AnalysisError{ID: NoInitial})
	}
}

func isNullHeader(header parser.Header) bool {
	return header.Name == ""
}

func (a *Analyzer) checkSemanticValidity(fsm *parser.FsmSyntax) {
	a.createStateEventAndActionLists(fsm)
	a.checkUndefinedStates(fsm)
	a.checkForUnusedStates(fsm)
	a.checkForDuplicateTransitions(fsm)
	a.checkThatAbstractStatesAreNotTargets(fsm)
	a.checkForInconsistentAbstraction(fsm)
	a.checkForMultiplyDefinedStateActions(fsm)
}

func (a *Analyzer) createStateEventAndActionLists(fsm *parser.FsmSyntax) {
	a.addStateNamesToStateList(fsm)
	a.addEntryAndExitActionsToActionList(fsm)
	a.addEventsToEventList(fsm)
	a.addTransitionActionsToActionList(fsm)
}

func (a *Analyzer) addTransitionActionsToActionList(fsm *parser.FsmSyntax) {
	for _, t := range fsm.Logic {
		for _, st := range t.SubTransitions {
			for _, action := range st.Actions {
				a.sm.Actions = addUnique(a.sm.Actions, action)
			}
		}
	}
}

func (a *Analyzer) addEventsToEventList(fsm *parser.FsmSyntax) {
	for _, t := range fsm.Logic {
		for _, st := range t.SubTransitions {
			if st.Event != "" {
				a.sm.Events = addUnique(a.sm.Events, st.Event)
			}
		}
	}
}

func (a *Analyzer) addEntryAndExitActionsToActionList(fsm *parser.FsmSyntax) {
	for _, t := range fsm.Logic {
		for _, entryAction := range t.State.EntryActions {
			a.sm.Actions = addUnique(a.sm.Actions, entryAction)
		}
		for _, exitAction := range t.State.ExitActions {
			a.sm.Actions = addUnique(a.sm.Actions, exitAction)
		}
	}
}

// addUnique appends s to list unless it is already there.
func addUnique(list []string, s string) []string {
	if slices.Contains(list, s) {
		return list
	}
	return append(list, s)
}

func (a *Analyzer) addStateNamesToStateList(fsm *parser.FsmSyntax) {
	for _, t := range fsm.Logic {
		a.sm.States[t.State.Name] = newState(t.State.Name)
	}
}

func (a *Analyzer) checkUndefinedStates(fsm *parser.FsmSyntax) {
	for _, t := range fsm.Logic {
		for _, superState := range t.State.SuperStates {
			a.checkUndefinedState(superState, UndefinedSuperState)
		}
		for _, st := range t.SubTransitions {
			a.checkUndefinedState(st.NextState, UndefinedState)
		}
	}

	if a.initialHeader.Value != "" && a.sm.States[a.initialHeader.Value] == nil {
		a.sm.addError(AnalysisError{UndefinedState, "initial: " + a.initialHeader.Value})
	}
}

func (a *Analyzer) checkForUnusedStates(fsm *parser.FsmSyntax) {
	a.findStatesDefinedButNotUsed(a.findUsedStates(fsm))
}

func (a *Analyzer) findUsedStates(fsm *parser.FsmSyntax) map[string]bool {
	usedStates := map[string]bool{a.initialHeader.Value: true}
	for _, t := range fsm.Logic {
		for _, superState := range t.State.SuperStates {
			usedStates[superState] = true
		}
		for _, st := range t.SubTransitions {
			if st.NextState == "" { // implicit use of current state.
				usedStates[t.State.Name] = true
			} else {
				usedStates[st.NextState] = true
			}
		}
	}
	return usedStates
}

func (a *Analyzer) findStatesDefinedButNotUsed(usedStates map[string]bool) {
	for _, definedState := range a.sm.SortedStates() {
		if !usedStates[definedState.Name] {
			a.sm.addError(AnalysisError{UnusedState, definedState.Name})
		}
	}
}

func (a *Analyzer) checkForDuplicateTransitions(fsm *parser.FsmSyntax) {
	transitionKeys := map[string]bool{}
	for _, t := range fsm.Logic {
		for _, st := range t.SubTransitions {
			key := fmt.Sprintf("%s(%s)", t.State.Name, nullable(st.Event))
			if transitionKeys[key] {
				a.sm.addError(AnalysisError{DuplicateTransition, key})
			} else {
				transitionKeys[key] = true
			}
		}
	}
}

func (a *Analyzer) checkThatAbstractStatesAreNotTargets(fsm *parser.FsmSyntax) {
	abstractStates := findAbstractStates(fsm)
	for _, t := range fsm.Logic {
		for _, st := range t.SubTransitions {
			if abstractStates[st.NextState] {
				a.sm.addError(AnalysisError{
					AbstractStateUsedAsNextState,
					fmt.Sprintf("%s(%s)->%s", t.State.Name, nullable(st.Event), st.NextState)})
			}
		}
	}
}

func findAbstractStates(fsm *parser.FsmSyntax) map[string]bool {
	abstractStates := map[string]bool{}
	for _, t := range fsm.Logic {
		if t.State.AbstractState {
			abstractStates[t.State.Name] = true
		}
	}
	return abstractStates
}

func (a *Analyzer) checkForInconsistentAbstraction(fsm *parser.FsmSyntax) {
	abstractStates := findAbstractStates(fsm)
	for _, t := range fsm.Logic {
		if !t.State.AbstractState && abstractStates[t.State.Name] {
			a.sm.Warnings = append(a.sm.Warnings, AnalysisError{InconsistentAbstraction, t.State.Name})
		}
	}
}

func (a *Analyzer) checkForMultiplyDefinedStateActions(fsm *parser.FsmSyntax) {
	firstActionsForState := map[string]string{}
	for _, t := range fsm.Logic {
		if specifiesStateActions(t) {
			actionsKey := makeActionsKey(t)
			if firstActions, ok := firstActionsForState[t.State.Name]; ok {
				if firstActions != actionsKey {
					a.sm.addError(AnalysisError{StateActionsMultiplyDefined, t.State.Name})
				}
			} else {
				firstActionsForState[t.State.Name] = actionsKey
			}
		}
	}
}

func specifiesStateActions(t *parser.Transition) bool {
	return len(t.State.EntryActions) != 0 || len(t.State.ExitActions) != 0
}

func makeActionsKey(t *parser.Transition) string {
	return strings.Join(slices.Concat(t.State.EntryActions, t.State.ExitActions), ",")
}

func (a *Analyzer) checkUndefinedState(referencedState string, errorCode ErrorID) {
	if referencedState != "" && a.sm.States[referencedState] == nil {
		a.sm.addError(AnalysisError{errorCode, referencedState})
	}
}

func (a *Analyzer) produceSemanticStateMachine(fsm *parser.FsmSyntax) {
	if len(a.sm.Errors) == 0 {
		a.compileHeaders()
		for _, t := range fsm.Logic {
			state := a.compileState(t)
			a.compileTransitions(t, state)
		}

		(&superClassCrawler{sm: a.sm}).checkSuperClassTransitions()
	}
}

func (a *Analyzer) compileHeaders() {
	a.sm.InitialState = a.sm.States[a.initialHeader.Value]
	a.sm.ActionClass = a.actionsHeader.Value
	a.sm.FsmName = a.fsmHeader.Value
}

func (a *Analyzer) compileState(t *parser.Transition) *State {
	state := a.sm.States[t.State.Name]
	state.EntryActions = append(state.EntryActions, t.State.EntryActions...)
	state.ExitActions = append(state.ExitActions, t.State.ExitActions...)
	state.AbstractState = state.AbstractState || t.State.AbstractState
	for _, superStateName := range t.State.SuperStates {
		state.addSuperState(a.sm.States[superStateName])
	}
	return state
}

func (a *Analyzer) compileTransitions(t *parser.Transition, state *State) {
	for _, st := range t.SubTransitions {
		a.compileTransition(state, st)
	}
}

func (a *Analyzer) compileTransition(state *State, st parser.SubTransition) {
	semanticTransition := &Transition{Event: st.Event, NextState: state}
	if st.NextState != "" {
		semanticTransition.NextState = a.sm.States[st.NextState]
	}
	semanticTransition.Actions = slices.Clone(st.Actions)
	state.Transitions = append(state.Transitions, semanticTransition)
}

type transitionTuple struct {
	currentState string
	event        string
	nextState    string
	actions      []string
}

// superClassCrawler reports concrete states that inherit conflicting
// transitions for the same event from unrelated super states.
type superClassCrawler struct {
	sm               *StateMachine
	concreteState    *State
	transitionTuples map[string]transitionTuple
}

func (c *superClassCrawler) checkSuperClassTransitions() {
	for _, state := range c.sm.SortedStates() {
		if !state.AbstractState {
			c.concreteState = state
			c.transitionTuples = map[string]transitionTuple{}
			c.checkTransitionsForState(c.concreteState)
		}
	}
}

func (c *superClassCrawler) checkTransitionsForState(state *State) {
	for _, superState := range state.SuperStates {
		c.checkTransitionsForState(superState)
	}
	c.checkStateForPreviouslyDefinedTransition(state)
}

func (c *superClassCrawler) checkStateForPreviouslyDefinedTransition(state *State) {
	for _, st := range state.Transitions {
		c.checkTransitionForPreviousDefinition(state, st)
	}
}

func (c *superClassCrawler) checkTransitionForPreviousDefinition(state *State, st *Transition) {
	thisTuple := transitionTuple{state.Name, st.Event, st.NextState.Name, st.Actions}
	if _, ok := c.transitionTuples[thisTuple.event]; ok {
		c.determineIfThePreviousDefinitionIsAnError(state, thisTuple)
	} else {
		c.transitionTuples[thisTuple.event] = thisTuple
	}
}

func (c *superClassCrawler) determineIfThePreviousDefinitionIsAnError(state *State, thisTuple transitionTuple) {
	previousTuple := c.transitionTuples[thisTuple.event]
	if !transitionsHaveSameOutcomes(thisTuple, previousTuple) {
		c.checkForOverriddenTransition(state, thisTuple, previousTuple)
	}
}

func (c *superClassCrawler) checkForOverriddenTransition(state *State, thisTuple, previousTuple transitionTuple) {
	definingState := c.sm.States[previousTuple.currentState]
	if !isSuperStateOf(definingState, state) {
		c.sm.addError(AnalysisError{ConflictingSuperstates, c.concreteState.Name + "|" + nullable(thisTuple.event)})
	} else {
		c.transitionTuples[thisTuple.event] = thisTuple
	}
}

func transitionsHaveSameOutcomes(t1, t2 transitionTuple) bool {
	return t1.nextState == t2.nextState && slices.Equal(t1.actions, t2.actions)
}

func isSuperStateOf(possibleSuperState, state *State) bool {
	if state == possibleSuperState {
		return true
	}
	for _, superState := range state.SuperStates {
		if isSuperStateOf(possibleSuperState, superState) {
			return true
		}
	}
	return false
}
