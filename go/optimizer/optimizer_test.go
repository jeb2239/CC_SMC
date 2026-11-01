package optimizer

import (
	"slices"
	"testing"

	"github.com/jeb2239/CC_SMC/go/lexer"
	"github.com/jeb2239/CC_SMC/go/parser"
	"github.com/jeb2239/CC_SMC/go/semantic"
	"github.com/jeb2239/CC_SMC/go/utilities"
)

func produceStateMachine(fsmSyntax string) *StateMachine {
	builder := parser.NewSyntaxBuilder()
	p := parser.New(builder)
	lexer.New(p).Lex(fsmSyntax)
	p.HandleEvent(parser.EventEOF, -1, -1)
	return Optimize(semantic.NewAnalyzer().Analyze(builder.Fsm()))
}

func produceStateMachineWithHeader(s string) *StateMachine {
	return produceStateMachine("fsm:f initial:i actions:a " + s)
}

func assertOptimization(t *testing.T, syntax, stateMachine string) *StateMachine {
	t.Helper()
	sm := produceStateMachineWithHeader(syntax)
	got := utilities.CompressWhiteSpace(sm.TransitionsString())
	if want := utilities.CompressWhiteSpace(stateMachine); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
	return sm
}

func assertStrings(t *testing.T, got []string, want ...string) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestBasicOptimizerFunctions(t *testing.T) {
	t.Run("header", func(t *testing.T) {
		sm := produceStateMachineWithHeader("{i e i -}")
		if want := (Header{Initial: "i", Fsm: "f", Actions: "a"}); sm.Header != want {
			t.Errorf("Header = %+v, want %+v", sm.Header, want)
		}
	})
	t.Run("statesArePreserved", func(t *testing.T) {
		assertStrings(t, produceStateMachineWithHeader("{i e s - s e i -}").States, "i", "s")
	})
	t.Run("abstractStatesAreRemoved", func(t *testing.T) {
		if states := produceStateMachineWithHeader("{(b) - - - i:b e i -}").States; slices.Contains(states, "b") {
			t.Errorf("States = %v, should not contain b", states)
		}
	})
	t.Run("eventsArePreserved", func(t *testing.T) {
		assertStrings(t, produceStateMachineWithHeader("{i e1 s - s e2 i -}").Events, "e1", "e2")
	})
	t.Run("actionsArePreserved", func(t *testing.T) {
		assertStrings(t, produceStateMachineWithHeader("{i e1 s a1 s e2 i a2}").Actions, "a1", "a2")
	})
	t.Run("simpleStateMachine", func(t *testing.T) {
		sm := assertOptimization(t,
			"{i e i a1}",
			""+
				"i {\n"+
				"  e i {a1}\n"+
				"}\n")
		if len(sm.Transitions) != 1 {
			t.Errorf("len(Transitions) = %d, want 1", len(sm.Transitions))
		}
	})
}

