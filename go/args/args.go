// Package args is a Go port of com.cleancoder.args, the command line parser
// SMC uses.
//
// A schema is a comma separated list of elements. Each element is a letter
// followed by a type marker:
//
//	x     boolean
//	x*    string
//	x#    integer
//	x##   double
//	x[*]  string array (the flag may be repeated)
//	x&    map, given as key1:val1,key2:val2
//
// Flags may be grouped, as in -xy. Parsing stops at the first argument that
// does not start with "-"; NextArgument returns its index.
package args

import (
	"strings"
	"unicode"
)

// Args holds the flags parsed from a command line.
type Args struct {
	marshalers      map[rune]argumentMarshaler
	argsFound       map[rune]bool
	currentArgument *argumentIterator
}

// New parses args according to schema. The error, if any, is an *Error.
func New(schema string, args []string) (*Args, error) {
	a := &Args{
		marshalers: map[rune]argumentMarshaler{},
		argsFound:  map[rune]bool{},
	}
	if err := a.parseSchema(schema); err != nil {
		return nil, err
	}
	if err := a.parseArgumentStrings(args); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *Args) parseSchema(schema string) error {
	for _, element := range strings.Split(schema, ",") {
		if element = strings.TrimSpace(element); len(element) > 0 {
			if err := a.parseSchemaElement(element); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *Args) parseSchemaElement(element string) error {
	elementID := []rune(element)[0]
	elementTail := string([]rune(element)[1:])
	if err := validateSchemaElementID(elementID); err != nil {
		return err
	}
	switch elementTail {
	case "":
		a.marshalers[elementID] = &booleanArgumentMarshaler{}
	case "*":
		a.marshalers[elementID] = &stringArgumentMarshaler{}
	case "#":
		a.marshalers[elementID] = &integerArgumentMarshaler{}
	case "##":
		a.marshalers[elementID] = &doubleArgumentMarshaler{}
	case "[*]":
		a.marshalers[elementID] = &stringArrayArgumentMarshaler{}
	case "&":
		a.marshalers[elementID] = &mapArgumentMarshaler{map[string]string{}}
	default:
		return &Error{Code: InvalidArgumentFormat, ArgumentID: elementID, Parameter: elementTail}
	}
	return nil
}

func validateSchemaElementID(elementID rune) error {
	if !unicode.IsLetter(elementID) {
		return &Error{Code: InvalidArgumentName, ArgumentID: elementID}
	}
	return nil
}

func (a *Args) parseArgumentStrings(argsList []string) error {
	a.currentArgument = &argumentIterator{args: argsList}
	for a.currentArgument.hasNext() {
		argString, _ := a.currentArgument.next()
		if strings.HasPrefix(argString, "-") {
			if err := a.parseArgumentCharacters(argString[1:]); err != nil {
				return err
			}
		} else {
			a.currentArgument.previous()
			break
		}
	}
	return nil
}

func (a *Args) parseArgumentCharacters(argChars string) error {
	for _, argChar := range argChars {
		if err := a.parseArgumentCharacter(argChar); err != nil {
			return err
		}
	}
	return nil
}

func (a *Args) parseArgumentCharacter(argChar rune) error {
	m, ok := a.marshalers[argChar]
	if !ok {
		return &Error{Code: UnexpectedArgument, ArgumentID: argChar}
	}
	a.argsFound[argChar] = true
	if err := m.set(a.currentArgument); err != nil {
		err.ArgumentID = argChar
		return err
	}
	return nil
}

// Has reports whether flag arg was given.
func (a *Args) Has(arg rune) bool {
	return a.argsFound[arg]
}

// NextArgument returns the index of the first argument after the flags.
func (a *Args) NextArgument() int {
	return a.currentArgument.nextIndex
}

func (a *Args) GetBoolean(arg rune) bool {
	if m, ok := a.marshalers[arg].(*booleanArgumentMarshaler); ok {
		return m.value
	}
	return false
}

func (a *Args) GetString(arg rune) string {
	if m, ok := a.marshalers[arg].(*stringArgumentMarshaler); ok {
		return m.value
	}
	return ""
}

func (a *Args) GetInt(arg rune) int {
	if m, ok := a.marshalers[arg].(*integerArgumentMarshaler); ok {
		return m.value
	}
	return 0
}

func (a *Args) GetDouble(arg rune) float64 {
	if m, ok := a.marshalers[arg].(*doubleArgumentMarshaler); ok {
		return m.value
	}
	return 0
}

func (a *Args) GetStringArray(arg rune) []string {
	if m, ok := a.marshalers[arg].(*stringArrayArgumentMarshaler); ok {
		return m.values
	}
	return []string{}
}

func (a *Args) GetMap(arg rune) map[string]string {
	if m, ok := a.marshalers[arg].(*mapArgumentMarshaler); ok {
		return m.value
	}
	return map[string]string{}
}

// argumentIterator stands in for Java's ListIterator over the arguments.
type argumentIterator struct {
	args      []string
	nextIndex int
}

func (it *argumentIterator) hasNext() bool {
	return it.nextIndex < len(it.args)
}

func (it *argumentIterator) next() (string, bool) {
	if !it.hasNext() {
		return "", false
	}
	it.nextIndex++
	return it.args[it.nextIndex-1], true
}

func (it *argumentIterator) previous() {
	it.nextIndex--
}
