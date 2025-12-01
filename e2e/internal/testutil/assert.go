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

package testutil

import (
	"cmp"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type PageInfo struct {
	HasNextPage     bool    `json:"hasNextPage"`
	HasPreviousPage bool    `json:"hasPreviousPage"`
	StartCursor     *string `json:"startCursor"`
	EndCursor       *string `json:"endCursor"`
}

func AssertFirstPage(t *testing.T, edgeCount int, pageInfo PageInfo, expectedCount int, expectMore bool) {
	t.Helper()
	assert.Equal(t, expectedCount, edgeCount, "unexpected number of edges")
	assert.Equal(t, expectMore, pageInfo.HasNextPage, "hasNextPage mismatch")
	assert.False(t, pageInfo.HasPreviousPage, "first page should not have previous page")
}

func AssertMiddlePage(t *testing.T, edgeCount int, pageInfo PageInfo, expectedCount int) {
	t.Helper()
	assert.Equal(t, expectedCount, edgeCount, "unexpected number of edges")
	assert.True(t, pageInfo.HasNextPage, "middle page should have next page")
	assert.True(t, pageInfo.HasPreviousPage, "middle page should have previous page")
}

func AssertLastPage(t *testing.T, edgeCount int, pageInfo PageInfo, expectedCount int, expectPrevious bool) {
	t.Helper()
	assert.Equal(t, expectedCount, edgeCount, "unexpected number of edges")
	assert.False(t, pageInfo.HasNextPage, "last page should not have next page")
	assert.Equal(t, expectPrevious, pageInfo.HasPreviousPage, "hasPreviousPage mismatch")
}

func AssertHasMorePages(t *testing.T, pageInfo PageInfo) {
	t.Helper()
	assert.True(t, pageInfo.HasNextPage, "expected more pages")
	assert.NotNil(t, pageInfo.EndCursor, "endCursor should be set when there are more pages")
}

func AssertHasPreviousPages(t *testing.T, pageInfo PageInfo) {
	t.Helper()
	assert.True(t, pageInfo.HasPreviousPage, "expected previous pages")
	assert.NotNil(t, pageInfo.StartCursor, "startCursor should be set when there are previous pages")
}

func AssertTimestampsOnCreate(t *testing.T, createdAt, updatedAt, beforeCreate time.Time) {
	t.Helper()
	assert.True(t, createdAt.After(beforeCreate), "createdAt should be after test start")
	assert.True(t, updatedAt.After(beforeCreate), "updatedAt should be after test start")
	assert.Equal(t, createdAt, updatedAt, "createdAt and updatedAt should be equal on create")
}

func AssertTimestampsOnUpdate(t *testing.T, createdAt, updatedAt, originalCreatedAt, originalUpdatedAt time.Time) {
	t.Helper()
	assert.Equal(t, originalCreatedAt, createdAt, "createdAt should not change on update")
	assert.True(t, updatedAt.After(originalUpdatedAt) || updatedAt.Equal(originalUpdatedAt),
		"updatedAt should be >= previous updatedAt")
}

func AssertOptionalStringEqual(t *testing.T, expected, actual *string, fieldName string) {
	t.Helper()
	if expected == nil {
		assert.Nil(t, actual, "%s should be nil", fieldName)
	} else {
		require.NotNil(t, actual, "%s should not be nil", fieldName)
		assert.Equal(t, *expected, *actual, "%s mismatch", fieldName)
	}
}

func AssertOrderedAscending[T cmp.Ordered](t *testing.T, values []T, fieldName string) {
	t.Helper()
	assert.True(t, slices.IsSorted(values), "%s should be in ascending order, got: %v", fieldName, values)
}

func AssertOrderedDescending[T cmp.Ordered](t *testing.T, values []T, fieldName string) {
	t.Helper()
	reversed := slices.Clone(values)
	slices.Reverse(reversed)
	assert.True(t, slices.IsSorted(reversed), "%s should be in descending order, got: %v", fieldName, values)
}

func AssertTimesOrderedAscending(t *testing.T, times []time.Time, fieldName string) {
	t.Helper()
	isSorted := slices.IsSortedFunc(times, func(a, b time.Time) int {
		return a.Compare(b)
	})
	assert.True(t, isSorted, "%s should be in ascending order", fieldName)
}

func AssertTimesOrderedDescending(t *testing.T, times []time.Time, fieldName string) {
	t.Helper()
	isSorted := slices.IsSortedFunc(times, func(a, b time.Time) int {
		return b.Compare(a)
	})
	assert.True(t, isSorted, "%s should be in descending order", fieldName)
}

func AssertNodeNotAccessible(t *testing.T, err error, nodeIsNil bool, resourceType string) {
	t.Helper()
	if err == nil {
		assert.True(t, nodeIsNil, "should not be able to access %s from another org", resourceType)
	}
	// If there's an error, that's also acceptable (access denied)
}

func RequireForbiddenError(t *testing.T, err error, msgAndArgs ...any) {
	t.Helper()
	RequireErrorCode(t, err, "FORBIDDEN", msgAndArgs...)
}

func RequireErrorCode(t *testing.T, err error, code string, msgAndArgs ...any) {
	t.Helper()
	require.Error(t, err, msgAndArgs...)

	var gqlErrors GraphQLErrors
	if !assert.ErrorAs(t, err, &gqlErrors) {
		t.Fatalf("expected GraphQL error, got: %T: %v", err, err)
	}

	if len(gqlErrors) == 0 {
		t.Fatalf("expected at least one GraphQL error, got none")
	}

	if gqlErrors[0].Code() != code {
		t.Fatalf("expected %s error code, got %q with message: %q",
			code, gqlErrors[0].Code(), gqlErrors[0].Message)
	}
}
