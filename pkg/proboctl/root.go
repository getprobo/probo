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

package proboctl

import (
	"fmt"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/proboctl/cookiebanner"
)

func NewRootCmd(version string) *cobra.Command {
	var pgDSN string

	cmd := &cobra.Command{
		Use:           "proboctl <command> [flags]",
		Short:         "Probo internal admin CLI",
		Long:          "proboctl is an internal tool for Probo staff to manage parameters not exposed through the public API.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().StringVar(&pgDSN, "pg-dsn", "", "PostgreSQL connection string (also PG_DSN env var)")

	newPG := func() (*pg.Client, error) {
		return newPGClient(pgDSN)
	}

	cmd.AddCommand(cookiebanner.NewCmdCookieBanner(newPG))

	return cmd
}

func newPGClient(flagDSN string) (*pg.Client, error) {
	dsn := flagDSN
	if dsn == "" {
		dsn = os.Getenv("PG_DSN")
	}

	var addr, user, pass, db string

	if dsn != "" {
		cfg, err := pgx.ParseConfig(dsn)
		if err != nil {
			return nil, fmt.Errorf("invalid DSN: %w", err)
		}
		addr = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
		user = cfg.User
		pass = cfg.Password
		db = cfg.Database
	} else {
		addr = envOrDefault("PG_ADDR", "localhost:5432")
		user = envOrDefault("PG_USERNAME", "postgres")
		pass = envOrDefault("PG_PASSWORD", "postgres")
		db = envOrDefault("PG_DATABASE", "probod")
	}

	poolSize := envIntOrDefault("PG_POOL_SIZE", 2)

	opts := []pg.Option{
		pg.WithAddr(addr),
		pg.WithUser(user),
		pg.WithPassword(pass),
		pg.WithDatabase(db),
		pg.WithPoolSize(int32(poolSize)),
	}

	return pg.NewClient(opts...)
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envIntOrDefault(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
		fmt.Fprintf(os.Stderr, "warning: invalid %s value %q, using default %d\n", key, v, fallback)
	}
	return fallback
}
