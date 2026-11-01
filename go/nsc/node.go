// Package nsc is a Go port of smc.generators.nestedSwitchCaseGenerator. It
// turns an optimized state machine into a language-neutral tree of nested
// switch/case nodes that an implementing Visitor renders as source code.
package nsc

// Node is a node of the nested switch/case tree.
type Node interface {
	Accept(v Visitor)
}

// Visitor renders the nodes of the tree. CompositeNode has no Visit method;
// it visits each of its children in turn.
type Visitor interface {
	VisitSwitchCase(n *SwitchCaseNode)
	VisitCase(n *CaseNode)
	VisitFunctionCall(n *FunctionCallNode)
	VisitEnum(n *EnumNode)
	VisitStateProperty(n *StatePropertyNode)
	VisitEventDelegators(n *EventDelegatorsNode)
	VisitFSMClass(n *FSMClassNode)
	VisitHandleEvent(n *HandleEventNode)
	VisitEnumerator(n *EnumeratorNode)
	VisitDefaultCase(n *DefaultCaseNode)
}

type SwitchCaseNode struct {
	VariableName string
	CaseNodes    []Node
}

func (n *SwitchCaseNode) Accept(v Visitor) { v.VisitSwitchCase(n) }

// GenerateCases visits each case in turn.
func (n *SwitchCaseNode) GenerateCases(v Visitor) {
	for _, c := range n.CaseNodes {
		c.Accept(v)
	}
}

type CaseNode struct {
	SwitchName     string
	CaseName       string
	CaseActionNode Node
}

func (n *CaseNode) Accept(v Visitor) { v.VisitCase(n) }

// FunctionCallNode is a call of FunctionName, with an optional Argument.
type FunctionCallNode struct {
	FunctionName string
	Argument     Node
}

func (n *FunctionCallNode) Accept(v Visitor) { v.VisitFunctionCall(n) }

type CompositeNode struct {
	nodes []Node
}

func (n *CompositeNode) Accept(v Visitor) {
	for _, node := range n.nodes {
		node.Accept(v)
	}
}

func (n *CompositeNode) Add(node Node) {
	n.nodes = append(n.nodes, node)
}

type EnumNode struct {
	Name        string
	Enumerators []string
}

func (n *EnumNode) Accept(v Visitor) { v.VisitEnum(n) }

type StatePropertyNode struct {
	InitialState string
}

func (n *StatePropertyNode) Accept(v Visitor) { v.VisitStateProperty(n) }

type EventDelegatorsNode struct {
	Events []string
}

func (n *EventDelegatorsNode) Accept(v Visitor) { v.VisitEventDelegators(n) }

// FSMClassNode is the root of the tree. ActionsName is empty when the FSM
// has no Actions header.
type FSMClassNode struct {
	Delegators    *EventDelegatorsNode
	EventEnum     *EnumNode
	StateEnum     *EnumNode
	StateProperty *StatePropertyNode
	HandleEvent   *HandleEventNode
	ClassName     string
	ActionsName   string
	Actions       []string
	States        []string
}

func (n *FSMClassNode) Accept(v Visitor) { v.VisitFSMClass(n) }

type HandleEventNode struct {
	SwitchCase *SwitchCaseNode
}

func (n *HandleEventNode) Accept(v Visitor) { v.VisitHandleEvent(n) }

type EnumeratorNode struct {
	Enumeration string
	Enumerator  string
}

func (n *EnumeratorNode) Accept(v Visitor) { v.VisitEnumerator(n) }

type DefaultCaseNode struct {
	State string
}

func (n *DefaultCaseNode) Accept(v Visitor) { v.VisitDefaultCase(n) }
