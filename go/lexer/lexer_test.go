package lexer

import (
	"fmt"
	"strings"
	"testing"
)

// tokenRecorder mirrors the TokenCollector in LexerTest.java.
type tokenRecorder struct {
	tokens []string
}

func (r *tokenRecorder) add(token string) { r.tokens = append(r.tokens, token) }

func (r *tokenRecorder) OpenBrace(line, pos int)         { r.add("OB") }
func (r *tokenRecorder) ClosedBrace(line, pos int)       { r.add("CB") }
func (r *tokenRecorder) OpenParen(line, pos int)         { r.add("OP") }
func (r *tokenRecorder) ClosedParen(line, pos int)       { r.add("CP") }
func (r *tokenRecorder) OpenAngle(line, pos int)         { r.add("OA") }
func (r *tokenRecorder) ClosedAngle(line, pos int)       { r.add("CA") }
func (r *tokenRecorder) Dash(line, pos int)              { r.add("D") }
func (r *tokenRecorder) Colon(line, pos int)             { r.add("C") }
func (r *tokenRecorder) Name(name string, line, pos int) { r.add("#" + name + "#") }
func (r *tokenRecorder) Error(line, pos int)             { r.add(fmt.Sprintf("E%d/%d", line, pos)) }

func assertLexResult(t *testing.T, input, expected string) {
	t.Helper()
	r := &tokenRecorder{}
	New(r).Lex(input)
	if got := strings.Join(r.tokens, ","); got != expected {
		t.Errorf("Lex(%q) = %q, want %q", input, got, expected)
	}
}

func TestSingleTokens(t *testing.T) {
	tests := []struct{ name, input, expected string }{
		{"findsOpenBrace", "{", "OB"},
		{"findsClosedBrace", "}", "CB"},
		{"findsOpenParen", "(", "OP"},
		{"findsClosedParen", ")", "CP"},
		{"findsOpenAngle", "<", "OA"},
		{"findsClosedAngle", ">", "CA"},
		{"findsDash", "-", "D"},
		{"findStarAsDash", "*", "D"},
		{"findsColon", ":", "C"},
		{"findsSimpleName", "name", "#name#"},
		{"findComplexName", "Room_222", "#Room_222#"},
		{"error", ".", "E1/1"},
		{"nothingButWhiteSpace", " ", ""},
		{"whiteSpaceBefore", "  \t\n  -", "D"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { assertLexResult(t, tt.input, tt.expected) })
	}
}

func TestComments(t *testing.T) {
	tests := []struct{ name, input, expected string }{
		{"commentAfterToken", "-//comment\n", "D"},
		{"commentLines", "//comment 1\n-//comment2\n//comment3\n-//comment4;", "D,D"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { assertLexResult(t, tt.input, tt.expected) })
	}
}

func TestMultipleTokens(t *testing.T) {
	tests := []struct{ name, input, expected string }{
		{"simpleSequence", "{}", "OB,CB"},
		{"complexSequence", "FSM:fsm{this}", "#FSM#,C,#fsm#,OB,#this#,CB"},
		{"allTokens", "{}()<>-: name .", "OB,CB,OP,CP,OA,CA,D,C,#name#,E1/15"},
		{"multipleLines", "FSM:fsm.\n{bob-.}", "#FSM#,C,#fsm#,E1/8,OB,#bob#,D,E2/6,CB"},
		// Not in LexerTest.java: a multi-byte character is one error at one column, as in Java.
		{"nonASCII", "é-.", "E1/1,D,E1/3"},
		{"carriageReturn", "{\r\n}\r\n", "OB,CB"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { assertLexResult(t, tt.input, tt.expected) })
	}
}
