// Package parser is a Go port of smc.parser.
package parser

/*
<FSM> ::= <header>* <logic>
<header> ::= "Actions:" <name> | "FSM:" <name> | "Initial:" <name>

<logic> ::= "{" <transition>* "}"

<transition> ::= <state-spec> <subtransition>
             |   <state-spec> "{" <subtransition>* "}"

<subtransition>   ::= <event-spec> <next-state> <action-spec>
<action-spec>     ::= <action> | "{" <action>* "}" | "-"
<state-spec>      ::= <state> <state-modifiers>
<state-modifiers> ::= "" | <state-modifier> | <state-modifier> <state-modifiers>
<state-modifier>  ::= ":" <state>
                  |   "<" <action-spec>
                  |   ">" <action-spec>

<next-state> ::= <state> | "-"
<event-spec> :: <event> | "-"
<action> ::= <name>
<state> ::= <name>
<event> ::= <name>
*/

import "github.com/jeb2239/CC_SMC/go/lexer"

var _ lexer.TokenCollector = (*Parser)(nil)

// Parser turns lexer tokens into Builder calls.
type Parser struct {
	state   ParserState
	builder Builder
}

// New returns a Parser that reports to builder.
func New(builder Builder) *Parser {
	return &Parser{state: StateHeader, builder: builder}
}

func (p *Parser) OpenBrace(line, pos int)   { p.HandleEvent(EventOpenBrace, line, pos) }
func (p *Parser) ClosedBrace(line, pos int) { p.HandleEvent(EventClosedBrace, line, pos) }
func (p *Parser) OpenParen(line, pos int)   { p.HandleEvent(EventOpenParen, line, pos) }
func (p *Parser) ClosedParen(line, pos int) { p.HandleEvent(EventClosedParen, line, pos) }
func (p *Parser) OpenAngle(line, pos int)   { p.HandleEvent(EventOpenAngle, line, pos) }
func (p *Parser) ClosedAngle(line, pos int) { p.HandleEvent(EventClosedAngle, line, pos) }
func (p *Parser) Dash(line, pos int)        { p.HandleEvent(EventDash, line, pos) }
func (p *Parser) Colon(line, pos int)       { p.HandleEvent(EventColon, line, pos) }

func (p *Parser) Name(name string, line, pos int) {
	p.builder.SetName(name)
	p.HandleEvent(EventName, line, pos)
}

func (p *Parser) Error(line, pos int) {
	p.builder.SyntaxError(line, pos)
}

type transition struct {
	currentState ParserState
	event        ParserEvent
	newState     ParserState
	action       func(Builder)
}

