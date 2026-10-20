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

package cmdutil

import (
	"os"
	"strings"

	"github.com/charmbracelet/huh"
)

// FillSecret trims a flag value, then an environment variable, then prompts
// when interactive. The stored secret is kept verbatim, so a trailing newline
// would authenticate as a different secret.
func FillSecret(interactive bool, value *string, envName, title string) error {
	*value = strings.TrimSpace(*value)
	if *value == "" {
		*value = strings.TrimSpace(os.Getenv(envName))
	}

	if *value == "" && interactive {
		err := huh.NewInput().
			Title(title).
			EchoMode(huh.EchoModePassword).
			Value(value).
			Run()
		if err != nil {
			return err
		}

		*value = strings.TrimSpace(*value)
	}

	return nil
}
