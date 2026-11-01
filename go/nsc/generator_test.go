package nsc

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jeb2239/CC_SMC/go/lexer"
	"github.com/jeb2239/CC_SMC/go/optimizer"
	"github.com/jeb2239/CC_SMC/go/parser"
	"github.com/jeb2239/CC_SMC/go/semantic"
)

const stdHead = "Initial: I FSM:f Actions:acts"

func produceStateMachine(fsmSyntax string) *optimizer.StateMachine {
	builder := parser.NewSyntaxBuilder()
	p := parser.New(builder)
	lexer.New(p).Lex(fsmSyntax)
	p.HandleEvent(parser.EventEOF, -1, -1)
	return optimizer.Optimize(semantic.NewAnalyzer().Analyze(builder.Fsm()))
}

// recorder is a visitor that appends to output.
type recorder interface {
	Visitor
	result() string
}

func assertGenerated(t *testing.T, implementer recorder, stt, switchCase string) {
	t.Helper()
	sm := produceStateMachine(stdHead + " " + stt)
	Generate(sm).Accept(implementer)
	if got := implementer.result(); got != switchCase {
		t.Errorf("got:\n%s\nwant:\n%s", got, switchCase)
	}
}

// javaList formats list the way Java's List.toString does.
func javaList(list []string) string {
	return "[" + strings.Join(list, ", ") + "]"
}

// emptyVisitor mirrors EmptyVisitor in NSCGeneratorTest.java. The visitors
// below embed it and override some of its methods; self is the outermost
// visitor, so nodes visited from here dispatch to those overrides as Java's
// virtual methods would.
type emptyVisitor struct {
	self   Visitor
	output string
}

func (v *emptyVisitor) result() string { return v.output }

func (v *emptyVisitor) VisitSwitchCase(n *SwitchCaseNode)           {}
func (v *emptyVisitor) VisitCase(n *CaseNode)                       {}
func (v *emptyVisitor) VisitFunctionCall(n *FunctionCallNode)       {}
func (v *emptyVisitor) VisitEnum(n *EnumNode)                       {}
func (v *emptyVisitor) VisitStateProperty(n *StatePropertyNode)     {}
func (v *emptyVisitor) VisitEventDelegators(n *EventDelegatorsNode) {}

func (v *emptyVisitor) VisitHandleEvent(n *HandleEventNode) {
	n.SwitchCase.Accept(v.self)
}

func (v *emptyVisitor) VisitEnumerator(n *EnumeratorNode) {
	v.output += n.Enumeration + "." + n.Enumerator
}

func (v *emptyVisitor) VisitDefaultCase(n *DefaultCaseNode) {
	v.output += fmt.Sprintf(" default(%s);", n.State)
}

func (v *emptyVisitor) VisitFSMClass(n *FSMClassNode) {
	n.Delegators.Accept(v.self)
	n.StateEnum.Accept(v.self)
	n.EventEnum.Accept(v.self)
	n.StateProperty.Accept(v.self)
	n.HandleEvent.Accept(v.self)
}

type testVisitor struct{ emptyVisitor }

func newTestVisitor() *testVisitor {
	v := &testVisitor{}
	v.self = v
	return v
}

func (v *testVisitor) VisitSwitchCase(n *SwitchCaseNode) {
	v.output += fmt.Sprintf("s %s {", n.VariableName)
	n.GenerateCases(v)
	v.output += "}"
}

func (v *testVisitor) VisitCase(n *CaseNode) {
	v.output += "case " + n.CaseName
	v.output += " {"
	n.CaseActionNode.Accept(v)
	v.output += "}"
}

func (v *testVisitor) VisitFunctionCall(n *FunctionCallNode) {
	v.output += fmt.Sprintf("%s(", n.FunctionName)
	if n.Argument != nil {
		n.Argument.Accept(v)
	}
	v.output += ") "
}

