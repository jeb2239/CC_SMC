package implementers

import (
	"slices"
	"strings"
	"testing"

	"github.com/jeb2239/CC_SMC/go/lexer"
	"github.com/jeb2239/CC_SMC/go/nsc"
	"github.com/jeb2239/CC_SMC/go/optimizer"
	"github.com/jeb2239/CC_SMC/go/parser"
	"github.com/jeb2239/CC_SMC/go/semantic"
	"github.com/jeb2239/CC_SMC/go/utilities"
)

func produceStateMachine(fsmSyntax string) *optimizer.StateMachine {
	builder := parser.NewSyntaxBuilder()
	p := parser.New(builder)
	lexer.New(p).Lex(fsmSyntax)
	p.HandleEvent(parser.EventEOF, -1, -1)
	return optimizer.Optimize(semantic.NewAnalyzer().Analyze(builder.Fsm()))
}

func assertWhitespaceEquivalent(t *testing.T, generated, expected string) {
	t.Helper()
	got, want := utilities.CompressWhiteSpace(generated), utilities.CompressWhiteSpace(expected)
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

const oneTransitionWithActions = "" +
	"Initial: I\n" +
	"Fsm: fsm\n" +
	"Actions: acts\n" +
	"{" +
	"  I E I A" +
	"}"

const oneTransitionWithNoActions = "" +
	"Initial: I\n" +
	"Fsm: fsm\n" +
	"{" +
	"  I E I A" +
	"}"

func TestJava(t *testing.T) {
	t.Run("oneTransitionWithPackageAndActions", func(t *testing.T) {
		implementer := NewJavaNestedSwitchCaseImplementer(map[string]string{"package": "thePackage"})
		nsc.Generate(produceStateMachine(oneTransitionWithActions)).Accept(implementer)
		assertWhitespaceEquivalent(t, implementer.Output(), ""+
			"package thePackage;\n"+
			"public abstract class fsm implements acts {\n"+
			"public abstract void unhandledTransition(String state, String event);\n"+
			"  private enum State {I}\n"+
			"  private enum Event {E}\n"+
			"  private State state = State.I;\n"+
			""+
			"  private void setState(State s) {state = s;}\n"+
			"  public void E() {handleEvent(Event.E);}\n"+
			"  private void handleEvent(Event event) {\n"+
			"    switch(state) {\n"+
			"      case I:\n"+
			"        switch(event) {\n"+
			"          case E:\n"+
			"            setState(State.I);\n"+
			"            A();\n"+
			"            break;\n"+
			"          default: unhandledTransition(state.name(), event.name()); break;\n"+
			"        }\n"+
			"        break;\n"+
			"    }\n"+
			"  }\n"+
			"}\n")
	})
	t.Run("oneTransitionWithActionsButNoPackage", func(t *testing.T) {
		implementer := NewJavaNestedSwitchCaseImplementer(map[string]string{})
		nsc.Generate(produceStateMachine(oneTransitionWithActions)).Accept(implementer)
		if want := "public abstract class fsm implements acts {\n"; !strings.HasPrefix(implementer.Output(), want) {
			t.Errorf("output does not start with %q:\n%s", want, implementer.Output())
		}
	})
	t.Run("oneTransitionWithNoActionsAndNoPackage", func(t *testing.T) {
		implementer := NewJavaNestedSwitchCaseImplementer(map[string]string{})
		nsc.Generate(produceStateMachine(oneTransitionWithNoActions)).Accept(implementer)
		output := implementer.Output()
		if want := "public abstract class fsm {\n"; !strings.HasPrefix(output, want) {
			t.Errorf("output does not start with %q:\n%s", want, output)
		}
		if want := "protected abstract void A();\n"; !strings.Contains(output, want) {
			t.Errorf("output does not contain %q:\n%s", want, output)
		}
	})
}

func TestC(t *testing.T) {
	t.Run("noAction_shouldBeError", func(t *testing.T) {
		implementer := NewCNestedSwitchCaseImplementer(map[string]string{})
		nsc.Generate(produceStateMachine(oneTransitionWithNoActions)).Accept(implementer)
		if got := implementer.Errors(); !slices.Equal(got, []CError{CNoAction}) {
			t.Errorf("Errors() = %v, want [NO_ACTION]", got)
		}
	})
	t.Run("oneTransition", func(t *testing.T) {
		implementer := NewCNestedSwitchCaseImplementer(map[string]string{})
		nsc.Generate(produceStateMachine(oneTransitionWithActions)).Accept(implementer)
		assertWhitespaceEquivalent(t, implementer.FsmHeader(), ""+
			"#ifndef FSM_H\n"+
			"#define FSM_H\n"+
			"struct acts;\n"+
			"struct fsm;\n"+
			"struct fsm *make_fsm(struct acts*);\n"+
			"void fsm_E(struct fsm*);\n"+
			"#endif\n")

		assertWhitespaceEquivalent(t, implementer.FsmImplementation(), ""+
			"#include <stdlib.h>\n"+
			"#include \"acts.h\"\n"+
			"#include \"fsm.h\"\n"+
			""+
			"enum Event {E};\n"+
			"enum State {I};\n"+
			""+
			"struct fsm {\n"+
			"  enum State state;\n"+
			"  struct acts *actions;\n"+
			"};\n"+
			""+
			"struct fsm *make_fsm(struct acts* actions) {\n"+
			"  struct fsm *fsm = malloc(sizeof(struct fsm));\n"+
			"  fsm->actions = actions;\n"+
			"  fsm->state = I;\n"+
			"  return fsm;\n"+
			"}\n"+
			""+
			"static void setState(struct fsm *fsm, enum State state) {\n"+
			"  fsm->state = state;\n"+
			"}\n"+
			""+
			"static void A(struct fsm *fsm) {\n"+
			"  fsm->actions->A();\n"+
			"}\n"+
			""+
			"static void processEvent(enum State state, enum Event event, struct fsm *fsm, char *event_name) {\n"+
			"  switch (state) {\n"+
			"    case I:\n"+
			"      switch (event) {\n"+
			"        case E:\n"+
			"          setState(fsm, I);\n"+
			"          A(fsm);\n"+
			"          break;\n"+
			"        default:\n"+
			"          (fsm->actions->unexpected_transition)(\"I\", event_name);\n"+
			"          break;\n"+
			"      }\n"+
			"      break;\n"+
			"  }\n"+
			"}\n"+
			""+
			"void fsm_E(struct fsm* fsm) {\n"+
			"  processEvent(fsm->state, E, fsm, \"E\");\n"+
			"}\n")
	})
}

func TestCpp(t *testing.T) {
	t.Run("noActions_shouldBeError", func(t *testing.T) {
		implementer := NewCppNestedSwitchCaseImplementer(map[string]string{})
		nsc.Generate(produceStateMachine(oneTransitionWithNoActions)).Accept(implementer)
		if got := implementer.Errors(); !slices.Equal(got, []CppError{CppNoActions}) {
			t.Errorf("Errors() = %v, want [NO_ACTIONS]", got)
		}
	})
	t.Run("oneTransition", func(t *testing.T) {
		implementer := NewCppNestedSwitchCaseImplementer(map[string]string{})
		nsc.Generate(produceStateMachine(oneTransitionWithActions)).Accept(implementer)
		assertWhitespaceEquivalent(t, implementer.Output(), ""+
			"#ifndef FSM_H\n"+
			"#define FSM_H\n"+
			"#include \"acts.h\"\n"+
			""+
			"class fsm : public acts {\n"+
			"public:\n"+
			"  fsm()\n"+
			"  : state(State_I)\n"+
			"  {}\n"+
			""+
			"  void E() {processEvent(Event_E, \"E\");}\n"+
			""+
			"private:\n"+
			"  enum State {State_I};\n"+
			"  State state;\n"+
			""+
			"  void setState(State s) {state=s;}\n"+
			""+
			"  enum Event {Event_E};\n"+
			""+
			"  void processEvent(Event event, const char* eventName) {\n"+
			"    switch (state) {\n"+
			"      case State_I:\n"+
			"        switch (event) {\n"+
			"          case Event_E:\n"+
			"            setState(State_I);\n"+
			"            A();\n"+
			"            break;\n"+
			""+
			"          default:\n"+
			"            unexpected_transition(\"I\", eventName);\n"+
			"            break;\n"+
			"        }\n"+
			"        break;\n"+
			"    }\n"+
			"  }\n"+
			"};\n"+
			"#endif\n")
	})
}
