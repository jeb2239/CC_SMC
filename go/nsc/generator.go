package nsc

import "github.com/jeb2239/CC_SMC/go/optimizer"

// Generate builds the nested switch/case tree for sm.
func Generate(sm *optimizer.StateMachine) *FSMClassNode {
	stateSwitch := &SwitchCaseNode{VariableName: "state"}
	addStateCases(stateSwitch, sm)
	return &FSMClassNode{
		ClassName:     sm.Header.Fsm,
		ActionsName:   sm.Header.Actions,
		StateEnum:     &EnumNode{"State", sm.States},
		EventEnum:     &EnumNode{"Event", sm.Events},
		Delegators:    &EventDelegatorsNode{sm.Events},
		StateProperty: &StatePropertyNode{sm.Header.Initial},
		HandleEvent:   &HandleEventNode{stateSwitch},
		Actions:       sm.Actions,
		States:        sm.States,
	}
}

func addStateCases(stateSwitch *SwitchCaseNode, sm *optimizer.StateMachine) {
	for _, t := range sm.Transitions {
		addStateCase(stateSwitch, t)
	}
}

func addStateCase(stateSwitch *SwitchCaseNode, t *optimizer.Transition) {
	stateCaseNode := &CaseNode{SwitchName: "State", CaseName: t.CurrentState}
	addEventCases(stateCaseNode, t)
	stateSwitch.CaseNodes = append(stateSwitch.CaseNodes, stateCaseNode)
}

func addEventCases(stateCaseNode *CaseNode, t *optimizer.Transition) {
	eventSwitch := &SwitchCaseNode{VariableName: "event"}
	stateCaseNode.CaseActionNode = eventSwitch
	for _, st := range t.SubTransitions {
		addEventCase(eventSwitch, st)
	}
	eventSwitch.CaseNodes = append(eventSwitch.CaseNodes, &DefaultCaseNode{t.CurrentState})
}

func addEventCase(eventSwitch *SwitchCaseNode, st *optimizer.SubTransition) {
	eventCaseNode := &CaseNode{SwitchName: "Event", CaseName: st.Event}
	addActions(st, eventCaseNode)
	eventSwitch.CaseNodes = append(eventSwitch.CaseNodes, eventCaseNode)
}

func addActions(st *optimizer.SubTransition, eventCaseNode *CaseNode) {
	actions := &CompositeNode{}
	addSetStateNode(st.NextState, actions)
	for _, action := range st.Actions {
		actions.Add(&FunctionCallNode{FunctionName: action})
	}
	eventCaseNode.CaseActionNode = actions
}

func addSetStateNode(stateName string, actions *CompositeNode) {
	enumeratorNode := &EnumeratorNode{"State", stateName}
	actions.Add(&FunctionCallNode{"setState", enumeratorNode})
}
