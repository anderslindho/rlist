package style

import (
	"io"
	"os"

	"golang.org/x/term"
)

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	fgRed    = "\033[31m"
	fgGreen  = "\033[32m"
	fgYellow = "\033[33m"
	fgCyan   = "\033[36m"
)

// Styler adds ANSI attributes when Out is a terminal and Enabled is true.
type Styler struct {
	Enabled bool
	Out     io.Writer
}

func NewStdout(enabled bool) *Styler {
	on := enabled && os.Getenv("NO_COLOR") == "" && term.IsTerminal(int(os.Stdout.Fd()))
	return &Styler{Enabled: on, Out: os.Stdout}
}

func (s *Styler) wrap(code, t string) string {
	if !s.Enabled {
		return t
	}
	return code + t + reset
}

func (s *Styler) Bold(t string) string   { return s.wrap(bold, t) }
func (s *Styler) Dim(t string) string    { return s.wrap(dim, t) }
func (s *Styler) Green(t string) string  { return s.wrap(fgGreen, t) }
func (s *Styler) Yellow(t string) string { return s.wrap(fgYellow, t) }
func (s *Styler) Cyan(t string) string   { return s.wrap(fgCyan, t) }
func (s *Styler) Red(t string) string    { return s.wrap(fgRed, t) }

func (s *Styler) Vis(v string) string {
	if !s.Enabled {
		return v
	}
	switch v {
	case "public":
		return s.Green(v)
	case "private":
		return s.Yellow(v)
	case "internal":
		return s.Cyan(v)
	default:
		return v
	}
}
