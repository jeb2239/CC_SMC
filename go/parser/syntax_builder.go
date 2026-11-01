package parser

var _ Builder = (*SyntaxBuilder)(nil)

// SyntaxBuilder is the Builder that assembles an FsmSyntax.
type SyntaxBuilder struct {
	fsm           FsmSyntax
	header        Header
	parsedName    string
	transition    *Transition
	subtransition SubTransition
}

// NewSyntaxBuilder returns an empty SyntaxBuilder.
func NewSyntaxBuilder() *SyntaxBuilder {
	return &SyntaxBuilder{}
}

func (b *SyntaxBuilder) NewHeaderWithName() {
	b.header = Header{Name: b.parsedName}
}

func (b *SyntaxBuilder) AddHeaderWithValue() {
	b.header.Value = b.parsedName
	b.fsm.Headers = append(b.fsm.Headers, b.header)
}

func (b *SyntaxBuilder) SetStateName() {
	b.transition = &Transition{State: StateSpec{Name: b.parsedName}}
	b.fsm.Logic = append(b.fsm.Logic, b.transition)
}

func (b *SyntaxBuilder) Done() {
	b.fsm.Done = true
}

func (b *SyntaxBuilder) SetSuperStateName() {
	b.SetStateName()
	b.transition.State.AbstractState = true
}

func (b *SyntaxBuilder) SetEvent() {
	b.subtransition = SubTransition{Event: b.parsedName}
}

func (b *SyntaxBuilder) SetNullEvent() {
	b.subtransition = SubTransition{}
}

func (b *SyntaxBuilder) SetEntryAction() {
	b.transition.State.EntryActions = append(b.transition.State.EntryActions, b.parsedName)
}

func (b *SyntaxBuilder) SetExitAction() {
	b.transition.State.ExitActions = append(b.transition.State.ExitActions, b.parsedName)
}

func (b *SyntaxBuilder) SetStateBase() {
	b.transition.State.SuperStates = append(b.transition.State.SuperStates, b.parsedName)
}

func (b *SyntaxBuilder) SetNextState() {
	b.subtransition.NextState = b.parsedName
}

func (b *SyntaxBuilder) SetNullNextState() {
	b.subtransition.NextState = ""
}

func (b *SyntaxBuilder) TransitionWithAction() {
	b.AddAction()
	b.TransitionNullAction()
}

func (b *SyntaxBuilder) TransitionNullAction() {
	b.transition.SubTransitions = append(b.transition.SubTransitions, b.subtransition)
}

func (b *SyntaxBuilder) AddAction() {
	b.subtransition.Actions = append(b.subtransition.Actions, b.parsedName)
}

func (b *SyntaxBuilder) TransitionWithActions() {
	b.TransitionNullAction()
}

func (b *SyntaxBuilder) HeaderError(state ParserState, event ParserEvent, line, pos int) {
	b.addError(TypeHeader, state, event, line, pos)
}

func (b *SyntaxBuilder) StateSpecError(state ParserState, event ParserEvent, line, pos int) {
	b.addError(TypeState, state, event, line, pos)
}

func (b *SyntaxBuilder) TransitionError(state ParserState, event ParserEvent, line, pos int) {
	b.addError(TypeTransition, state, event, line, pos)
}

func (b *SyntaxBuilder) TransitionGroupError(state ParserState, event ParserEvent, line, pos int) {
	b.addError(TypeTransitionGroup, state, event, line, pos)
}

func (b *SyntaxBuilder) EndError(state ParserState, event ParserEvent, line, pos int) {
	b.addError(TypeEnd, state, event, line, pos)
}

func (b *SyntaxBuilder) SyntaxError(line, pos int) {
	b.fsm.Errors = append(b.fsm.Errors, SyntaxError{TypeSyntax, "", line, pos})
}

func (b *SyntaxBuilder) addError(t ErrorType, state ParserState, event ParserEvent, line, pos int) {
	b.fsm.Errors = append(b.fsm.Errors, SyntaxError{t, state.String() + "|" + event.String(), line, pos})
}

func (b *SyntaxBuilder) SetName(name string) {
	b.parsedName = name
}

// Fsm returns the syntax tree built so far.
func (b *SyntaxBuilder) Fsm() *FsmSyntax {
	return &b.fsm
}
