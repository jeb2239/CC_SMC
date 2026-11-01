// Package implementers is a Go port of smc.implementers: nsc.Visitors that
// render a nested switch/case tree as source code in a target language.
package implementers

import (
	"fmt"
	"strings"

	"github.com/jeb2239/CC_SMC/go/nsc"
	"github.com/jeb2239/CC_SMC/go/utilities"
)

var _ nsc.Visitor = (*JavaNestedSwitchCaseImplementer)(nil)

// JavaNestedSwitchCaseImplementer renders a Java abstract class. The
// "package" flag sets the package statement.
type JavaNestedSwitchCaseImplementer struct {
	output      strings.Builder
	flags       map[string]string
	javaPackage string
}

func NewJavaNestedSwitchCaseImplementer(flags map[string]string) *JavaNestedSwitchCaseImplementer {
	return &JavaNestedSwitchCaseImplementer{flags: flags, javaPackage: flags["package"]}
}

func (i *JavaNestedSwitchCaseImplementer) VisitSwitchCase(n *nsc.SwitchCaseNode) {
	fmt.Fprintf(&i.output, "switch(%s) {\n", n.VariableName)
	n.GenerateCases(i)
	i.output.WriteString("}\n")
}

func (i *JavaNestedSwitchCaseImplementer) VisitCase(n *nsc.CaseNode) {
	fmt.Fprintf(&i.output, "case %s:\n", n.CaseName)
	n.CaseActionNode.Accept(i)
	i.output.WriteString("break;\n")
}

func (i *JavaNestedSwitchCaseImplementer) VisitFunctionCall(n *nsc.FunctionCallNode) {
	fmt.Fprintf(&i.output, "%s(", n.FunctionName)
	if n.Argument != nil {
		n.Argument.Accept(i)
	}
	i.output.WriteString(");\n")
}

func (i *JavaNestedSwitchCaseImplementer) VisitEnum(n *nsc.EnumNode) {
	fmt.Fprintf(&i.output, "private enum %s {%s}\n", n.Name, utilities.CommaList(n.Enumerators))
}

func (i *JavaNestedSwitchCaseImplementer) VisitStateProperty(n *nsc.StatePropertyNode) {
	fmt.Fprintf(&i.output, "private State state = State.%s;\n", n.InitialState)
	i.output.WriteString("private void setState(State s) {state = s;}\n")
}

func (i *JavaNestedSwitchCaseImplementer) VisitEventDelegators(n *nsc.EventDelegatorsNode) {
	for _, event := range n.Events {
		fmt.Fprintf(&i.output, "public void %s() {handleEvent(Event.%s);}\n", event, event)
	}
}

func (i *JavaNestedSwitchCaseImplementer) VisitFSMClass(n *nsc.FSMClassNode) {
	if i.javaPackage != "" {
		i.output.WriteString("package " + i.javaPackage + ";\n")
	}

	actionsName := n.ActionsName
	if actionsName == "" {
		fmt.Fprintf(&i.output, "public abstract class %s {\n", n.ClassName)
	} else {
		fmt.Fprintf(&i.output, "public abstract class %s implements %s {\n", n.ClassName, actionsName)
	}

	i.output.WriteString("public abstract void unhandledTransition(String state, String event);\n")
	n.StateEnum.Accept(i)
	n.EventEnum.Accept(i)
	n.StateProperty.Accept(i)
	n.Delegators.Accept(i)
	n.HandleEvent.Accept(i)
	if actionsName == "" {
		for _, action := range n.Actions {
			fmt.Fprintf(&i.output, "protected abstract void %s();\n", action)
		}
	}
	i.output.WriteString("}\n")
}

func (i *JavaNestedSwitchCaseImplementer) VisitHandleEvent(n *nsc.HandleEventNode) {
	i.output.WriteString("private void handleEvent(Event event) {\n")
	n.SwitchCase.Accept(i)
	i.output.WriteString("}\n")
}

func (i *JavaNestedSwitchCaseImplementer) VisitEnumerator(n *nsc.EnumeratorNode) {
	fmt.Fprintf(&i.output, "%s.%s", n.Enumeration, n.Enumerator)
}

func (i *JavaNestedSwitchCaseImplementer) VisitDefaultCase(n *nsc.DefaultCaseNode) {
	i.output.WriteString("default: unhandledTransition(state.name(), event.name()); break;\n")
}

func (i *JavaNestedSwitchCaseImplementer) Output() string {
	return i.output.String()
}
