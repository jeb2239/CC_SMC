package parser

import (
	"testing"

	"github.com/jeb2239/CC_SMC/go/lexer"
)

func parse(s string) *FsmSyntax {
	builder := NewSyntaxBuilder()
	parser := New(builder)
	lexer.New(parser).Lex(s)
	parser.HandleEvent(EventEOF, -1, -1)
	return builder.Fsm()
}

func assertParseResult(t *testing.T, s, expected string) {
	t.Helper()
	if got := parse(s).String(); got != expected {
		t.Errorf("parse(%q):\ngot:\n%s\nwant:\n%s", s, got, expected)
	}
}

func assertParseError(t *testing.T, s, expected string) {
	t.Helper()
	if got := parse(s).FirstError(); got != expected {
		t.Errorf("parse(%q) error = %q, want %q", s, got, expected)
	}
}

func TestIncremental(t *testing.T) {
	tests := []struct{ name, input, expected string }{
		{"parseOneHeader", "N:V{}", "N:V\n.\n"},
		{"parseManyHeaders", "  N1 : V1\tN2 : V2\n{}", "N1:V1\nN2:V2\n.\n"},
		{"noHeader", " {}", ".\n"},
		{"simpleTransition", "{ s e ns a }", "{\n  s e ns a\n}\n.\n"},
		{"transitionWithNullAction", "{s e ns -}", "{\n  s e ns {}\n}\n.\n"},
		{"transitionWithManyActions", "{s e ns {a1 a2}}", "{\n  s e ns {a1 a2}\n}\n.\n"},
		{"stateWithSubTransition", "{s {e ns a}}", "{\n  s e ns a\n}\n.\n"},
		{"stateWithSeveralSubTransitions", "{s {e1 ns a1 e2 ns a2}}",
			"{\n  s {\n    e1 ns a1\n    e2 ns a2\n  }\n}\n.\n"},
		{"manyTransitions", "{s1 e1 s2 a1 s2 e2 s3 a2}", "{\n  s1 e1 s2 a1\n  s2 e2 s3 a2\n}\n.\n"},
		{"superState", "{(ss) e s a}", "{\n  (ss) e s a\n}\n.\n"},
		{"entryAction", "{s <ea e ns a}", "{\n  s <ea e ns a\n}\n.\n"},
		{"exitAction", "{s >xa e ns a}", "{\n  s >xa e ns a\n}\n.\n"},
		{"derivedState", "{s:ss e ns a}", "{\n  s:ss e ns a\n}\n.\n"},
		{"allStateAdornments", "{(s)<ea>xa:ss e ns a}", "{\n  (s):ss <ea >xa e ns a\n}\n.\n"},
		{"stateWithNoSubTransitions", "{s {}}", "{\n  s {\n  }\n}\n.\n"},
		{"stateWithAllDashes", "{s - - -}", "{\n  s null null {}\n}\n.\n"},
		{"multipleSuperStates", "{s :x :y - - -}", "{\n  s:x:y null null {}\n}\n.\n"},
		{"multipleEntryActions", "{s <x <y - - -}", "{\n  s <x <y null null {}\n}\n.\n"},
		{"multipleExitActions", "{s >x >y - - -}", "{\n  s >x >y null null {}\n}\n.\n"},
		{"multipleEntryAndExitActionsWithBraces", "{s <{u v} >{w x} - - -}",
			"{\n  s <u <v >w >x null null {}\n}\n.\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { assertParseResult(t, tt.input, tt.expected) })
	}
}

func TestAcceptance(t *testing.T) {
	tests := []struct{ name, input, expected string }{
		{"simpleOneCoinTurnstile",
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
				"Actions:Turnstile\n" +
				"FSM:OneCoinTurnstile\n" +
				"Initial:Locked\n" +
				"{\n" +
				"  Locked Coin Unlocked {alarmOff unlock}\n" +
				"  Locked Pass Locked alarmOn\n" +
				"  Unlocked Coin Unlocked thankyou\n" +
				"  Unlocked Pass Locked lock\n" +
				"}\n" +
				".\n"},
		{"twoCoinTurnstileWithoutSuperState",
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
				"Actions:Turnstile\n" +
				"FSM:TwoCoinTurnstile\n" +
				"Initial:Locked\n" +
				"{\n" +
				"  Locked {\n" +
				"    Pass Alarming alarmOn\n" +
				"    Coin FirstCoin {}\n" +
				"    Reset Locked {lock alarmOff}\n" +
				"  }\n" +
				"  Alarming Reset Locked {lock alarmOff}\n" +
				"  FirstCoin {\n" +
				"    Pass Alarming {}\n" +
				"    Coin Unlocked unlock\n" +
				"    Reset Locked {lock alarmOff}\n" +
				"  }\n" +
				"  Unlocked {\n" +
				"    Pass Locked lock\n" +
				"    Coin null thankyou\n" +
				"    Reset Locked {lock alarmOff}\n" +
				"  }\n" +
				"}\n" +
				".\n"},
		{"twoCoinTurnstileWithSuperState",
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
				"Actions:Turnstile\n" +
				"FSM:TwoCoinTurnstile\n" +
				"Initial:Locked\n" +
				"{\n" +
				"  (Base) Reset Locked lock\n" +
				"  Locked:Base {\n" +
				"    Pass Alarming {}\n" +
				"    Coin FirstCoin {}\n" +
				"  }\n" +
				"  Alarming:Base <alarmOn >alarmOff null null {}\n" +
				"  FirstCoin:Base {\n" +
				"    Pass Alarming {}\n" +
				"    Coin Unlocked unlock\n" +
				"  }\n" +
				"  Unlocked:Base {\n" +
				"    Pass Locked lock\n" +
				"    Coin null thankyou\n" +
				"  }\n" +
				"}\n" +
				".\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { assertParseResult(t, tt.input, tt.expected) })
	}
}

func TestErrors(t *testing.T) {
	tests := []struct{ name, input, expected string }{
		{"parseNothing", "", "Syntax error: HEADER. HEADER|EOF. line -1, position -1.\n"},
		{"headerWithNoColonOrValue", "A {s e ns a}",
			"Syntax error: HEADER. HEADER_COLON|OPEN_BRACE. line 1, position 2.\n"},
		{"headerWithNoValue", "A: {s e ns a}",
			"Syntax error: HEADER. HEADER_VALUE|OPEN_BRACE. line 1, position 3.\n"},
		{"transitionWayTooShort", "{s}",
			"Syntax error: STATE. STATE_MODIFIER|CLOSED_BRACE. line 1, position 2.\n"},
		{"transitionTooShort", "{s e}",
			"Syntax error: TRANSITION. SINGLE_EVENT|CLOSED_BRACE. line 1, position 4.\n"},
		{"transitionNoAction", "{s e ns}",
			"Syntax error: TRANSITION. SINGLE_NEXT_STATE|CLOSED_BRACE. line 1, position 7.\n"},
		{"noClosingBrace", "{", "Syntax error: STATE. STATE_SPEC|EOF. line -1, position -1.\n"},
		{"initialStateDash", "{- e ns a}", "Syntax error: STATE. STATE_SPEC|DASH. line 1, position 1.\n"},
		{"lexicalError", "{.}", "Syntax error: SYNTAX. . line 1, position 2.\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { assertParseError(t, tt.input, tt.expected) })
	}
}
