// Package optimizer is a Go port of smc.optimizer. It flattens super states
// and entry and exit actions out of a semantic.StateMachine.
package optimizer

import (
	"slices"

	"github.com/jeb2239/CC_SMC/go/semantic"
)

// Optimize turns ast, which must have no Errors, into a StateMachine.
func Optimize(ast *semantic.StateMachine) *StateMachine {
	sm := &StateMachine{}
	addHeader(sm, ast)
	addLists(sm, ast)
	addTransitions(sm, ast)
	return sm
}

func addTransitions(sm *StateMachine, ast *semantic.StateMachine) {
	for _, s := range ast.SortedStates() {
		if !s.AbstractState {
			(&stateOptimizer{currentState: s, eventsForThisState: map[string]bool{}}).addTransitionsForState(sm)
		}
	}
}

type stateOptimizer struct {
	currentState       *semantic.State
	eventsForThisState map[string]bool
}

func (o *stateOptimizer) addTransitionsForState(sm *StateMachine) {
	transition := &Transition{CurrentState: o.currentState.Name}
	o.addSubTransitions(transition)
	sm.Transitions = append(sm.Transitions, transition)
}

func (o *stateOptimizer) addSubTransitions(transition *Transition) {
	for _, stateInHierarchy := range o.makeRootFirstHierarchyOfStates() {
		o.addStateTransitions(transition, stateInHierarchy)
	}
}

func (o *stateOptimizer) makeRootFirstHierarchyOfStates() []*semantic.State {
	hierarchy := addAllStatesInHierarchyLeafFirst(o.currentState, nil)
	slices.Reverse(hierarchy)
	return hierarchy
}

func (o *stateOptimizer) addStateTransitions(transition *Transition, state *semantic.State) {
	for _, semanticTransition := range state.Transitions {
		if o.eventExistsAndHasNotBeenOverridden(semanticTransition.Event) {
			o.addSubTransition(semanticTransition, transition)
		}
	}
}

func (o *stateOptimizer) eventExistsAndHasNotBeenOverridden(event string) bool {
	return event != "" && !o.eventsForThisState[event]
}

func (o *stateOptimizer) addSubTransition(semanticTransition *semantic.Transition, transition *Transition) {
	o.eventsForThisState[semanticTransition.Event] = true
	subTransition := &SubTransition{}
	o.optimizeSubTransition(semanticTransition, subTransition)
	transition.SubTransitions = append(transition.SubTransitions, subTransition)
}

func (o *stateOptimizer) optimizeSubTransition(semanticTransition *semantic.Transition, subTransition *SubTransition) {
	subTransition.Event = semanticTransition.Event
	subTransition.NextState = semanticTransition.NextState.Name
	addExitActions(subTransition, o.currentState)
	addEntryActions(subTransition, semanticTransition.NextState)
	subTransition.Actions = append(subTransition.Actions, semanticTransition.Actions...)
}

func addEntryActions(subTransition *SubTransition, entryState *semantic.State) {
	for _, superState := range addAllStatesInHierarchyLeafFirst(entryState, nil) {
		subTransition.Actions = append(subTransition.Actions, superState.EntryActions...)
	}
}

func addExitActions(subTransition *SubTransition, exitState *semantic.State) {
	hierarchy := addAllStatesInHierarchyLeafFirst(exitState, nil)
	slices.Reverse(hierarchy)
	for _, superState := range hierarchy {
		subTransition.Actions = append(subTransition.Actions, superState.ExitActions...)
	}
}

func addAllStatesInHierarchyLeafFirst(state *semantic.State, hierarchy []*semantic.State) []*semantic.State {
	for _, superState := range state.SuperStates {
		if !slices.Contains(hierarchy, superState) {
			hierarchy = addAllStatesInHierarchyLeafFirst(superState, hierarchy)
		}
	}
	return append(hierarchy, state)
}

func addHeader(sm *StateMachine, ast *semantic.StateMachine) {
	sm.Header = Header{
		Fsm:     ast.FsmName,
		Initial: ast.InitialState.Name,
		Actions: ast.ActionClass,
	}
}

func addLists(sm *StateMachine, ast *semantic.StateMachine) {
	for _, s := range ast.SortedStates() {
		if !s.AbstractState {
			sm.States = append(sm.States, s.Name)
		}
	}
	sm.Events = slices.Clone(ast.Events)
	sm.Actions = slices.Clone(ast.Actions)
}
