package semantic

import (
	"slices"
	"testing"

	"github.com/jeb2239/CC_SMC/go/lexer"
	"github.com/jeb2239/CC_SMC/go/parser"
)

func produceAst(s string) *StateMachine {
	builder := parser.NewSyntaxBuilder()
	p := parser.New(builder)
	lexer.New(p).Lex(s)
	p.HandleEvent(parser.EventEOF, -1, -1)
	return NewAnalyzer().Analyze(builder.Fsm())
}

func containsAll(errors []AnalysisError, items ...AnalysisError) bool {
	for _, item := range items {
		if !slices.Contains(errors, item) {
			return false
		}
	}
	return true
}

// assertHasItems mirrors assertThat(errors, hasItems(items...)).
func assertHasItems(t *testing.T, errors []AnalysisError, items ...AnalysisError) {
	t.Helper()
	if !containsAll(errors, items...) {
		t.Errorf("%v does not contain all of %v", errors, items)
	}
}

// assertNotHasItems mirrors assertThat(errors, not(hasItems(items...))).
func assertNotHasItems(t *testing.T, errors []AnalysisError, items ...AnalysisError) {
	t.Helper()
	if containsAll(errors, items...) {
		t.Errorf("%v contains all of %v", errors, items)
	}
}

func TestHeaderErrors(t *testing.T) {
	t.Run("noHeaders", func(t *testing.T) {
		errors := produceAst("{}").Errors
		assertHasItems(t, errors, AnalysisError{ID: NoFsm}, AnalysisError{ID: NoInitial})
	})
	t.Run("missingActions", func(t *testing.T) {
		errors := produceAst("FSM:f Initial:i {}").Errors
		assertNotHasItems(t, errors, AnalysisError{ID: NoFsm}, AnalysisError{ID: NoInitial})
	})
	t.Run("missingFsm", func(t *testing.T) {
		errors := produceAst("actions:a Initial:i {}").Errors
		assertNotHasItems(t, errors, AnalysisError{ID: NoInitial})
		assertHasItems(t, errors, AnalysisError{ID: NoFsm})
	})
	t.Run("missingInitial", func(t *testing.T) {
		errors := produceAst("Actions:a Fsm:f {}").Errors
		assertNotHasItems(t, errors, AnalysisError{ID: NoFsm})
		assertHasItems(t, errors, AnalysisError{ID: NoInitial})
	})
	t.Run("nothingMissing", func(t *testing.T) {
		errors := produceAst("Initial: f Actions:a Fsm:f {}").Errors
		assertNotHasItems(t, errors, AnalysisError{ID: NoInitial}, AnalysisError{ID: NoFsm})
	})
	t.Run("unexpectedHeader", func(t *testing.T) {
		errors := produceAst("X: x{s - - -}").Errors
		assertHasItems(t, errors, AnalysisError{InvalidHeader, parser.Header{Name: "X", Value: "x"}})
	})
	t.Run("duplicateHeader", func(t *testing.T) {
		errors := produceAst("fsm:f fsm:x{s - - -}").Errors
		assertHasItems(t, errors, AnalysisError{ExtraHeaderIgnored, parser.Header{Name: "fsm", Value: "x"}})
	})
	t.Run("initialStateMustBeDefined", func(t *testing.T) {
		errors := produceAst("initial: i {s - - -}").Errors
		assertHasItems(t, errors, AnalysisError{UndefinedState, "initial: i"})
	})
}

