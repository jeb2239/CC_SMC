package args

import (
	"errors"
	"math"
	"slices"
	"testing"
)

func mustParse(t *testing.T, schema string, arguments ...string) *Args {
	t.Helper()
	a, err := New(schema, arguments)
	if err != nil {
		t.Fatalf("New(%q, %q): %v", schema, arguments, err)
	}
	return a
}

func assertError(t *testing.T, schema string, arguments []string, code ErrorCode, argumentID rune, parameter string) {
	t.Helper()
	_, err := New(schema, arguments)
	var argsErr *Error
	if !errors.As(err, &argsErr) {
		t.Fatalf("New(%q, %q) error = %v, want an *Error", schema, arguments, err)
	}
	want := Error{code, argumentID, parameter}
	if *argsErr != want {
		t.Errorf("New(%q, %q) error = %+v, want %+v", schema, arguments, *argsErr, want)
	}
}

func assertNextArgument(t *testing.T, a *Args, want int) {
	t.Helper()
	if got := a.NextArgument(); got != want {
		t.Errorf("NextArgument() = %d, want %d", got, want)
	}
}

func TestCreateWithNoSchemaOrArguments(t *testing.T) {
	assertNextArgument(t, mustParse(t, ""), 0)
}

func TestErrors(t *testing.T) {
	tests := []struct {
		name       string
		schema     string
		arguments  []string
		code       ErrorCode
		argumentID rune
		parameter  string
	}{
		{"WithNoSchemaButWithOneArgument", "", []string{"-x"}, UnexpectedArgument, 'x', ""},
		{"WithNoSchemaButWithMultipleArguments", "", []string{"-x", "-y"}, UnexpectedArgument, 'x', ""},
		{"NonLetterSchema", "*", nil, InvalidArgumentName, '*', ""},
		{"InvalidArgumentFormat", "f~", nil, InvalidArgumentFormat, 'f', "~"},
		{"MissingStringArgument", "x*", []string{"-x"}, MissingString, 'x', ""},
		{"InvalidInteger", "x#", []string{"-x", "Forty two"}, InvalidInteger, 'x', "Forty two"},
		{"MissingInteger", "x#", []string{"-x"}, MissingInteger, 'x', ""},
		{"InvalidDouble", "x##", []string{"-x", "Forty two"}, InvalidDouble, 'x', "Forty two"},
		{"MissingDouble", "x##", []string{"-x"}, MissingDouble, 'x', ""},
		{"MissingStringArrayElement", "x[*]", []string{"-x"}, MissingString, 'x', ""},
		{"malFormedMapArgument", "f&", []string{"-f", "key1:val1,key2"}, MalformedMap, 'f', ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertError(t, tt.schema, tt.arguments, tt.code, tt.argumentID, tt.parameter)
		})
	}
}

func TestSimpleBooleanPresent(t *testing.T) {
	a := mustParse(t, "x", "-x")
	if !a.GetBoolean('x') {
		t.Error("GetBoolean('x') = false")
	}
	assertNextArgument(t, a, 1)
}

func TestSimpleStringPresent(t *testing.T) {
	a := mustParse(t, "x*", "-x", "param")
	if !a.Has('x') || a.GetString('x') != "param" {
		t.Errorf("Has('x') = %v, GetString('x') = %q", a.Has('x'), a.GetString('x'))
	}
	assertNextArgument(t, a, 2)
}

func TestSpacesInFormat(t *testing.T) {
	a := mustParse(t, "x, y", "-xy")
	if !a.Has('x') || !a.Has('y') {
		t.Errorf("Has('x') = %v, Has('y') = %v", a.Has('x'), a.Has('y'))
	}
	assertNextArgument(t, a, 1)
}

func TestSimpleIntPresent(t *testing.T) {
	a := mustParse(t, "x#", "-x", "42")
	if !a.Has('x') || a.GetInt('x') != 42 {
		t.Errorf("Has('x') = %v, GetInt('x') = %d", a.Has('x'), a.GetInt('x'))
	}
	assertNextArgument(t, a, 2)
}

