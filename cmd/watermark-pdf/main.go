// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

package main

import (
	"flag"
	"fmt"
	"os"

	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/watermarkpdf"
)

func main() {
	var (
		email  = flag.String("email", "", "Email address to include in the watermark (required)")
		input  = flag.String("input", "", "Path to the input PDF file (required)")
		output = flag.String("output", "", "Path to the output PDF file (defaults to stdout)")
	)

	flag.Parse()

	if *email == "" || *input == "" {
		fmt.Fprintln(os.Stderr, "Error: -email and -input are required")
		flag.Usage()
		os.Exit(1)
	}

	addr := mail.Addr(*email)

	pdfData, err := os.ReadFile(*input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot read input file: %v\n", err)
		os.Exit(1)
	}

	watermarkedPDF, err := watermarkpdf.AddConfidentialWithTimestamp(pdfData, addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot watermark PDF: %v\n", err)
		os.Exit(1)
	}

	if *output == "" {
		if _, err := os.Stdout.Write(watermarkedPDF); err != nil {
			fmt.Fprintf(os.Stderr, "Error: cannot write to stdout: %v\n", err)
			os.Exit(1)
		}
	} else {
		if err := os.WriteFile(*output, watermarkedPDF, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error: cannot write output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Watermarked PDF written to %s\n", *output)
	}
}
