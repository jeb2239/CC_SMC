package optimizer

import (
	"fmt"
	"strings"
)

// StateMachine is the final output of the finite state machine parser, a
// port of smc.OptimizedStateMachine. Code generators use it as their input.
type StateMachine struct {
	States      []string
	Events      []string
	Actions     []string
	Header      Header
	Transitions []*Transition
}

type Header struct {
	Initial string
	Fsm     string
	Actions string
}

type Transition struct {
	CurrentState   string
	SubTransitions []*SubTransition
}

type SubTransition struct {
	Event     string
	NextState string
	Actions   []string
}

// TransitionsString formats every transition.
func (sm *StateMachine) TransitionsString() string {
	var sb strings.Builder
	for _, t := range sm.Transitions {
		sb.WriteString(t.String())
	}
	return sb.String()
}

func (sm *StateMachine) String() string {
	transitionsString := strings.ReplaceAll(sm.TransitionsString(), "\n", "\n  ")
	transitionsString = strings.TrimSuffix(transitionsString, "  ")
	return fmt.Sprintf("Initial: %s\nFsm: %s\nActions:%s\n{\n  %s}\n",
		sm.Header.Initial, sm.Header.Fsm, sm.Header.Actions, transitionsString)
}

func (t *Transition) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s {\n", t.CurrentState)
	for _, st := range t.SubTransitions {
		sb.WriteString(st.String())
	}
	sb.WriteString("}\n")
	return sb.String()
}

func (st *SubTransition) String() string {
	return fmt.Sprintf("  %s %s {%s}\n", st.Event, st.NextState, strings.Join(st.Actions, " "))
}
