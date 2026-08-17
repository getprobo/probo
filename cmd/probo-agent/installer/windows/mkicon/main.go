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

package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"

	"github.com/tc-hib/winres"
	"go.probo.inc/probo/pkg/deviceagent/tray"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("mkicon", flag.ContinueOnError)
	pngPath := fs.String("png", "", "path to icon_color.png")
	icoPath := fs.String("ico", "", "write a multi-size ICO to this path")
	sysoPath := fs.String("syso", "", "write a Windows COFF .syso to this path")
	arch := fs.String("arch", "", "GOARCH for the .syso (amd64 or arm64)")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("cannot parse flags: %w", err)
	}

	if *pngPath == "" {
		return fmt.Errorf("cannot run mkicon: -png is required")
	}

	if *icoPath == "" && *sysoPath == "" {
		return fmt.Errorf("cannot run mkicon: need -ico and/or -syso")
	}

	if *icoPath != "" {
		if err := writeProductICO(*pngPath, *icoPath); err != nil {
			return fmt.Errorf("cannot write product ico: %w", err)
		}
	}

	if *sysoPath != "" {
		if *arch == "" {
			return fmt.Errorf("cannot write syso: -arch is required")
		}

		if err := writeProductSyso(*pngPath, *sysoPath, *arch); err != nil {
			return fmt.Errorf("cannot write product syso: %w", err)
		}
	}

	return nil
}

func writeProductICO(pngPath, icoPath string) error {
	ico, err := productICO(pngPath)
	if err != nil {
		return fmt.Errorf("cannot build product ico: %w", err)
	}

	if err := os.WriteFile(icoPath, ico, 0o644); err != nil {
		return fmt.Errorf("cannot write ico: %w", err)
	}

	return nil
}

func writeProductSyso(pngPath, sysoPath, arch string) error {
	ico, err := productICO(pngPath)
	if err != nil {
		return fmt.Errorf("cannot build product ico: %w", err)
	}

	icon, err := winres.LoadICO(bytes.NewReader(ico))
	if err != nil {
		return fmt.Errorf("cannot parse ico: %w", err)
	}

	target, err := winresArch(arch)
	if err != nil {
		return fmt.Errorf("cannot select syso arch: %w", err)
	}

	var rs winres.ResourceSet
	if err := rs.SetIcon(winres.ID(1), icon); err != nil {
		return fmt.Errorf("cannot set icon resource: %w", err)
	}

	out, err := os.Create(sysoPath)
	if err != nil {
		return fmt.Errorf("cannot create syso: %w", err)
	}

	if err := rs.WriteObject(out, target); err != nil {
		_ = out.Close()
		return fmt.Errorf("cannot write syso: %w", err)
	}

	if err := out.Close(); err != nil {
		return fmt.Errorf("cannot close syso: %w", err)
	}

	return nil
}

func productICO(pngPath string) ([]byte, error) {
	pngBytes, err := os.ReadFile(pngPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read png: %w", err)
	}

	ico, err := tray.PNGToMultiSizeICO(pngBytes, 16, 32, 48, 256)
	if err != nil {
		return nil, fmt.Errorf("cannot encode ico: %w", err)
	}

	return ico, nil
}

func winresArch(arch string) (winres.Arch, error) {
	switch arch {
	case "amd64":
		return winres.ArchAMD64, nil
	case "arm64":
		return winres.ArchARM64, nil
	default:
		return "", fmt.Errorf("cannot map arch %q: want amd64 or arm64", arch)
	}
}
