package implementers

import (
	"fmt"
	"strings"

	"github.com/jeb2239/CC_SMC/go/nsc"
	"github.com/jeb2239/CC_SMC/go/utilities"
)

var _ nsc.Visitor = (*DartNestedSwitchCaseImplementer)(nil)

// DartError is an error found by DartNestedSwitchCaseImplementer.
type DartError int

const (
	DartNoActions DartError = iota
)

func (e DartError) String() string { return [...]string{"NO_ACTIONS"}[e] }

// DartNestedSwitchCaseImplementer renders a Dart abstract class that extends
// the actions class.
type DartNestedSwitchCaseImplementer struct {
	fsmName     string
	actionsName string
	output      strings.Builder
	errors      []DartError
	flags       map[string]string
}

func NewDartNestedSwitchCaseImplementer(flags map[string]string) *DartNestedSwitchCaseImplementer {
	return &DartNestedSwitchCaseImplementer{flags: flags}
}

func (i *DartNestedSwitchCaseImplementer) VisitSwitchCase(n *nsc.SwitchCaseNode) {
	fmt.Fprintf(&i.output, "switch (%s) {\n", n.VariableName)
	n.GenerateCases(i)
	i.output.WriteString("}\n")
}

func (i *DartNestedSwitchCaseImplementer) VisitCase(n *nsc.CaseNode) {
	fmt.Fprintf(&i.output, "case %s.%s:\n", n.SwitchName, n.CaseName)
	n.CaseActionNode.Accept(i)
	i.output.WriteString("break;\n\n")
}

func (i *DartNestedSwitchCaseImplementer) VisitFunctionCall(n *nsc.FunctionCallNode) {
	fmt.Fprintf(&i.output, "%s(", n.FunctionName)
	if n.Argument != nil {
		n.Argument.Accept(i)
	}
	i.output.WriteString(");\n")
}

func (i *DartNestedSwitchCaseImplementer) VisitEnum(n *nsc.EnumNode) {
	fmt.Fprintf(&i.output, "\nenum %s {%s}\n\n", n.Name, utilities.CommaList(n.Enumerators))
}

func (i *DartNestedSwitchCaseImplementer) VisitStateProperty(n *nsc.StatePropertyNode) {
	i.output.WriteString("State." + n.InitialState)
}

func (i *DartNestedSwitchCaseImplementer) VisitEventDelegators(n *nsc.EventDelegatorsNode) {
	for _, event := range n.Events {
		fmt.Fprintf(&i.output, "\t%s() {_processEvent(Event.%s, \"%s\");}\n", event, event, event)
	}
}

func (i *DartNestedSwitchCaseImplementer) VisitFSMClass(n *nsc.FSMClassNode) {
	if n.ActionsName == "" {
		i.errors = append(i.errors, DartNoActions)
		return
	}

	i.fsmName = n.ClassName
	i.actionsName = n.ActionsName

	i.output.WriteString("import 'package:meta/meta.dart';\n\n")
	fmt.Fprintf(&i.output, "import '%s.dart';\n", i.actionsName)

	n.StateEnum.Accept(i)
	n.EventEnum.Accept(i)

	fmt.Fprintf(&i.output, "\n"+
		"abstract class %s extends %s {\n"+
		"\tState state;\n\n"+
		"\t%s({@required this.state = ", i.fsmName, i.actionsName, i.fsmName)
	n.StateProperty.Accept(i)
	i.output.WriteString("})\n\t : assert(state != null);\n\n")

	n.Delegators.Accept(i)
	i.output.WriteString("\n\tsetState(State s) {state=s;}\n\n")

	n.HandleEvent.Accept(i)
	i.output.WriteString("}\n")
}

func (i *DartNestedSwitchCaseImplementer) VisitHandleEvent(n *nsc.HandleEventNode) {
	i.output.WriteString("\t_processEvent(final Event event, final String eventName) {\n")
	n.SwitchCase.Accept(i)
	i.output.WriteString("}\n\n")
}

func (i *DartNestedSwitchCaseImplementer) VisitEnumerator(n *nsc.EnumeratorNode) {
	i.output.WriteString(n.Enumeration + "." + n.Enumerator)
}

func (i *DartNestedSwitchCaseImplementer) VisitDefaultCase(n *nsc.DefaultCaseNode) {
	fmt.Fprintf(&i.output, ""+
		"default:\n"+
		"unexpected_transition(\"%s\", eventName);\n"+
		"break;\n", n.State)
}

func (i *DartNestedSwitchCaseImplementer) Output() string {
	return i.output.String()
}

func (i *DartNestedSwitchCaseImplementer) Errors() []DartError {
	return i.errors
}