func TestStateErrors(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		err    AnalysisError
		wanted bool
	}{
		{"nullNextStateIsNotUndefined", "{s - - -}", AnalysisError{ID: UndefinedState}, false},
		{"undefinedState", "{s - s2 -}", AnalysisError{UndefinedState, "s2"}, true},
		{"noUndefinedStates", "{s - s -}", AnalysisError{UndefinedState, "s2"}, false},
		{"undefinedSuperState", "{s:ss - - -}", AnalysisError{UndefinedSuperState, "ss"}, true},
		{"superStateDefined", "{ss - - - s:ss - - -}", AnalysisError{UndefinedSuperState, "s2"}, false},
		{"unusedStates", "{s e n -}", AnalysisError{UnusedState, "s"}, true},
		{"noUnusedStates", "{s e s -}", AnalysisError{UnusedState, "s"}, false},
		{"nextStateNullIsImplicitUse", "{s e - -}", AnalysisError{UnusedState, "s"}, false},
		{"usedAsBaseIsValidUsage", "{b e n - s:b e2 s -}", AnalysisError{UnusedState, "b"}, false},
		{"usedAsInitialIsValidUsage", "initial: b {b e n -}", AnalysisError{UnusedState, "b"}, false},
		{"errorIfSuperStatesHaveConflictingTransitions",
			"" +
				"FSM: f Actions: act Initial: s" +
				"{" +
				"  (ss1) e1 s1 -" +
				"  (ss2) e1 s2 -" +
				"  s :ss1 :ss2 e2 s3 a" +
				"  s2 e s -" +
				"  s1 e s -" +
				"  s3 e s -" +
				"}",
			AnalysisError{ConflictingSuperstates, "s|e1"}, true},
		{"noErrorForOverriddenTransition",
			"" +
				"FSM: f Actions: act Initial: s" +
				"{" +
				"  (ss1) e1 s1 -" +
				"  s :ss1 e1 s3 a" +
				"  s1 e s -" +
				"  s3 e s -" +
				"}",
			AnalysisError{ConflictingSuperstates, "s|e1"}, false},
		{"noErrorIfSuperStatesHaveIdenticalTransitions",
			"" +
				"FSM: f Actions: act Initial: s" +
				"{" +
				"  (ss1) e1 s1 ax" +
				"  (ss2) e1 s1 ax" +
				"  s :ss1 :ss2 e2 s3 a" +
				"  s1 e s -" +
				"  s3 e s -" +
				"}",
			AnalysisError{ConflictingSuperstates, "s|e1"}, false},
		{"errorIfSuperstatesHaveDifferentActionsInSameTransitions",
			"" +
				"FSM: f Actions: act Initial: s" +
				"{" +
				"  (ss1) e1 s1 a1" +
				"  (ss2) e1 s1 a2" +
				"  s :ss1 :ss2 e2 s3 a" +
				"  s1 e s -" +
				"  s3 e s -" +
				"}",
			AnalysisError{ConflictingSuperstates, "s|e1"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := produceAst(tt.input).Errors
			if tt.wanted {
				assertHasItems(t, errors, tt.err)
			} else {
				assertNotHasItems(t, errors, tt.err)
			}
		})
	}
}

func TestTransitionErrors(t *testing.T) {
	t.Run("duplicateTransitions", func(t *testing.T) {
		errors := produceAst("{s e - - s e - -}").Errors
		assertHasItems(t, errors, AnalysisError{DuplicateTransition, "s(e)"})
	})
	t.Run("noDuplicateTransitions", func(t *testing.T) {
		errors := produceAst("{s e - -}").Errors
		assertNotHasItems(t, errors, AnalysisError{DuplicateTransition, "s(e)"})
	})
	t.Run("abstractStatesCantBeTargets", func(t *testing.T) {
		errors := produceAst("{(as) e - - s e as -}").Errors
		assertHasItems(t, errors, AnalysisError{AbstractStateUsedAsNextState, "s(e)->as"})
	})
	t.Run("abstractStatesCanBeUsedAsSuperStates", func(t *testing.T) {
		errors := produceAst("{(as) e - - s:as e s -}").Errors
		assertNotHasItems(t, errors, AnalysisError{AbstractStateUsedAsNextState, "s(e)->s"})
	})
	t.Run("entryAndExitActionsNotMultiplyDefined", func(t *testing.T) {
		errors := produceAst("" +
			"{" +
			"  s - - - " +
			"  s - - -" +
			"  es - - -" +
			"  es <x - - - " +
			"  es <x - - -" +
			"  xs >x - - -" +
			"  xs >{x} - - -" +
			"}").Errors
		assertNotHasItems(t, errors, AnalysisError{StateActionsMultiplyDefined, "s"})
		assertNotHasItems(t, errors, AnalysisError{StateActionsMultiplyDefined, "es"})
		assertNotHasItems(t, errors, AnalysisError{StateActionsMultiplyDefined, "xs"})
	})
	t.Run("errorIfStateHasMultipleEntryActionDefinitions", func(t *testing.T) {
		errors := produceAst("{s - - - ds <x - - - ds <y - - -}").Errors
		assertNotHasItems(t, errors, AnalysisError{StateActionsMultiplyDefined, "s"})
		assertHasItems(t, errors, AnalysisError{StateActionsMultiplyDefined, "ds"})
	})
	t.Run("errorIfStateHasMultipleExitActionDefinitions", func(t *testing.T) {
		errors := produceAst("{ds >x - - - ds >y - -}").Errors
		assertHasItems(t, errors, AnalysisError{StateActionsMultiplyDefined, "ds"})
	})
	t.Run("errorIfStateHasMultiplyDefinedEntryAndExitActions", func(t *testing.T) {
		errors := produceAst("{ds >x - - - ds <y - -}").Errors
		assertHasItems(t, errors, AnalysisError{StateActionsMultiplyDefined, "ds"})
	})
}

func TestWarnings(t *testing.T) {
	warnings := produceAst("{(ias) e - - ias e - - (cas) e - -}").Warnings
	assertNotHasItems(t, warnings, AnalysisError{InconsistentAbstraction, "cas"})
	assertHasItems(t, warnings, AnalysisError{InconsistentAbstraction, "ias"})
}

