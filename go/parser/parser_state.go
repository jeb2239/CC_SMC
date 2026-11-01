package parser

// ParserState is a state of the Parser's state machine.
type ParserState int

const (
	StateHeader ParserState = iota
	StateHeaderColon
	StateHeaderValue
	StateStateSpec
	StateSuperStateName
	StateSuperStateClose
	StateStateModifier
	StateExitAction
	StateEntryAction
	StateStateBase
	StateSingleEvent
	StateSingleNextState
	StateSingleActionGroup
	StateSingleActionGroupName
	StateSubtransitionGroup
	StateGroupEvent
	StateGroupNextState
	StateGroupActionGroup
	StateGroupActionGroupName
	StateMultipleEntryActions
	StateMultipleExitActions
	StateEnd
)

// The names match the Java enum so error messages are identical.
var parserStateNames = [...]string{
	"HEADER",
	"HEADER_COLON",
	"HEADER_VALUE",
	"STATE_SPEC",
	"SUPER_STATE_NAME",
	"SUPER_STATE_CLOSE",
	"STATE_MODIFIER",
	"EXIT_ACTION",
	"ENTRY_ACTION",
	"STATE_BASE",
	"SINGLE_EVENT",
	"SINGLE_NEXT_STATE",
	"SINGLE_ACTION_GROUP",
	"SINGLE_ACTION_GROUP_NAME",
	"SUBTRANSITION_GROUP",
	"GROUP_EVENT",
	"GROUP_NEXT_STATE",
	"GROUP_ACTION_GROUP",
	"GROUP_ACTION_GROUP_NAME",
	"MULTIPLE_ENTRY_ACTIONS",
	"MULTIPLE_EXIT_ACTIONS",
	"END",
}

func (s ParserState) String() string { return parserStateNames[s] }