func TestOptimizations(t *testing.T) {
	tests := []struct{ name, syntax, stateMachine string }{
		// Entry and exit actions.
		{"entryFunctionsAdded",
			"" +
				"{" +
				"  i e s a1" +
				"  i e2 s a2" +
				"  s <n1 <n2 e i -" +
				"}",
			"" +
				"i {\n" +
				"  e s {n1 n2 a1}\n" +
				"  e2 s {n1 n2 a2}\n" +
				"}\n" +
				"s {\n" +
				"  e i {}\n" +
				"}\n"},
		{"exitFunctionsAdded",
			"" +
				"{" +
				"  i >x2 >x1 e s a1" +
				"  i e2 s a2" +
				"  s e i -" +
				"}",
			"" +
				"i {\n" +
				"  e s {x2 x1 a1}\n" +
				"  e2 s {x2 x1 a2}\n" +
				"}\n" +
				"s {\n" +
				"  e i {}\n" +
				"}\n"},
		{"firstSuperStateEntryAndExitActionsAreAdded",
			"" +
				"{" +
				"  (ib) >ibx1 >ibx2 - - -" +
				"  (sb) <sbn1 <sbn2 - - -" +
				"  i:ib >x e s a" +
				"  s:sb <n e i -" +
				"}",
			"" +
				"i {\n" +
				"  e s {x ibx1 ibx2 sbn1 sbn2 n a}\n" +
				"}\n" +
				"s {\n" +
				"  e i {}\n" +
				"}\n"},
		{"multipleSuperStateEntryAndExitActionsAreAdded",
			"" +
				"{" +
				"  (ib1) >ib1x - - -" +
				"  (ib2) : ib1 >ib2x - - -" +
				"  (sb1) <sb1n- - -" +
				"  (sb2) :sb1 <sb2n- - -" +
				"  i:ib2 >x e s a" +
				"  s:sb2 <n e i -" +
				"}",
			"" +
				"i {\n" +
				"  e s {x ib2x ib1x sb1n sb2n n a}\n" +
				"}\n" +
				"s {\n" +
				"  e i {}\n" +
				"}\n"},
		{"diamondSuperStateEntryAndExitActionsAreAdded",
			"" +
				"{" +
				"  (ib1) >ib1x - - -" +
				"  (ib2) : ib1 >ib2x - - -" +
				"  (ib3) : ib1 >ib3x - - -" +
				"  (sb1) <sb1n - - -" +
				"  (sb2) :sb1 <sb2n - - -" +
				"  (sb3) :sb1 <sb3n - - -" +
				"  i:ib2 :ib3 >x e s a" +
				"  s :sb2 :sb3 <n e i -" +
				"}",
			"" +
				"i {\n" +
				"  e s {x ib3x ib2x ib1x sb1n sb2n sb3n n a}\n" +
				"}\n" +
				"s {\n" +
				"  e i {}\n" +
				"}\n"},

		// Super state transitions.
		{"simpleInheritanceOfTransitions",
			"" +
				"{" +
				"  (b) be s ba" +
				"  i:b e s a" +
				"  s e i -" +
				"}",
			"" +
				"i {\n" +
				"  e s {a}\n" +
				"  be s {ba}\n" +
				"}\n" +
				"s {\n" +
				"  e i {}\n" +
				"}\n"},
		{"deepInheritanceOfTransitions",
			"" +
				"{" +
				"  (b1) {" +
				"    b1e1 s b1a1" +
				"    b1e2 s b1a2" +
				"  }" +
				"  (b2):b1 b2e s b2a" +
				"  i:b2 e s a" +
				"  s e i -" +
				"}",
			"" +
				"i {\n" +
				"  e s {a}\n" +
				"  b2e s {b2a}\n" +
				"  b1e1 s {b1a1}\n" +
				"  b1e2 s {b1a2}\n" +
				"}\n" +
				"s {\n" +
				"  e i {}\n" +
				"}\n"},
		{"multipleInheritanceOfTransitions",
			"" +
				"{" +
				"  (b1) b1e s b1a" +
				"  (b2) b2e s b2a" +
				"  i:b1 :b2 e s a" +
				"  s e i -" +
				"}",
			"" +
				"i {\n" +
				"  e s {a}\n" +
				"  b2e s {b2a}\n" +
				"  b1e s {b1a}\n" +
				"}\n" +
				"s {\n" +
				"  e i {}\n" +
				"}\n"},
		{"diamondInheritanceOfTransitions",
			"" +
				"{" +
				"  (b) be s ba" +
				"  (b1):b b1e s b1a" +
				"  (b2):b b2e s b2a" +
				"  i:b1 :b2 e s a" +
				"  s e i -" +
				"}",
			"" +
				"i {\n" +
				"  e s {a}\n" +
				"  b2e s {b2a}\n" +
				"  b1e s {b1a}\n" +
				"  be s {ba}\n" +
				"}\n" +
				"s {\n" +
				"  e i {}\n" +
				"}\n"},
		{"overridingTransitions",
			"" +
				"{" +
				"  (b) e s2 a2" +
				"  i:b e s a" +
				"  s e i -" +
				"  s2 e i -" +
				"}",
			"" +
				"i {\n" +
				"  e s {a}\n" +
				"}\n" +
				"s {\n" +
				"  e i {}\n" +
				"}\n" +
				"s2 {\n" +
				"  e i {}\n" +
				"}\n"},
		{"eliminationOfDuplicateTransitions",
			"" +
				"{" +
				"  (b) e s a" +
				"  i:b e s a" +
				"  s e i -" +
				"}",
			"" +
				"i {\n" +
				"  e s {a}\n" +
				"}\n" +
				"s {\n" +
				"  e i {}\n" +
				"}\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { assertOptimization(t, tt.syntax, tt.stateMachine) })
	}
}

func TestAcceptance(t *testing.T) {
	sm := produceStateMachine("" +
		"Actions: Turnstile\n" +
		"FSM: TwoCoinTurnstile\n" +
		"Initial: Locked\n" +
		"{" +
		"    (Base)  Reset  Locked  lock" +
		"" +
		"  Locked : Base {" +
		"    Pass  Alarming  -" +
		"    Coin  FirstCoin -" +
		"  }" +
		"" +
		"  Alarming : Base <alarmOn >alarmOff -  -  -" +
		"" +
		"  FirstCoin : Base {" +
		"    Pass  Alarming  -" +
		"    Coin  Unlocked  unlock" +
		"  }" +
		"" +
		"  Unlocked : Base {" +
		"    Pass  Locked  lock" +
		"    Coin  -       thankyou" +
		"}")
	expected := "" +
		"Initial: Locked\n" +
		"Fsm: TwoCoinTurnstile\n" +
		"Actions:Turnstile\n" +
		"{\n" +
		"  Alarming {\n" +
		"    Reset Locked {alarmOff lock}\n" +
		"  }\n" +
		"  FirstCoin {\n" +
		"    Pass Alarming {alarmOn}\n" +
		"    Coin Unlocked {unlock}\n" +
		"    Reset Locked {lock}\n" +
		"  }\n" +
		"  Locked {\n" +
		"    Pass Alarming {alarmOn}\n" +
		"    Coin FirstCoin {}\n" +
		"    Reset Locked {lock}\n" +
		"  }\n" +
		"  Unlocked {\n" +
		"    Pass Locked {lock}\n" +
		"    Coin Unlocked {thankyou}\n" +
		"    Reset Locked {lock}\n" +
		"  }\n" +
		"}\n"
	if got := sm.String(); got != expected {
		t.Errorf("got:\n%s\nwant:\n%s", got, expected)
	}
}
