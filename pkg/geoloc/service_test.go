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

package geoloc_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/geoloc"
)

type locationBlock struct {
	addressFamily   int16
	ipStart         string
	ipEnd           string
	countryCode     coredata.CountryCode
	subdivisionCode *coredata.SubdivisionCode
}

func insertLocationBlocks(t *testing.T, client *pg.Client, blocks ...locationBlock) {
	t.Helper()

	err := client.WithConn(
		context.Background(),
		func(ctx context.Context, conn pg.Querier) error {
			if _, err := conn.Exec(ctx, `TRUNCATE common_ip_location_blocks`); err != nil {
				return err
			}

			for _, block := range blocks {
				if _, err := conn.Exec(
					ctx,
					`
INSERT INTO common_ip_location_blocks (
	address_family,
	ip_start,
	ip_end,
	country_code,
	subdivision_code
) VALUES (
	@address_family,
	@ip_start::inet,
	@ip_end::inet,
	@country_code,
	@subdivision_code
)
`,
					pgx.StrictNamedArgs{
						"address_family":   block.addressFamily,
						"ip_start":         block.ipStart,
						"ip_end":           block.ipEnd,
						"country_code":     block.countryCode.String(),
						"subdivision_code": block.subdivisionCode,
					},
				); err != nil {
					return err
				}
			}

			return nil
		},
	)
	require.NoError(t, err)
}

func TestLookupLocationWideIPv6Block(t *testing.T) {
	client := test.PGClient(t)
	service := geoloc.NewService(client)

	insertLocationBlocks(
		t,
		client,
		locationBlock{
			addressFamily: 6,
			ipStart:       "2001:db8::",
			ipEnd:         "2001:db8:ffff:ffff:ffff:ffff:ffff:ffff",
			countryCode:   coredata.CountryCodeUS,
		},
		locationBlock{
			addressFamily: 6,
			ipStart:       "2001:db8:1::",
			ipEnd:         "2001:db8:1:ffff:ffff:ffff:ffff:ffff",
			countryCode:   coredata.CountryCodeFR,
		},
	)

	location, err := service.LookupLocation(context.Background(), "2001:db8:1::1")
	require.NoError(t, err)
	assert.Equal(t, coredata.CountryCodeFR, location.CountryCode)

	location, err = service.LookupLocation(context.Background(), "2001:db8::1")
	require.NoError(t, err)
	assert.Equal(t, coredata.CountryCodeUS, location.CountryCode)
}

func TestLookupLocationSameStartNestedRanges(t *testing.T) {
	client := test.PGClient(t)
	service := geoloc.NewService(client)
	subdivision := coredata.SubdivisionCode("US-CA")

	insertLocationBlocks(
		t,
		client,
		locationBlock{
			addressFamily: 4,
			ipStart:       "10.0.0.0",
			ipEnd:         "10.0.255.255",
			countryCode:   coredata.CountryCodeUS,
		},
		locationBlock{
			addressFamily:   4,
			ipStart:         "10.0.0.0",
			ipEnd:           "10.0.0.255",
			countryCode:     coredata.CountryCodeUS,
			subdivisionCode: &subdivision,
		},
	)

	location, err := service.LookupLocation(context.Background(), "10.0.0.200")
	require.NoError(t, err)
	assert.Equal(t, coredata.CountryCodeUS, location.CountryCode)
	assert.Equal(t, &subdivision, location.SubdivisionCode)

	location, err = service.LookupLocation(context.Background(), "10.0.1.0")
	require.NoError(t, err)
	assert.Equal(t, coredata.CountryCodeUS, location.CountryCode)
	assert.Nil(t, location.SubdivisionCode)
}

func TestLookupLocationNestedRanges(t *testing.T) {
	client := test.PGClient(t)
	service := geoloc.NewService(client)
	subdivision := coredata.SubdivisionCode("US-CA")

	insertLocationBlocks(
		t,
		client,
		locationBlock{
			addressFamily: 4,
			ipStart:       "10.0.0.0",
			ipEnd:         "10.0.255.255",
			countryCode:   coredata.CountryCodeUS,
		},
		locationBlock{
			addressFamily:   4,
			ipStart:         "10.0.0.128",
			ipEnd:           "10.0.0.255",
			countryCode:     coredata.CountryCodeUS,
			subdivisionCode: &subdivision,
		},
	)

	location, err := service.LookupLocation(context.Background(), "10.0.0.200")
	require.NoError(t, err)
	assert.Equal(t, coredata.CountryCodeUS, location.CountryCode)
	assert.Equal(t, &subdivision, location.SubdivisionCode)

	location, err = service.LookupLocation(context.Background(), "10.0.1.0")
	require.NoError(t, err)
	assert.Equal(t, coredata.CountryCodeUS, location.CountryCode)
	assert.Nil(t, location.SubdivisionCode)
}
