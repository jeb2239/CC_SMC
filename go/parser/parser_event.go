package parser

// ParserEvent is a token kind fed to the Parser's state machine.
type ParserEvent int

const (
	EventName ParserEvent = iota
	EventOpenBrace
	EventClosedBrace
	EventOpenParen
	EventClosedParen
	EventOpenAngle
	EventClosedAngle
	EventDash
	EventEOF
	EventColon
)

// The names match the Java enum so error messages are identical.
var parserEventNames = [...]string{
	"NAME",
	"OPEN_BRACE",
	"CLOSED_BRACE",
	"OPEN_PAREN",
	"CLOSED_PAREN",
	"OPEN_ANGLE",
	"CLOSED_ANGLE",
	"DASH",
	"EOF",
	"COLON",
}

func (e ParserEvent) String() string { return parserEventNames[e] }
