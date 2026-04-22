// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

package probod

import (
	"crypto/x509"
	"encoding/pem"
	"time"

	"go.gearno.de/kit/pg"
)

type PgConfig struct {
	Addr                   string `json:"addr"`
	Username               string `json:"username"`
	Password               string `json:"password"`
	Database               string `json:"database"`
	PoolSize               int32  `json:"pool-size"`
	MinPoolSize            int32  `json:"min-pool-size"`
	MaxConnIdleTimeSeconds int    `json:"max-conn-idle-time-seconds"`
	MaxConnLifetimeSeconds int    `json:"max-conn-lifetime-seconds"`
	CACertBundle           string `json:"ca-cert-bundle"`
	Debug                  bool   `json:"debug"`
}

func (cfg PgConfig) Options(options ...pg.Option) []pg.Option {
	opts := []pg.Option{
		pg.WithAddr(cfg.Addr),
		pg.WithUser(cfg.Username),
		pg.WithPassword(cfg.Password),
		pg.WithDatabase(cfg.Database),
		pg.WithPoolSize(cfg.PoolSize),
	}

	if cfg.MinPoolSize > 0 {
		opts = append(opts, pg.WithMinPoolSize(cfg.MinPoolSize))
	}

	if cfg.MaxConnIdleTimeSeconds > 0 {
		opts = append(
			opts,
			pg.WithMaxConnIdleTime(
				time.Duration(cfg.MaxConnIdleTimeSeconds)*time.Second,
			),
		)
	}

	if cfg.MaxConnLifetimeSeconds > 0 {
		opts = append(
			opts,
			pg.WithMaxConnLifetime(
				time.Duration(cfg.MaxConnLifetimeSeconds)*time.Second,
			),
		)
	}

	if cfg.Debug {
		opts = append(opts, pg.WithDebug())
	}

	if cfg.CACertBundle != "" {
		var certs []*x509.Certificate
		pemData := []byte(cfg.CACertBundle)

		for len(pemData) > 0 {
			var block *pem.Block
			block, pemData = pem.Decode(pemData)
			if block == nil {
				break
			}
			if block.Type != "CERTIFICATE" {
				continue
			}

			cert, err := x509.ParseCertificate(block.Bytes)
			if err == nil {
				certs = append(certs, cert)
			}
		}

		if len(certs) > 0 {
			opts = append(opts, pg.WithTLS(certs))
		}
	}

	opts = append(opts, options...)

	return opts
}
