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