func stateNames(sm *StateMachine) []string {
	var names []string
	for _, s := range sm.SortedStates() {
		names = append(names, s.Name)
	}
	return names
}

func assertStrings(t *testing.T, got []string, want ...string) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestLists(t *testing.T) {
	t.Run("oneState", func(t *testing.T) {
		assertStrings(t, stateNames(produceAst("{s - - -}")), "s")
	})
	t.Run("manyStates", func(t *testing.T) {
		assertStrings(t, stateNames(produceAst("{s1 - - - s2 - - - s3 - - -}")), "s1", "s2", "s3")
	})
	t.Run("statesAreKeyedByName", func(t *testing.T) {
		ast := produceAst("{s1 - - - s2 - - - s3 - - -}")
		for _, name := range []string{"s1", "s2", "s3"} {
			if s := ast.States[name]; s == nil || s.Name != name {
				t.Errorf("States[%q] = %v", name, s)
			}
		}
	})
	t.Run("manyEvents", func(t *testing.T) {
		assertStrings(t, produceAst("{s1 e1 - - s2 e2 - - s3 e3 - -}").Events, "e1", "e2", "e3")
	})
	t.Run("manyEventsButNoDuplicates", func(t *testing.T) {
		assertStrings(t, produceAst("{s1 e1 - - s2 e2 - - s3 e1 - -}").Events, "e1", "e2")
	})
	t.Run("noNullEvents", func(t *testing.T) {
		assertStrings(t, produceAst("{(s1) - - -}").Events)
	})
	t.Run("manyActionsButNoDuplicates", func(t *testing.T) {
		assertStrings(t, produceAst("{s1 e1 - {a1 a2} s2 e2 - {a3 a1}}").Actions, "a1", "a2", "a3")
	})
	t.Run("entryAndExitActionsAreCountedAsActions", func(t *testing.T) {
		assertStrings(t, produceAst("{s <ea >xa - - a}").Actions, "ea", "xa", "a")
	})
}

func TestLogic(t *testing.T) {
	tests := []struct{ name, syntax, ast string }{
		{"oneTransition", "{s e s a}",
			"" +
				"{\n" +
				"  s {\n" +
				"    e s {a}\n" +
				"  }\n" +
				"}\n"},
		{"twoTransitionsAreAggregated", "{s e1 s a s e2 s a}",
			"" +
				"{\n" +
				"  s {\n" +
				"    e1 s {a}\n" +
				"    e2 s {a}\n" +
				"  }\n" +
				"}\n"},
		{"superStatesAreAggregated", "{s:b1 e1 s a s:b2 e2 s a (b1) e s - (b2) e s -}",
			"" +
				"{\n" +
				"  (b1) {\n" +
				"    e s {}\n" +
				"  }\n" +
				"\n" +
				"  (b2) {\n" +
				"    e s {}\n" +
				"  }\n" +
				"\n" +
				"  s :b1 :b2 {\n" +
				"    e1 s {a}\n" +
				"    e2 s {a}\n" +
				"  }\n" +
				"}\n"},
		{"nullNextStateRefersToSelf", "{s e - a}",
			"" +
				"{\n" +
				"  s {\n" +
				"    e s {a}\n" +
				"  }\n" +
				"}\n"},
		{"actionsRemainInOrder", "{s e s {the quick brown fox jumped over the lazy dogs back}}",
			"" +
				"{\n" +
				"  s {\n" +
				"    e s {the quick brown fox jumped over the lazy dogs back}\n" +
				"  }\n" +
				"}\n"},
		{"entryAndExitActionsRemainInOrder", "{s <{d o} <g >{c a} >t e s a}",
			"" +
				"{\n" +
				"  s <d <o <g >c >a >t {\n" +
				"    e s {a}\n" +
				"  }\n" +
				"}\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := produceAst("initial: s fsm:f actions:a " + tt.syntax).StatesString(); got != tt.ast {
				t.Errorf("got:\n%s\nwant:\n%s", got, tt.ast)
			}
		})
	}
}