func TestSwitchCase(t *testing.T) {
	tests := []struct{ name, stt, switchCase string }{
		{"OneTransition",
			"{I e I a}",
			"s state {case I {s event {case e {setState(State.I) a() } default(I);}}}"},
		{"twoTransitions",
			"{I e1 S a1 S e2 I a2}",
			"" +
				"s state {" +
				"case I {s event {case e1 {setState(State.S) a1() } default(I);}}" +
				"case S {s event {case e2 {setState(State.I) a2() } default(S);}}" +
				"}"},
		{"twoStatesTwoEventsFourActions",
			"" +
				"{" +
				"  I e1 S a1 " +
				"  I e2 - a2" +
				"  S e1 I a3" +
				"  S e2 - a4" +
				"}",
			"" +
				"s state {" +
				"case I {s event {case e1 {setState(State.S) a1() }" +
				"case e2 {setState(State.I) a2() } default(I);}}" +
				"case S {s event {case e1 {setState(State.I) a3() }" +
				"case e2 {setState(State.S) a4() } default(S);}}}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { assertGenerated(t, newTestVisitor(), tt.stt, tt.switchCase) })
	}
}

type enumVisitor struct{ emptyVisitor }

func (v *enumVisitor) VisitEnum(n *EnumNode) {
	v.output += fmt.Sprintf("enum %s %s ", n.Name, javaList(n.Enumerators))
}

func TestEnum(t *testing.T) {
	v := &enumVisitor{}
	v.self = v
	assertGenerated(t, v,
		""+
			"{"+
			"  I e1 S a1 "+
			"  I e2 - a2"+
			"  S e1 I a3"+
			"  S e2 - a4"+
			"}",
		"enum State [I, S] enum Event [e1, e2] ")
}

type statePropertyVisitor struct{ emptyVisitor }

func (v *statePropertyVisitor) VisitStateProperty(n *StatePropertyNode) {
	v.output += "state property = " + n.InitialState
}

func TestStateProperty(t *testing.T) {
	v := &statePropertyVisitor{}
	v.self = v
	assertGenerated(t, v, "{I e I a}", "state property = I")
}

type eventDelegatorVisitor struct{ emptyVisitor }

func (v *eventDelegatorVisitor) VisitEventDelegators(n *EventDelegatorsNode) {
	v.output += "delegators " + javaList(n.Events)
}

func TestEventDelegators(t *testing.T) {
	v := &eventDelegatorVisitor{}
	v.self = v
	assertGenerated(t, v,
		""+
			"{"+
			"  I e1 S a1 "+
			"  I e2 - a2"+
			"  S e1 I a3"+
			"  S e2 - a4"+
			"}",
		"delegators [e1, e2]")
}

type handleEventVisitor struct{ emptyVisitor }

func (v *handleEventVisitor) VisitSwitchCase(n *SwitchCaseNode) {
	v.output += "s"
}

func (v *handleEventVisitor) VisitHandleEvent(n *HandleEventNode) {
	v.output += "he("
	n.SwitchCase.Accept(v)
	v.output += ")"
}

func TestHandleEvent(t *testing.T) {
	v := &handleEventVisitor{}
	v.self = v
	assertGenerated(t, v, "{I e I a}", "he(s)")
}

type fsmClassVisitor struct{ emptyVisitor }

func (v *fsmClassVisitor) VisitSwitchCase(n *SwitchCaseNode)           { v.output += "sc" }
func (v *fsmClassVisitor) VisitEnum(n *EnumNode)                       { v.output += "e " }
func (v *fsmClassVisitor) VisitStateProperty(n *StatePropertyNode)     { v.output += "p " }
func (v *fsmClassVisitor) VisitEventDelegators(n *EventDelegatorsNode) { v.output += "d " }
func (v *fsmClassVisitor) VisitHandleEvent(n *HandleEventNode)         { v.output += "he " }

func (v *fsmClassVisitor) VisitFSMClass(n *FSMClassNode) {
	v.output += fmt.Sprintf("class %s:%s {", n.ClassName, n.ActionsName)
	n.Delegators.Accept(v)
	n.StateEnum.Accept(v)
	n.EventEnum.Accept(v)
	n.StateProperty.Accept(v)
	n.HandleEvent.Accept(v)
	n.HandleEvent.SwitchCase.Accept(v)
	v.output += "}"
}

func TestFsmClass(t *testing.T) {
	v := &fsmClassVisitor{}
	v.self = v
	assertGenerated(t, v, "{I e I a}", "class f:acts {d e e p he sc}")
}
