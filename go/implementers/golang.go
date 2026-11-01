package implementers

import (
	"fmt"
	"strings"

	"github.com/jeb2239/CC_SMC/go/nsc"
	"github.com/jeb2239/CC_SMC/go/utilities"
)

var _ nsc.Visitor = (*GoNestedSwitchCaseImplementer)(nil)

// GoError is an error found by GoNestedSwitchCaseImplementer.
type GoError int

const (
	GoNoActions GoError = iota
)

func (e GoError) String() string { return [...]string{"NO_ACTIONS"}[e] }

// GoNestedSwitchCaseImplementer renders a Go package holding the FSM type.
// The output is meant to be run through gofmt.
type GoNestedSwitchCaseImplementer struct {
	fsmName     string
	actionsName string
	output      strings.Builder
	actions     []string
	errors      []GoError
	states      []string
	flags       map[string]string
}

func NewGoNestedSwitchCaseImplementer(flags map[string]string) *GoNestedSwitchCaseImplementer {
	return &GoNestedSwitchCaseImplementer{flags: flags}
}

func (i *GoNestedSwitchCaseImplementer) VisitSwitchCase(n *nsc.SwitchCaseNode) {
	fmt.Fprintf(&i.output, "\tswitch %s {\n", n.VariableName)
	n.GenerateCases(i)
	i.output.WriteString("}\n")
}

func (i *GoNestedSwitchCaseImplementer) VisitCase(n *nsc.CaseNode) {
	fmt.Fprintf(&i.output, "\tcase %s%s:\n", strings.ToLower(n.SwitchName), n.CaseName)
	n.CaseActionNode.Accept(i)
	i.output.WriteString("\n\n")
}

func (i *GoNestedSwitchCaseImplementer) VisitFunctionCall(n *nsc.FunctionCallNode) {
	fmt.Fprintf(&i.output, "%s(", n.FunctionName)
	if n.Argument != nil {
		n.Argument.Accept(i)
	}
	i.output.WriteString(")\n")
}

func (i *GoNestedSwitchCaseImplementer) VisitEnum(n *nsc.EnumNode) {
	name := strings.ToLower(n.Name)
	fmt.Fprintf(&i.output, "const (\n%s)\n\n",
		utilities.IotaList(name+"T", utilities.AddPrefix(name, n.Enumerators)))
}

func (i *GoNestedSwitchCaseImplementer) VisitStateProperty(n *nsc.StatePropertyNode) {
	i.output.WriteString("state" + n.InitialState)
}

func (i *GoNestedSwitchCaseImplementer) VisitEventDelegators(n *nsc.EventDelegatorsNode) {
	for _, event := range n.Events {
		fmt.Fprintf(&i.output, "func (f *%s) %s() { f.processEvent(event%s, \"%s\") }\n", i.fsmName, event, event, event)
	}
}

func (i *GoNestedSwitchCaseImplementer) VisitFSMClass(n *nsc.FSMClassNode) {
	if n.ActionsName == "" {
		i.errors = append(i.errors, GoNoActions)
		return
	}

	i.fsmName = n.ClassName
	i.actionsName = n.ActionsName
	i.actions = n.Actions
	i.states = n.States

	packageName := strings.ToLower(i.fsmName)
	fmt.Fprintf(&i.output, ""+
		"// Package %s is an auto-generated Finite State Machine.\n"+
		"// DO NOT EDIT.\n"+
		"package %s\n\n"+
		"// %s is the Finite State Machine.\n"+
		"type %s struct {\n"+
		"\tactions %s\n"+
		"\tstate stateT\n"+
		"}\n\n"+
		"// New returns a new %s.\n"+
		"func New(actions %s) *%s {\n"+
		"\t return &%s{actions: actions, state: ",
		packageName, packageName, i.fsmName, i.fsmName,
		i.actionsName, i.fsmName, i.actionsName, i.fsmName, i.fsmName)
	n.StateProperty.Accept(i)
	i.output.WriteString("}\n}\n\n")

	n.Delegators.Accept(i)

	i.output.WriteString("type stateT int\n")
	n.StateEnum.Accept(i)
	i.output.WriteString("type eventT int\n")
	n.EventEnum.Accept(i)
	n.HandleEvent.Accept(i)
}

func (i *GoNestedSwitchCaseImplementer) VisitHandleEvent(n *nsc.HandleEventNode) {
	fmt.Fprintf(&i.output, ""+
		"func (f *%s) processEvent(event eventT, eventName string) {\n"+
		"\tstate := f.state\n"+
		"\tsetState := func(s stateT) { f.state = s; state = s }\n",
		i.fsmName)

	for _, action := range i.actions {
		fmt.Fprintf(&i.output, "\t%s := func() { f.actions.%s() }\n", action, utilities.Capitalize(action))
	}
	i.output.WriteString("\n")

	for _, state := range i.states {
		fmt.Fprintf(&i.output, "\tconst State%s = state%s\n", state, utilities.Capitalize(state))
	}
	i.output.WriteString("\n")

	n.SwitchCase.Accept(i)
	i.output.WriteString("}\n\n")
}

func (i *GoNestedSwitchCaseImplementer) VisitEnumerator(n *nsc.EnumeratorNode) {
	i.output.WriteString(n.Enumeration + n.Enumerator)
}

func (i *GoNestedSwitchCaseImplementer) VisitDefaultCase(n *nsc.DefaultCaseNode) {
	fmt.Fprintf(&i.output, ""+
		"\tdefault:\n"+
		"\t\tf.actions.UnexpectedTransition(\"%s\", eventName);\n\n",
		n.State)
}

func (i *GoNestedSwitchCaseImplementer) Output() string {
	return i.output.String()
}

func (i *GoNestedSwitchCaseImplementer) Errors() []GoError {
	return i.errors
}