func TestAcceptance(t *testing.T) {
	tests := []struct{ name, input, expected string }{
		{"subwayTurnstileOne",
			"" +
				"Actions: Turnstile\n" +
				"FSM: OneCoinTurnstile\n" +
				"Initial: Locked\n" +
				"{\n" +
				"  Locked\tCoin\tUnlocked\t{alarmOff unlock}\n" +
				"  Locked \tPass\tLocked\t\talarmOn\n" +
				"  Unlocked\tCoin\tUnlocked\tthankyou\n" +
				"  Unlocked\tPass\tLocked\t\tlock\n" +
				"}",
			"" +
				"Actions: Turnstile\n" +
				"FSM: OneCoinTurnstile\n" +
				"Initial: Locked{\n" +
				"  Locked {\n" +
				"    Coin Unlocked {alarmOff unlock}\n" +
				"    Pass Locked {alarmOn}\n" +
				"  }\n" +
				"\n" +
				"  Unlocked {\n" +
				"    Coin Unlocked {thankyou}\n" +
				"    Pass Locked {lock}\n" +
				"  }\n" +
				"}\n"},
		{"subwayTurnstileTwo",
			"" +
				"Actions: Turnstile\n" +
				"FSM: TwoCoinTurnstile\n" +
				"Initial: Locked\n" +
				"{\n" +
				"\tLocked {\n" +
				"\t\tPass\tAlarming\talarmOn\n" +
				"\t\tCoin\tFirstCoin\t-\n" +
				"\t\tReset\tLocked\t{lock alarmOff}\n" +
				"\t}\n" +
				"\t\n" +
				"\tAlarming\tReset\tLocked {lock alarmOff}\n" +
				"\t\n" +
				"\tFirstCoin {\n" +
				"\t\tPass\tAlarming\t-\n" +
				"\t\tCoin\tUnlocked\tunlock\n" +
				"\t\tReset\tLocked {lock alarmOff}\n" +
				"\t}\n" +
				"\t\n" +
				"\tUnlocked {\n" +
				"\t\tPass\tLocked\tlock\n" +
				"\t\tCoin\t-\t\tthankyou\n" +
				"\t\tReset\tLocked {lock alarmOff}\n" +
				"\t}\n" +
				"}",
			"" +
				"Actions: Turnstile\n" +
				"FSM: TwoCoinTurnstile\n" +
				"Initial: Locked{\n" +
				"  Alarming {\n" +
				"    Reset Locked {lock alarmOff}\n" +
				"  }\n" +
				"\n" +
				"  FirstCoin {\n" +
				"    Pass Alarming {}\n" +
				"    Coin Unlocked {unlock}\n" +
				"    Reset Locked {lock alarmOff}\n" +
				"  }\n" +
				"\n" +
				"  Locked {\n" +
				"    Pass Alarming {alarmOn}\n" +
				"    Coin FirstCoin {}\n" +
				"    Reset Locked {lock alarmOff}\n" +
				"  }\n" +
				"\n" +
				"  Unlocked {\n" +
				"    Pass Locked {lock}\n" +
				"    Coin Unlocked {thankyou}\n" +
				"    Reset Locked {lock alarmOff}\n" +
				"  }\n" +
				"}\n"},
		{"subwayTurnstileThree",
			"" +
				"Actions: Turnstile\n" +
				"FSM: TwoCoinTurnstile\n" +
				"Initial: Locked\n" +
				"{\n" +
				"    (Base)\tReset\tLocked\tlock\n" +
				"\n" +
				"\tLocked : Base {\n" +
				"\t\tPass\tAlarming\t-\n" +
				"\t\tCoin\tFirstCoin\t-\n" +
				"\t}\n" +
				"\t\n" +
				"\tAlarming : Base\t<alarmOn >alarmOff -\t-\t-\n" +
				"\t\n" +
				"\tFirstCoin : Base {\n" +
				"\t\tPass\tAlarming\t-\n" +
				"\t\tCoin\tUnlocked\tunlock\n" +
				"\t}\n" +
				"\t\n" +
				"\tUnlocked : Base {\n" +
				"\t\tPass\tLocked\tlock\n" +
				"\t\tCoin\t-\t\tthankyou\n" +
				"\t}\n" +
				"}",
			"" +
				"Actions: Turnstile\n" +
				"FSM: TwoCoinTurnstile\n" +
				"Initial: Locked{\n" +
				"  Alarming :Base <alarmOn >alarmOff {\n" +
				"    null Alarming {}\n" +
				"  }\n" +
				"\n" +
				"  (Base) {\n" +
				"    Reset Locked {lock}\n" +
				"  }\n" +
				"\n" +
				"  FirstCoin :Base {\n" +
				"    Pass Alarming {}\n" +
				"    Coin Unlocked {unlock}\n" +
				"  }\n" +
				"\n" +
				"  Locked :Base {\n" +
				"    Pass Alarming {}\n" +
				"    Coin FirstCoin {}\n" +
				"  }\n" +
				"\n" +
				"  Unlocked :Base {\n" +
				"    Pass Locked {lock}\n" +
				"    Coin Unlocked {thankyou}\n" +
				"  }\n" +
				"}\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := produceAst(tt.input).String(); got != tt.expected {
				t.Errorf("got:\n%s\nwant:\n%s", got, tt.expected)
			}
		})
	}
}
