package terminal

import "testing"

func TestTextEscapesControlCharacters(t *testing.T) {
	if got, want := Text("text\x1b[31m\r\a\u0085"), `text\x1b[31m\x0d\x07\x85`; got != want {
		t.Errorf("Text() = %q, want %q", got, want)
	}
}

func TestLinesPreservesLineBreaks(t *testing.T) {
	if got, want := Lines("first\x1b\nsecond\r"), "first\\x1b\nsecond\\x0d"; got != want {
		t.Errorf("Lines() = %q, want %q", got, want)
	}
}