var transitions = []transition{
	{StateHeader, EventName, StateHeaderColon, Builder.NewHeaderWithName},
	{StateHeader, EventOpenBrace, StateStateSpec, nil},
	{StateHeaderColon, EventColon, StateHeaderValue, nil},
	{StateHeaderValue, EventName, StateHeader, Builder.AddHeaderWithValue},
	{StateStateSpec, EventOpenParen, StateSuperStateName, nil},
	{StateStateSpec, EventName, StateStateModifier, Builder.SetStateName},
	{StateStateSpec, EventClosedBrace, StateEnd, Builder.Done},
	{StateSuperStateName, EventName, StateSuperStateClose, Builder.SetSuperStateName},
	{StateSuperStateClose, EventClosedParen, StateStateModifier, nil},
	{StateStateModifier, EventOpenAngle, StateEntryAction, nil},
	{StateStateModifier, EventClosedAngle, StateExitAction, nil},
	{StateStateModifier, EventColon, StateStateBase, nil},
	{StateStateModifier, EventName, StateSingleEvent, Builder.SetEvent},
	{StateStateModifier, EventDash, StateSingleEvent, Builder.SetNullEvent},
	{StateStateModifier, EventOpenBrace, StateSubtransitionGroup, nil},
	{StateEntryAction, EventName, StateStateModifier, Builder.SetEntryAction},
	{StateEntryAction, EventOpenBrace, StateMultipleEntryActions, nil},
	{StateMultipleEntryActions, EventName, StateMultipleEntryActions, Builder.SetEntryAction},
	{StateMultipleEntryActions, EventClosedBrace, StateStateModifier, nil},
	{StateExitAction, EventName, StateStateModifier, Builder.SetExitAction},
	{StateExitAction, EventOpenBrace, StateMultipleExitActions, nil},
	{StateMultipleExitActions, EventName, StateMultipleExitActions, Builder.SetExitAction},
	{StateMultipleExitActions, EventClosedBrace, StateStateModifier, nil},
	{StateStateBase, EventName, StateStateModifier, Builder.SetStateBase},
	{StateSingleEvent, EventName, StateSingleNextState, Builder.SetNextState},
	{StateSingleEvent, EventDash, StateSingleNextState, Builder.SetNullNextState},
	{StateSingleNextState, EventName, StateStateSpec, Builder.TransitionWithAction},
	{StateSingleNextState, EventDash, StateStateSpec, Builder.TransitionNullAction},
	{StateSingleNextState, EventOpenBrace, StateSingleActionGroup, nil},
	{StateSingleActionGroup, EventName, StateSingleActionGroupName, Builder.AddAction},
	{StateSingleActionGroup, EventClosedBrace, StateStateSpec, Builder.TransitionNullAction},
	{StateSingleActionGroupName, EventName, StateSingleActionGroupName, Builder.AddAction},
	{StateSingleActionGroupName, EventClosedBrace, StateStateSpec, Builder.TransitionWithActions},
	{StateSubtransitionGroup, EventClosedBrace, StateStateSpec, nil},
	{StateSubtransitionGroup, EventName, StateGroupEvent, Builder.SetEvent},
	{StateSubtransitionGroup, EventDash, StateGroupEvent, Builder.SetNullEvent},
	{StateGroupEvent, EventName, StateGroupNextState, Builder.SetNextState},
	{StateGroupEvent, EventDash, StateGroupNextState, Builder.SetNullNextState},
	{StateGroupNextState, EventName, StateSubtransitionGroup, Builder.TransitionWithAction},
	{StateGroupNextState, EventDash, StateSubtransitionGroup, Builder.TransitionNullAction},
	{StateGroupNextState, EventOpenBrace, StateGroupActionGroup, nil},
	{StateGroupActionGroup, EventName, StateGroupActionGroupName, Builder.AddAction},
	{StateGroupActionGroup, EventClosedBrace, StateSubtransitionGroup, Builder.TransitionNullAction},
	{StateGroupActionGroupName, EventName, StateGroupActionGroupName, Builder.AddAction},
	{StateGroupActionGroupName, EventClosedBrace, StateSubtransitionGroup, Builder.TransitionWithActions},
	{StateEnd, EventEOF, StateEnd, nil},
}

// HandleEvent feeds one event to the state machine. Callers send EventEOF
// after lexing to finish the parse.
func (p *Parser) HandleEvent(event ParserEvent, line, pos int) {
	for _, t := range transitions {
		if t.currentState == p.state && t.event == event {
			p.state = t.newState
			if t.action != nil {
				t.action(p.builder)
			}
			return
		}
	}
	p.handleEventError(event, line, pos)
}

func (p *Parser) handleEventError(event ParserEvent, line, pos int) {
	switch p.state {
	case StateHeader, StateHeaderColon, StateHeaderValue:
		p.builder.HeaderError(p.state, event, line, pos)

	case StateStateSpec, StateSuperStateName, StateSuperStateClose, StateStateModifier,
		StateExitAction, StateEntryAction, StateStateBase:
		p.builder.StateSpecError(p.state, event, line, pos)

	case StateSingleEvent, StateSingleNextState, StateSingleActionGroup, StateSingleActionGroupName:
		p.builder.TransitionError(p.state, event, line, pos)

	case StateSubtransitionGroup, StateGroupEvent, StateGroupNextState,
		StateGroupActionGroup, StateGroupActionGroupName:
		p.builder.TransitionGroupError(p.state, event, line, pos)

	case StateEnd:
		p.builder.EndError(p.state, event, line, pos)
	}
}