func TestSimpleDoublePresent(t *testing.T) {
	a := mustParse(t, "x##", "-x", "42.3")
	if !a.Has('x') || math.Abs(a.GetDouble('x')-42.3) > .001 {
		t.Errorf("Has('x') = %v, GetDouble('x') = %v", a.Has('x'), a.GetDouble('x'))
	}
}

func TestStringArray(t *testing.T) {
	a := mustParse(t, "x[*]", "-x", "alpha")
	if got := a.GetStringArray('x'); !a.Has('x') || !slices.Equal(got, []string{"alpha"}) {
		t.Errorf("Has('x') = %v, GetStringArray('x') = %q", a.Has('x'), got)
	}
}

func TestManyStringArrayElements(t *testing.T) {
	a := mustParse(t, "x[*]", "-x", "alpha", "-x", "beta", "-x", "gamma")
	if got := a.GetStringArray('x'); !a.Has('x') || !slices.Equal(got, []string{"alpha", "beta", "gamma"}) {
		t.Errorf("Has('x') = %v, GetStringArray('x') = %q", a.Has('x'), got)
	}
}

func TestMapArgument(t *testing.T) {
	a := mustParse(t, "f&", "-f", "key1:val1,key2:val2")
	m := a.GetMap('f')
	if !a.Has('f') || m["key1"] != "val1" || m["key2"] != "val2" {
		t.Errorf("Has('f') = %v, GetMap('f') = %v", a.Has('f'), m)
	}
}

func TestOneMapArgument(t *testing.T) {
	a := mustParse(t, "f&", "-f", "key1:val1")
	if m := a.GetMap('f'); !a.Has('f') || m["key1"] != "val1" {
		t.Errorf("Has('f') = %v, GetMap('f') = %v", a.Has('f'), m)
	}
}

func TestExtraArguments(t *testing.T) {
	a := mustParse(t, "x,y*", "-x", "-y", "alpha", "beta")
	if !a.GetBoolean('x') || a.GetString('y') != "alpha" {
		t.Errorf("GetBoolean('x') = %v, GetString('y') = %q", a.GetBoolean('x'), a.GetString('y'))
	}
	assertNextArgument(t, a, 3)
}

func TestExtraArgumentsThatLookLikeFlags(t *testing.T) {
	a := mustParse(t, "x,y", "-x", "alpha", "-y", "beta")
	if !a.Has('x') || a.Has('y') || !a.GetBoolean('x') || a.GetBoolean('y') {
		t.Errorf("Has: x=%v y=%v; GetBoolean: x=%v y=%v", a.Has('x'), a.Has('y'), a.GetBoolean('x'), a.GetBoolean('y'))
	}
	assertNextArgument(t, a, 1)
}

func TestErrorMessages(t *testing.T) {
	tests := []struct {
		name     string
		err      Error
		expected string
	}{
		{"UnexpectedMessage", Error{UnexpectedArgument, 'x', ""}, "Argument -x unexpected."},
		{"MissingStringMessage", Error{MissingString, 'x', ""}, "Could not find string parameter for -x."},
		{"InvalidIntegerMessage", Error{InvalidInteger, 'x', "Forty two"}, "Argument -x expects an integer but was 'Forty two'."},
		{"MissingIntegerMessage", Error{MissingInteger, 'x', ""}, "Could not find integer parameter for -x."},
		{"InvalidDoubleMessage", Error{InvalidDouble, 'x', "Forty two"}, "Argument -x expects a double but was 'Forty two'."},
		{"MissingDoubleMessage", Error{MissingDouble, 'x', ""}, "Could not find double parameter for -x."},
		{"InvalidArgumentName", Error{InvalidArgumentName, '#', ""}, "'#' is not a valid argument name."},
		{"InvalidFormat", Error{InvalidArgumentFormat, 'x', "$"}, "'$' is not a valid argument format."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.ErrorMessage(); got != tt.expected {
				t.Errorf("ErrorMessage() = %q, want %q", got, tt.expected)
			}
		})
	}
}
