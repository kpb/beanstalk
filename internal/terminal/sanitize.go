// Package terminal provides safe display formatting for terminal output.
package terminal

import (
	"fmt"
	"strings"
	"unicode"
)

// Text escapes control characters so task content cannot alter terminal state.
func Text(value string) string {
	var output strings.Builder
	for _, character := range value {
		if !unicode.IsControl(character) {
			output.WriteRune(character)
			continue
		}
		if character <= 0xff {
			fmt.Fprintf(&output, `\x%02x`, character)
		} else {
			fmt.Fprintf(&output, `\u%04x`, character)
		}
	}
	return output.String()
}

// Lines escapes controls in each line while preserving body line breaks.
func Lines(value string) string {
	lines := strings.Split(value, "\n")
	for index, line := range lines {
		lines[index] = Text(line)
	}
	return strings.Join(lines, "\n")
}
