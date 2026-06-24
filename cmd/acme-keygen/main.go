// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	"context"
	"flag"
	"fmt"
	"os"

	"go.probo.inc/probo/pkg/crypto/keys"
	"go.probo.inc/probo/pkg/crypto/pem"
	"golang.org/x/crypto/acme"
)

func main() {
	var (
		email     = flag.String("email", "", "Email address for ACME account (required)")
		keyType   = flag.String("key-type", "EC256", "Key type: EC256, EC384, RSA2048, RSA4096")
		directory = flag.String("directory", "https://acme-v02.api.letsencrypt.org/directory", "ACME directory URL")
	)

	flag.Parse()

	if *email == "" {
		fmt.Fprintln(os.Stderr, "Error: -email is required")
		flag.Usage()
		os.Exit(1)
	}

	accountKey, err := keys.Generate(keys.Type(*keyType))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating key: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Generated %s account key\n", *keyType)

	client := &acme.Client{
		Key:          accountKey,
		DirectoryURL: *directory,
	}

	ctx := context.Background()
	account := &acme.Account{
		Contact: []string{"mailto:" + *email},
	}

	registeredAccount, err := client.Register(ctx, account, acme.AcceptTOS)
	if err != nil {
		if err == acme.ErrAccountAlreadyExists {
			fmt.Fprintf(os.Stderr, "Account already exists for this key\n")
		} else {
			fmt.Fprintf(os.Stderr, "Error registering account: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Fprintf(os.Stderr, "Successfully registered ACME account\n")
		fmt.Fprintf(os.Stderr, "Account URI: %s\n", registeredAccount.URI)
	}

	keyPEM, err := pem.EncodePrivateKey(accountKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding key: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "\n=== ACME Account Private Key (PEM) ===\n")
	fmt.Fprintf(os.Stderr, "Add this to your configuration under custom-domains.acme.account-key:\n\n")
	fmt.Print(string(keyPEM))
}
