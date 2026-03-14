// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package iostreams

import (
	"bytes"
	"io"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"golang.org/x/term"
)

type IOStreams struct {
	In     io.ReadCloser
	Out    io.Writer
	ErrOut io.Writer

	// ForceNonInteractive disables all interactive prompts. Set by the
	// --no-interactive global flag or the PROBO_NO_INTERACTIVE env var.
	ForceNonInteractive bool

	// ForceNoColor disables ANSI color output. Set by the --no-color
	// global flag, the NO_COLOR env var, or TERM=dumb.
	ForceNoColor bool
}

func (s *IOStreams) IsInteractive() bool {
	if s.ForceNonInteractive {
		return false
	}
	return s.isStdinTTY() && s.isStdoutTTY()
}

func (s *IOStreams) isStdinTTY() bool {
	if f, ok := s.In.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}

func (s *IOStreams) isStdoutTTY() bool {
	if f, ok := s.Out.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}

func (s *IOStreams) IsStdoutTTY() bool {
	if s.ForceNonInteractive {
		return false
	}
	return s.isStdoutTTY()
}

func (s *IOStreams) ColorEnabled() bool {
	if s.ForceNoColor {
		return false
	}
	return s.isStdoutTTY()
}

// ApplyColorProfile configures the lipgloss default renderer based on
// the current color settings. Call this after ForceNoColor has been set.
func (s *IOStreams) ApplyColorProfile() {
	if s.ForceNoColor {
		lipgloss.SetColorProfile(termenv.Ascii)
	}
}

func System() *IOStreams {
	return &IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}
}

func Test() (*IOStreams, *bytes.Buffer, *bytes.Buffer) {
	out := new(bytes.Buffer)
	errOut := new(bytes.Buffer)
	return &IOStreams{
		In:     io.NopCloser(new(bytes.Buffer)),
		Out:    out,
		ErrOut: errOut,
	}, out, errOut
}
