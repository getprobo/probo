// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package deviceagent

import (
	"context"
	"time"
)

func platformString() string {
	return "WINDOWS"
}

func collectOSVersion() string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out, _ := runQuiet(ctx, "cmd", "/C", "ver")
	if out != "" {
		return out
	}

	out, _ = runQuiet(ctx, "uname", "-sr")

	return out
}

func collectHardwareUUID() string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// `wmic` is deprecated; `Get-CimInstance` requires PowerShell.
	out, _ := runQuiet(
		ctx,
		"powershell",
		"-NoProfile",
		"-Command",
		"(Get-CimInstance Win32_ComputerSystemProduct).UUID",
	)
	if out != "" {
		return out
	}

	return hashFallbackUUID()
}

func collectSerialNumber() string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out, _ := runQuiet(
		ctx,
		"powershell",
		"-NoProfile",
		"-Command",
		"(Get-CimInstance Win32_BIOS).SerialNumber",
	)

	return out
}
