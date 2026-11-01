package implementers

import (
	"fmt"
	"strings"

	"github.com/jeb2239/CC_SMC/go/nsc"
	"github.com/jeb2239/CC_SMC/go/utilities"
)

var _ nsc.Visitor = (*CNestedSwitchCaseImplementer)(nil)

// CError is an error found by CNestedSwitchCaseImplementer.
type CError int

const (
	CNoAction CError = iota
)

func (e CError) String() string { return [...]string{"NO_ACTION"}[e] }

// CNestedSwitchCaseImplementer renders a C header and implementation file.
type CNestedSwitchCaseImplementer struct {
	fsmName           string
	actionsName       string
	fsmHeader         strings.Builder
	fsmImplementation strings.Builder
	errors            []CError
	flags             map[string]string
}

func NewCNestedSwitchCaseImplementer(flags map[string]string) *CNestedSwitchCaseImplementer {
	return &CNestedSwitchCaseImplementer{flags: flags}
}

func (i *CNestedSwitchCaseImplementer) VisitSwitchCase(n *nsc.SwitchCaseNode) {
	fmt.Fprintf(&i.fsmImplementation, "switch (%s) {\n", n.VariableName)
	n.GenerateCases(i)
	i.fsmImplementation.WriteString("}\n")
}

func (i *CNestedSwitchCaseImplementer) VisitCase(n *nsc.CaseNode) {
	fmt.Fprintf(&i.fsmImplementation, "case %s:\n", n.CaseName)
	n.CaseActionNode.Accept(i)
	i.fsmImplementation.WriteString("break;\n\n")
}

func (i *CNestedSwitchCaseImplementer) VisitFunctionCall(n *nsc.FunctionCallNode) {
	fmt.Fprintf(&i.fsmImplementation, "%s(fsm", n.FunctionName)
	if n.Argument != nil {
		i.fsmImplementation.WriteString(", ")
		n.Argument.Accept(i)
	}
	i.fsmImplementation.WriteString(");\n")
}

func (i *CNestedSwitchCaseImplementer) VisitEnum(n *nsc.EnumNode) {
	fmt.Fprintf(&i.fsmImplementation, "enum %s {%s};\n", n.Name, utilities.CommaList(n.Enumerators))
}

func (i *CNestedSwitchCaseImplementer) VisitStateProperty(n *nsc.StatePropertyNode) {
	fmt.Fprintf(&i.fsmImplementation, ""+
		"struct %s *make_%s(struct %s* actions) {\n"+
		"\tstruct %s *fsm = malloc(sizeof(struct %s));\n"+
		"\tfsm->actions = actions;\n"+
		"\tfsm->state = %s;\n"+
		"\treturn fsm;\n"+
		"}\n\n",
		i.fsmName, i.fsmName, i.actionsName, i.fsmName, i.fsmName, n.InitialState)

	fmt.Fprintf(&i.fsmImplementation, ""+
		"static void setState(struct %s *fsm, enum State state) {\n"+
		"\tfsm->state = state;\n"+
		"}\n\n", i.fsmName)
}

func (i *CNestedSwitchCaseImplementer) VisitEventDelegators(n *nsc.EventDelegatorsNode) {
	for _, event := range n.Events {
		fmt.Fprintf(&i.fsmHeader, "void %s_%s(struct %s*);\n", i.fsmName, event, i.fsmName)

		fmt.Fprintf(&i.fsmImplementation, ""+
			"void %s_%s(struct %s* fsm) {\n"+
			"\tprocessEvent(fsm->state, %s, fsm, \"%s\");\n"+
			"}\n", i.fsmName, event, i.fsmName, event, event)
	}
}

func (i *CNestedSwitchCaseImplementer) VisitFSMClass(n *nsc.FSMClassNode) {
	if n.ActionsName == "" {
		i.errors = append(i.errors, CNoAction)
		return
	}
	i.actionsName = n.ActionsName
	i.fsmName = n.ClassName

	i.fsmImplementation.WriteString("#include <stdlib.h>\n")
	fmt.Fprintf(&i.fsmImplementation, "#include \"%s.h\"\n", toLowerCamelCase(i.actionsName))
	fmt.Fprintf(&i.fsmImplementation, "#include \"%s.h\"\n\n", toLowerCamelCase(i.fsmName))
	n.EventEnum.Accept(i)
	n.StateEnum.Accept(i)

	fmt.Fprintf(&i.fsmImplementation, "\n"+
		"struct %s {\n"+
		"\tenum State state;\n"+
		"\tstruct %s *actions;\n"+
		"};\n\n", i.fsmName, i.actionsName)

	n.StateProperty.Accept(i)

	for _, action := range n.Actions {
		fmt.Fprintf(&i.fsmImplementation, ""+
			"static void %s(struct %s *fsm) {\n"+
			"\tfsm->actions->%s();\n"+
			"}\n\n", action, i.fsmName, action)
	}
	n.HandleEvent.Accept(i)

	includeGuard := strings.ToUpper(i.fsmName)
	fmt.Fprintf(&i.fsmHeader, "#ifndef %s_H\n#define %s_H\n\n", includeGuard, includeGuard)
	fmt.Fprintf(&i.fsmHeader, "struct %s;\n", i.actionsName)
	fmt.Fprintf(&i.fsmHeader, "struct %s;\n", i.fsmName)
	fmt.Fprintf(&i.fsmHeader, "struct %s *make_%s(struct %s*);\n", i.fsmName, i.fsmName, i.actionsName)
	n.Delegators.Accept(i)
	i.fsmHeader.WriteString("#endif\n")
}

func (i *CNestedSwitchCaseImplementer) VisitHandleEvent(n *nsc.HandleEventNode) {
	fmt.Fprintf(&i.fsmImplementation, ""+
		"static void processEvent(enum State state, enum Event event, struct %s *fsm, char *event_name) {\n",
		i.fsmName)
	n.SwitchCase.Accept(i)
	i.fsmImplementation.WriteString("}\n\n")
}

func (i *CNestedSwitchCaseImplementer) VisitEnumerator(n *nsc.EnumeratorNode) {
	i.fsmImplementation.WriteString(n.Enumerator)
}

func (i *CNestedSwitchCaseImplementer) VisitDefaultCase(n *nsc.DefaultCaseNode) {
	fmt.Fprintf(&i.fsmImplementation, ""+
		"default:\n"+
		"(fsm->actions->unexpected_transition)(\"%s\", event_name);\n"+
		"break;\n", n.State)
}

func (i *CNestedSwitchCaseImplementer) FsmHeader() string {
	return i.fsmHeader.String()
}

func (i *CNestedSwitchCaseImplementer) FsmImplementation() string {
	return i.fsmImplementation.String()
}

func (i *CNestedSwitchCaseImplementer) Errors() []CError {
	return i.errors
}

func toLowerCamelCase(s string) string {
	if len(s) < 2 {
		return strings.ToLower(s)
	}
	return strings.ToLower(s[:1]) + s[1:]
}
