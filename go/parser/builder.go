package parser

// Builder receives the semantic actions fired by the Parser.
type Builder interface {
	NewHeaderWithName()
	AddHeaderWithValue()
	SetStateName()
	Done()
	SetSuperStateName()
	SetEvent()
	SetNullEvent()
	SetEntryAction()
	SetExitAction()
	SetStateBase()
	SetNextState()
	SetNullNextState()
	TransitionWithAction()
	TransitionNullAction()
	AddAction()
	TransitionWithActions()
	HeaderError(state ParserState, event ParserEvent, line, pos int)
	StateSpecError(state ParserState, event ParserEvent, line, pos int)
	TransitionError(state ParserState, event ParserEvent, line, pos int)
	TransitionGroupError(state ParserState, event ParserEvent, line, pos int)
	EndError(state ParserState, event ParserEvent, line, pos int)
	SyntaxError(line, pos int)
	SetName(name string)
}
