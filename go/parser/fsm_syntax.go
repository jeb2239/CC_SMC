package parser

import (
	"fmt"
	"strings"
)

// FsmSyntax is the syntax tree produced by SyntaxBuilder.
//
// Where the Java version uses null for a "-" event or next state, this one
// uses the empty string; the lexer never produces an empty name.
type FsmSyntax struct {
	Headers []Header
	Logic   []*Transition
	Errors  []SyntaxError
	Done    bool
}

type Header struct {
	Name  string
	Value string
}

type Transition struct {
	State          StateSpec
	SubTransitions []SubTransition
}

type StateSpec struct {
	Name          string
	SuperStates   []string
	EntryActions  []string
	ExitActions   []string
	AbstractState bool
}

type SubTransition struct {
	Event     string
	NextState string
	Actions   []string
}

type SyntaxError struct {
	Type       ErrorType
	Msg        string
	LineNumber int
	Position   int
}

func (e SyntaxError) Error() string {
	return fmt.Sprintf("Syntax Error Line: %d, Position: %d.  (%s) %s", e.LineNumber, e.Position, e.Type, e.Msg)
}

// ErrorType classifies a SyntaxError.
type ErrorType int

const (
	TypeHeader ErrorType = iota
	TypeState
	TypeTransition
	TypeTransitionGroup
	TypeEnd
	TypeSyntax
)

var errorTypeNames = [...]string{"HEADER", "STATE", "TRANSITION", "TRANSITION_GROUP", "END", "SYNTAX"}

func (t ErrorType) String() string { return errorTypeNames[t] }

func (f *FsmSyntax) String() string {
	done := ""
	if f.Done {
		done = ".\n"
	}
	return f.formatHeaders() + f.formatLogic() + done + f.FirstError()
}

func (f *FsmSyntax) formatHeaders() string {
	var sb strings.Builder
	for _, header := range f.Headers {
		fmt.Fprintf(&sb, "%s:%s\n", header.Name, header.Value)
	}
	return sb.String()
}

func (f *FsmSyntax) formatLogic() string {
	if len(f.Logic) == 0 {
		return ""
	}
	var sb strings.Builder
	for _, transition := range f.Logic {
		fmt.Fprintf(&sb, "  %s %s\n", formatStateName(transition.State), formatSubTransitions(transition))
	}
	return fmt.Sprintf("{\n%s}\n", sb.String())
}

func formatStateName(stateSpec StateSpec) string {
	stateName := stateSpec.Name
	if stateSpec.AbstractState {
		stateName = "(" + stateName + ")"
	}
	for _, superState := range stateSpec.SuperStates {
		stateName += ":" + superState
	}
	for _, entryAction := range stateSpec.EntryActions {
		stateName += " <" + entryAction
	}
	for _, exitAction := range stateSpec.ExitActions {
		stateName += " >" + exitAction
	}
	return stateName
}

func formatSubTransitions(transition *Transition) string {
	if len(transition.SubTransitions) == 1 {
		return formatSubTransition(transition.SubTransitions[0])
	}
	formatted := "{\n"
	for _, subtransition := range transition.SubTransitions {
		formatted += "    " + formatSubTransition(subtransition) + "\n"
	}
	return formatted + "  }"
}

func formatSubTransition(subtransition SubTransition) string {
	return fmt.Sprintf("%s %s %s",
		nullable(subtransition.Event),
		nullable(subtransition.NextState),
		formatActions(subtransition))
}

func formatActions(subtransition SubTransition) string {
	if len(subtransition.Actions) == 1 {
		return subtransition.Actions[0]
	}
	return "{" + strings.Join(subtransition.Actions, " ") + "}"
}

// nullable prints an absent name the way Java prints null.
func nullable(name string) string {
	if name == "" {
		return "null"
	}
	return name
}

// FirstError formats the first syntax error, or returns "" if there are none.
func (f *FsmSyntax) FirstError() string {
	if len(f.Errors) == 0 {
		return ""
	}
	e := f.Errors[0]
	return fmt.Sprintf("Syntax error: %s. %s. line %d, position %d.\n", e.Type, e.Msg, e.LineNumber, e.Position)
}
