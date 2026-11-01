package implementers

import (
	"fmt"
	"strings"

	"github.com/jeb2239/CC_SMC/go/nsc"
	"github.com/jeb2239/CC_SMC/go/utilities"
)

var _ nsc.Visitor = (*CppNestedSwitchCaseImplementer)(nil)

// CppError is an error found by CppNestedSwitchCaseImplementer.
type CppError int

const (
	CppNoActions CppError = iota
)

func (e CppError) String() string { return [...]string{"NO_ACTIONS"}[e] }

// CppNestedSwitchCaseImplementer renders a C++ header holding a class that
// derives from the actions class.
type CppNestedSwitchCaseImplementer struct {
	fsmName     string
	actionsName string
	output      strings.Builder
	errors      []CppError
	flags       map[string]string
}

func NewCppNestedSwitchCaseImplementer(flags map[string]string) *CppNestedSwitchCaseImplementer {
	return &CppNestedSwitchCaseImplementer{flags: flags}
}

func (i *CppNestedSwitchCaseImplementer) VisitSwitchCase(n *nsc.SwitchCaseNode) {
	fmt.Fprintf(&i.output, "switch (%s) {\n", n.VariableName)
	n.GenerateCases(i)
	i.output.WriteString("}\n")
}

func (i *CppNestedSwitchCaseImplementer) VisitCase(n *nsc.CaseNode) {
	fmt.Fprintf(&i.output, "case %s_%s:\n", n.SwitchName, n.CaseName)
	n.CaseActionNode.Accept(i)
	i.output.WriteString("break;\n\n")
}

func (i *CppNestedSwitchCaseImplementer) VisitFunctionCall(n *nsc.FunctionCallNode) {
	fmt.Fprintf(&i.output, "%s(", n.FunctionName)
	if n.Argument != nil {
		n.Argument.Accept(i)
	}
	i.output.WriteString(");\n")
}

func (i *CppNestedSwitchCaseImplementer) VisitEnum(n *nsc.EnumNode) {
	fmt.Fprintf(&i.output, "\tenum %s {%s};\n",
		n.Name,
		utilities.CommaList(utilities.AddPrefix(n.Name+"_", n.Enumerators)))
}

func (i *CppNestedSwitchCaseImplementer) VisitStateProperty(n *nsc.StatePropertyNode) {
	i.output.WriteString("State_" + n.InitialState)
}

func (i *CppNestedSwitchCaseImplementer) VisitEventDelegators(n *nsc.EventDelegatorsNode) {
	for _, event := range n.Events {
		fmt.Fprintf(&i.output, "\tvoid %s() {processEvent(Event_%s, \"%s\");}\n", event, event, event)
	}
}

func (i *CppNestedSwitchCaseImplementer) VisitFSMClass(n *nsc.FSMClassNode) {
	if n.ActionsName == "" {
		i.errors = append(i.errors, CppNoActions)
		return
	}

	i.fsmName = n.ClassName
	includeGuard := strings.ToUpper(i.fsmName)
	fmt.Fprintf(&i.output, "#ifndef %s_H\n#define %s_H\n\n", includeGuard, includeGuard)

	i.actionsName = n.ActionsName
	fmt.Fprintf(&i.output, "#include \"%s.h\"\n", i.actionsName)

	fmt.Fprintf(&i.output, "\n"+
		"class %s : public %s {\n"+
		"public:\n"+
		"\t%s()\n\t: state(", i.fsmName, i.actionsName, i.fsmName)
	n.StateProperty.Accept(i)
	i.output.WriteString(")\n\t{}\n\n")

	n.Delegators.Accept(i)
	i.output.WriteString("\nprivate:\n")
	n.StateEnum.Accept(i)
	i.output.WriteString("\tState state;\n")
	i.output.WriteString("\tvoid setState(State s) {state=s;}\n")
	n.EventEnum.Accept(i)
	n.HandleEvent.Accept(i)

	i.output.WriteString("};\n\n")
	i.output.WriteString("#endif\n")
}

func (i *CppNestedSwitchCaseImplementer) VisitHandleEvent(n *nsc.HandleEventNode) {
	i.output.WriteString("\tvoid processEvent(Event event, const char* eventName) {\n")
	n.SwitchCase.Accept(i)
	i.output.WriteString("}\n\n")
}

func (i *CppNestedSwitchCaseImplementer) VisitEnumerator(n *nsc.EnumeratorNode) {
	i.output.WriteString(n.Enumeration + "_" + n.Enumerator)
}

func (i *CppNestedSwitchCaseImplementer) VisitDefaultCase(n *nsc.DefaultCaseNode) {
	fmt.Fprintf(&i.output, ""+
		"default:\n"+
		"unexpected_transition(\"%s\", eventName);\n"+
		"break;\n", n.State)
}

func (i *CppNestedSwitchCaseImplementer) Output() string {
	return i.output.String()
}

func (i *CppNestedSwitchCaseImplementer) Errors() []CppError {
	return i.errors
}
