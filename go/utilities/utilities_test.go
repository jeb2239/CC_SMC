package utilities

import "testing"

func TestCompressWhiteSpace(t *testing.T) {
	tests := []struct{ name, input, expected string }{
		{"emptyString", "", ""},
		{"NoWhiteSpace", "stringwithnowhitespace", "stringwithnowhitespace"},
		{"oneSpace", "one space", "one space"},
		{"manyWordsWithSingleSpaces", "many words with single spaces", "many words with single spaces"},
		{"oneTab", "one\ttab", "one tab"},
		{"twoTabs", "two\t\ttabs", "two tabs"},
		{"oneReturn", "one\nreturn", "one\nreturn"},
		{"returnsAndSpaces", "word \n word", "word\nword"},
		{"startingWhitespace", "\n  this", "\nthis"},
		{"acceptanceTest",
			"this  is\n\na\t\t  string     \n     \twith\n\nmany\n\n\n\t  whitespaces",
			"this is\na string\nwith\nmany\nwhitespaces"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CompressWhiteSpace(tt.input); got != tt.expected {
				t.Errorf("CompressWhiteSpace(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
