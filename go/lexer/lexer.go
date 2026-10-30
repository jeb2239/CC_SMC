// Package lexer is a Go port of smc.lexer.Lexer.
package lexer

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// Lexer splits SMC source into tokens and reports them to a TokenCollector.
type Lexer struct {
	collector  TokenCollector
	lineNumber int
	position   int // byte offset into the current line
	column     int // character offset into the current line, as reported to the collector
}

// New returns a Lexer that reports tokens to collector.
func New(collector TokenCollector) *Lexer {
	return &Lexer{collector: collector}
}

// Lex tokenizes s.
func (l *Lexer) Lex(s string) {
	l.lineNumber = 1
	for line := range strings.SplitSeq(s, "\n") {
		l.lexLine(line)
		l.lineNumber++
	}
}

func (l *Lexer) lexLine(line string) {
	l.position, l.column = 0, 0
	for l.position < len(line) {
		l.lexToken(line)
	}
}

func (l *Lexer) lexToken(line string) {
	if !l.findToken(line) {
		l.collector.Error(l.lineNumber, l.column+1)
		_, size := utf8.DecodeRuneInString(line[l.position:])
		l.advance(size, 1)
	}
}

func (l *Lexer) findToken(line string) bool {
	return l.findWhiteSpace(line) ||
		l.findSingleCharacterToken(line) ||
		l.findName(line)
}

// Java's \s also matches \x0B; Go's does not, so spell the class out.
var whitePattern = regexp.MustCompile(`^[ \t\n\x0B\f\r]+`)
var commentPattern = regexp.MustCompile(`^//.*`)
var whitePatterns = []*regexp.Regexp{whitePattern, commentPattern}

func (l *Lexer) findWhiteSpace(line string) bool {
	for _, pattern := range whitePatterns {
		if match := pattern.FindString(line[l.position:]); match != "" {
			l.advance(len(match), utf8.RuneCountInString(match))
			return true
		}
	}
	return false
}

func (l *Lexer) findSingleCharacterToken(line string) bool {
	switch line[l.position] {
	case '{':
		l.collector.OpenBrace(l.lineNumber, l.column)
	case '}':
		l.collector.ClosedBrace(l.lineNumber, l.column)
	case '(':
		l.collector.OpenParen(l.lineNumber, l.column)
	case ')':
		l.collector.ClosedParen(l.lineNumber, l.column)
	case '<':
		l.collector.OpenAngle(l.lineNumber, l.column)
	case '>':
		l.collector.ClosedAngle(l.lineNumber, l.column)
	case '-':
		l.collector.Dash(l.lineNumber, l.column)
	case '*':
		l.collector.Dash(l.lineNumber, l.column)
	case ':':
		l.collector.Colon(l.lineNumber, l.column)
	default:
		return false
	}
	l.advance(1, 1)
	return true
}

var namePattern = regexp.MustCompile(`^\w+`)

func (l *Lexer) findName(line string) bool {
	if match := namePattern.FindString(line[l.position:]); match != "" {
		l.collector.Name(match, l.lineNumber, l.column)
		l.advance(len(match), len(match))
		return true
	}
	return false
}

func (l *Lexer) advance(bytes, chars int) {
	l.position += bytes
	l.column += chars
}
