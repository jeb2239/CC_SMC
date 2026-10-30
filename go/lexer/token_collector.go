package lexer

// TokenCollector receives the tokens found by a Lexer.
type TokenCollector interface {
	OpenBrace(line, pos int)
	ClosedBrace(line, pos int)
	OpenParen(line, pos int)
	ClosedParen(line, pos int)
	OpenAngle(line, pos int)
	ClosedAngle(line, pos int)
	Dash(line, pos int)
	Colon(line, pos int)
	Name(name string, line, pos int)
	Error(line, pos int)
}
