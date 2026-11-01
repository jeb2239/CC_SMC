package args

import "fmt"

// ErrorCode classifies an Error.
type ErrorCode int

const (
	OK ErrorCode = iota
	InvalidArgumentFormat
	UnexpectedArgument
	InvalidArgumentName
	MissingString
	MissingInteger
	InvalidInteger
	MissingDouble
	InvalidDouble
	MalformedMap
	MissingMap
)

// Error describes a bad schema or command line.
type Error struct {
	Code       ErrorCode
	ArgumentID rune
	Parameter  string
}

func (e *Error) Error() string { return e.ErrorMessage() }

// ErrorMessage describes the error for a user.
func (e *Error) ErrorMessage() string {
	switch e.Code {
	case OK:
		return "TILT: Should not get here."
	case UnexpectedArgument:
		return fmt.Sprintf("Argument -%c unexpected.", e.ArgumentID)
	case MissingString:
		return fmt.Sprintf("Could not find string parameter for -%c.", e.ArgumentID)
	case InvalidInteger:
		return fmt.Sprintf("Argument -%c expects an integer but was '%s'.", e.ArgumentID, e.Parameter)
	case MissingInteger:
		return fmt.Sprintf("Could not find integer parameter for -%c.", e.ArgumentID)
	case InvalidDouble:
		return fmt.Sprintf("Argument -%c expects a double but was '%s'.", e.ArgumentID, e.Parameter)
	case MissingDouble:
		return fmt.Sprintf("Could not find double parameter for -%c.", e.ArgumentID)
	case InvalidArgumentName:
		return fmt.Sprintf("'%c' is not a valid argument name.", e.ArgumentID)
	case InvalidArgumentFormat:
		return fmt.Sprintf("'%s' is not a valid argument format.", e.Parameter)
	// The Java version has no message for these two.
	case MalformedMap:
		return fmt.Sprintf("Argument -%c expects key:value pairs separated by commas.", e.ArgumentID)
	case MissingMap:
		return fmt.Sprintf("Could not find map parameter for -%c.", e.ArgumentID)
	}
	return ""
}
