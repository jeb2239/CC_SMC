// Package utilities is a Go port of smc.Utilities.
package utilities

import (
	"regexp"
	"strings"
)

// CommaList joins names with commas.
func CommaList(names []string) string {
	return strings.Join(names, ",")
}

// IotaList formats names as the body of a Go const block whose first
// element is typeName = iota.
func IotaList(typeName string, names []string) string {
	var sb strings.Builder
	for i, name := range names {
		sb.WriteString("\t" + name)
		if i == 0 {
			sb.WriteString(" " + typeName + " = iota")
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// Capitalize upper-cases the first character of s.
func Capitalize(s string) string {
	if len(s) < 2 {
		return strings.ToUpper(s)
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// AddPrefix returns a copy of list with prefix prepended to every element.
func AddPrefix(prefix string, list []string) []string {
	result := make([]string, 0, len(list))
	for _, element := range list {
		result = append(result, prefix+element)
	}
	return result
}

var (
	newlines         = regexp.MustCompile(`\n+`)
	blanks           = regexp.MustCompile(`[\t ]+`)
	spacesAroundLine = regexp.MustCompile(` *\n *`)
)

// CompressWhiteSpace collapses runs of newlines, and of tabs and spaces, and
// strips spaces around line breaks.
func CompressWhiteSpace(s string) string {
	s = newlines.ReplaceAllString(s, "\n")
	s = blanks.ReplaceAllString(s, " ")
	return spacesAroundLine.ReplaceAllString(s, "\n")
}
