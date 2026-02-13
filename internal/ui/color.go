package ui

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

func colorEnabled() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return term.IsTerminal(int(os.Stderr.Fd()))
}

func wrap(code, s string) string {
	if !colorEnabled() {
		return s
	}
	return fmt.Sprintf("\033[%sm%s\033[0m", code, s)
}

// Red returns s wrapped in red ANSI color.
func Red(s string) string { return wrap("31", s) }

// Yellow returns s wrapped in yellow ANSI color.
func Yellow(s string) string { return wrap("33", s) }

// Green returns s wrapped in green ANSI color.
func Green(s string) string { return wrap("32", s) }

// Bold returns s wrapped in bold ANSI style.
func Bold(s string) string { return wrap("1", s) }
