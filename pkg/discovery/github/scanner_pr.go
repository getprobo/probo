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

package github

import (
	"context"
	"strconv"
	"strings"
)

type (
	pullRequestItem struct {
		Number   int     `json:"number"`
		MergedAt *string `json:"merged_at"`
	}

	pullReviewItem struct {
		State string `json:"state"`
	}
)

func (s *discoveryScanner) scanRepoPullRequestPractice(
	ctx context.Context,
	repo repoListItem,
	agg *repoScanAggregate,
) {
	endpoint, err := s.api.repoEndpoint(s.org, repo.Name, "pulls")
	if err != nil {
		return
	}

	endpoint, err = withPerPage(endpoint, 20)
	if err != nil {
		return
	}

	endpoint, err = appendQuery(endpoint, "state", "closed")
	if err != nil {
		return
	}

	var pulls []pullRequestItem

	if _, err := s.api.getJSON(ctx, endpoint, &pulls); err != nil {
		return
	}

	reviewed := 0
	sampled := 0

	for _, pull := range pulls {
		if pull.MergedAt == nil || *pull.MergedAt == "" {
			continue
		}

		sampled++

		if s.pullRequestHasApproval(ctx, repo, pull.Number) {
			reviewed++
		}

		if sampled >= 10 {
			break
		}
	}

	if sampled == 0 {
		return
	}

	agg.PRSampled += sampled
	agg.PRReviewed += reviewed

	if reviewed*100/sampled >= 80 {
		agg.WithDeFactoPRReview++
	}
}

func (s *discoveryScanner) pullRequestHasApproval(
	ctx context.Context,
	repo repoListItem,
	number int,
) bool {
	endpoint, err := s.api.repoEndpoint(
		s.org,
		repo.Name,
		"pulls",
		strconv.Itoa(number),
		"reviews",
	)
	if err != nil {
		return false
	}

	var reviews []pullReviewItem

	if _, err := s.api.getJSON(ctx, endpoint, &reviews); err != nil {
		return false
	}

	for _, review := range reviews {
		if strings.EqualFold(review.State, "APPROVED") {
			return true
		}
	}

	return false
}
