package args

import (
	"strconv"
	"strings"
)

// argumentMarshaler consumes a flag's parameters, if any, from the
// arguments. It leaves the returned Error's ArgumentID for Args to fill in.
type argumentMarshaler interface {
	set(currentArgument *argumentIterator) *Error
}

type booleanArgumentMarshaler struct {
	value bool
}

func (m *booleanArgumentMarshaler) set(currentArgument *argumentIterator) *Error {
	m.value = true
	return nil
}

type stringArgumentMarshaler struct {
	value string
}

func (m *stringArgumentMarshaler) set(currentArgument *argumentIterator) *Error {
	parameter, ok := currentArgument.next()
	if !ok {
		return &Error{Code: MissingString}
	}
	m.value = parameter
	return nil
}

type integerArgumentMarshaler struct {
	value int
}

func (m *integerArgumentMarshaler) set(currentArgument *argumentIterator) *Error {
	parameter, ok := currentArgument.next()
	if !ok {
		return &Error{Code: MissingInteger}
	}
	value, err := strconv.ParseInt(parameter, 10, 32)
	if err != nil {
		return &Error{Code: InvalidInteger, Parameter: parameter}
	}
	m.value = int(value)
	return nil
}

type doubleArgumentMarshaler struct {
	value float64
}

func (m *doubleArgumentMarshaler) set(currentArgument *argumentIterator) *Error {
	parameter, ok := currentArgument.next()
	if !ok {
		return &Error{Code: MissingDouble}
	}
	value, err := strconv.ParseFloat(strings.TrimSpace(parameter), 64)
	if err != nil {
		return &Error{Code: InvalidDouble, Parameter: parameter}
	}
	m.value = value
	return nil
}

type stringArrayArgumentMarshaler struct {
	values []string
}

func (m *stringArrayArgumentMarshaler) set(currentArgument *argumentIterator) *Error {
	parameter, ok := currentArgument.next()
	if !ok {
		return &Error{Code: MissingString}
	}
	m.values = append(m.values, parameter)
	return nil
}

type mapArgumentMarshaler struct {
	value map[string]string
}

func (m *mapArgumentMarshaler) set(currentArgument *argumentIterator) *Error {
	parameter, ok := currentArgument.next()
	if !ok {
		return &Error{Code: MissingMap}
	}
	for _, entry := range javaSplit(parameter, ",") {
		entryComponents := javaSplit(entry, ":")
		if len(entryComponents) != 2 {
			return &Error{Code: MalformedMap}
		}
		m.value[entryComponents[0]] = entryComponents[1]
	}
	return nil
}

// javaSplit splits s like Java's String.split with a literal separator,
// which drops trailing empty strings.
func javaSplit(s, sep string) []string {
	parts := strings.Split(s, sep)
	for len(parts) > 1 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}
