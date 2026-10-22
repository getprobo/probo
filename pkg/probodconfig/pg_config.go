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

package probodconfig

import (
	"crypto/x509"
	"encoding/pem"
	"time"

	"go.gearno.de/kit/pg"
)

type PgConfig struct {
	Addr                         string `json:"addr,omitempty"`
	Username                     string `json:"username,omitempty"`
	Password                     string `json:"password,omitempty"`
	Database                     string `json:"database,omitempty"`
	PoolSize                     int32  `json:"pool-size"`
	MinPoolSize                  int32  `json:"min-pool-size"`
	MaxConnIdleTimeSeconds       int    `json:"max-conn-idle-time-seconds"`
	MaxConnLifetimeSeconds       int    `json:"max-conn-lifetime-seconds"`
	MaxConnLifetimeJitterSeconds int    `json:"max-conn-lifetime-jitter-seconds"`
	HealthCheckPeriodSeconds     int    `json:"health-check-period-seconds"`
	CACertBundle                 string `json:"ca-cert-bundle,omitempty"`
	Debug                        bool   `json:"debug"`
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

	if cfg.MaxConnLifetimeJitterSeconds > 0 {
		opts = append(
			opts,
			pg.WithMaxConnLifetimeJitter(
				time.Duration(cfg.MaxConnLifetimeJitterSeconds)*time.Second,
			),
		)
	}

	if cfg.HealthCheckPeriodSeconds > 0 {
		opts = append(
			opts,
			pg.WithHealthCheckPeriod(
				time.Duration(cfg.HealthCheckPeriodSeconds)*time.Second,
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
