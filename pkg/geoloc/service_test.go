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
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/geoloc"
)

type blockSource struct {
	blocks []coredata.IPLocationBlock
	index  int
}

func (s *blockSource) Next() bool {
	if s.index >= len(s.blocks) {
		return false
	}

	s.index++

	return true
}

func (s *blockSource) LocationBlock() (coredata.IPLocationBlock, error) {
	return s.blocks[s.index-1], nil
}

func (s *blockSource) Err() error {
	return nil
}

func TestReplaceLocationsAndLookupBoundaries(t *testing.T) {
	client := test.PGClient(t)
	service := geoloc.NewService(client)
	subdivision := coredata.SubdivisionCode("US-CA")

	count, err := service.ReplaceLocations(
		context.Background(),
		&blockSource{
			blocks: []coredata.IPLocationBlock{
				{
					AddressFamily:   4,
					IPStart:         netip.MustParseAddr("10.0.0.0"),
					IPEnd:           netip.MustParseAddr("10.0.0.255"),
					CountryCode:     coredata.CountryCodeUS,
					SubdivisionCode: &subdivision,
				},
				{
					AddressFamily: 4,
					IPStart:       netip.MustParseAddr("10.0.2.0"),
					IPEnd:         netip.MustParseAddr("10.0.2.255"),
					CountryCode:   coredata.CountryCodeFR,
				},
				{
					AddressFamily: 6,
					IPStart:       netip.MustParseAddr("2001:db8::"),
					IPEnd:         netip.MustParseAddr("2001:db8::ffff"),
					CountryCode:   coredata.CountryCodeGB,
				},
			},
		},
	)
	require.NoError(t, err)
	require.Equal(t, int64(3), count)

	for _, ip := range []string{"10.0.0.0", "10.0.0.255"} {
		location, err := service.LookupLocation(context.Background(), ip)
		require.NoError(t, err)
		assert.Equal(t, coredata.CountryCodeUS, location.CountryCode)
		assert.Equal(t, &subdivision, location.SubdivisionCode)
	}

	location, err := service.LookupLocation(context.Background(), "10.0.1.0")
	require.NoError(t, err)
	assert.Equal(t, coredata.IPLocation{}, location)

	for _, ip := range []string{"2001:db8::", "2001:db8::ffff"} {
		location, err := service.LookupLocation(context.Background(), ip)
		require.NoError(t, err)
		assert.Equal(t, coredata.CountryCodeGB, location.CountryCode)
		assert.Nil(t, location.SubdivisionCode)
	}

	_, err = service.ReplaceLocations(
		context.Background(),
		&blockSource{
			blocks: []coredata.IPLocationBlock{
				{
					AddressFamily: 4,
					IPStart:       netip.MustParseAddr("192.0.2.0"),
					IPEnd:         netip.MustParseAddr("192.0.2.127"),
					CountryCode:   coredata.CountryCodeUS,
				},
				{
					AddressFamily: 4,
					IPStart:       netip.MustParseAddr("192.0.2.0"),
					IPEnd:         netip.MustParseAddr("192.0.2.127"),
					CountryCode:   coredata.CountryCodeFR,
				},
			},
		},
	)
	require.Error(t, err)

	location, err = service.LookupLocation(context.Background(), "10.0.0.0")
	require.NoError(t, err)
	assert.Equal(t, coredata.CountryCodeUS, location.CountryCode)
	assert.Equal(t, &subdivision, location.SubdivisionCode)
}

func TestLookupLocationWideIPv6Block(t *testing.T) {
	client := test.PGClient(t)
	service := geoloc.NewService(client)

	_, err := service.ReplaceLocations(
		context.Background(),
		&blockSource{
			blocks: []coredata.IPLocationBlock{
				{
					AddressFamily: 6,
					IPStart:       netip.MustParseAddr("2001:db8::"),
					IPEnd:         netip.MustParseAddr("2001:db8:ffff:ffff:ffff:ffff:ffff:ffff"),
					CountryCode:   coredata.CountryCodeUS,
				},
				{
					AddressFamily: 6,
					IPStart:       netip.MustParseAddr("2001:db8:1::"),
					IPEnd:         netip.MustParseAddr("2001:db8:1:ffff:ffff:ffff:ffff:ffff"),
					CountryCode:   coredata.CountryCodeFR,
				},
			},
		},
	)
	require.NoError(t, err)

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

	_, err := service.ReplaceLocations(
		context.Background(),
		&blockSource{
			blocks: []coredata.IPLocationBlock{
				{
					AddressFamily: 4,
					IPStart:       netip.MustParseAddr("10.0.0.0"),
					IPEnd:         netip.MustParseAddr("10.0.255.255"),
					CountryCode:   coredata.CountryCodeUS,
				},
				{
					AddressFamily:   4,
					IPStart:         netip.MustParseAddr("10.0.0.0"),
					IPEnd:           netip.MustParseAddr("10.0.0.255"),
					CountryCode:     coredata.CountryCodeUS,
					SubdivisionCode: &subdivision,
				},
			},
		},
	)
	require.NoError(t, err)

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

	_, err := service.ReplaceLocations(
		context.Background(),
		&blockSource{
			blocks: []coredata.IPLocationBlock{
				{
					AddressFamily: 4,
					IPStart:       netip.MustParseAddr("10.0.0.0"),
					IPEnd:         netip.MustParseAddr("10.0.255.255"),
					CountryCode:   coredata.CountryCodeUS,
				},
				{
					AddressFamily:   4,
					IPStart:         netip.MustParseAddr("10.0.0.128"),
					IPEnd:           netip.MustParseAddr("10.0.0.255"),
					CountryCode:     coredata.CountryCodeUS,
					SubdivisionCode: &subdivision,
				},
			},
		},
	)
	require.NoError(t, err)

	location, err := service.LookupLocation(context.Background(), "10.0.0.200")
	require.NoError(t, err)
	assert.Equal(t, coredata.CountryCodeUS, location.CountryCode)
	assert.Equal(t, &subdivision, location.SubdivisionCode)

	location, err = service.LookupLocation(context.Background(), "10.0.1.0")
	require.NoError(t, err)
	assert.Equal(t, coredata.CountryCodeUS, location.CountryCode)
	assert.Nil(t, location.SubdivisionCode)
}
