// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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

func (s *IOStreams) IsStdinTTY() bool {
	return s.isStdinTTY()
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
