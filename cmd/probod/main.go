// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	"fmt"
	"os"
	"time"

	"go.gearno.de/kit/unit"
	"go.probo.inc/probo/pkg/probod"
)

var (
	version string = "unknown"
	env     string = "unknown"
)

func main() {
	impl := probod.New()
	unit := unit.NewUnit(impl, "probod", version, env)

	err := unit.Run()
	if err != nil && err != context.Canceled {
		fmt.Fprintf(
			os.Stderr,
			`{"time": %q, "msg": %q, "version": %q, "environment": %q, "level": "ERROR", "name": "probod", "error": %q}\n`,
			time.Now().Format(time.RFC3339),
			"cannot run probod",
			version,
			env,
			err.Error(),
		)
		os.Exit(1)
	}

	os.Exit(0)
}
