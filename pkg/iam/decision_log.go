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

package iam

import (
	"context"
	"time"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
)

const effectError = "error"

// DecisionRecord is a structured authorization decision for logging.
type DecisionRecord struct {
	Effect     string
	Action     string
	ResourceID gid.GID
	Principal  gid.GID
	PolicyID   string
	Reason     string
	Latency    time.Duration
}

func newDecisionRecord(
	result policy.EvaluationResult,
	principal gid.GID,
	resourceID gid.GID,
	action string,
	role string,
	latency time.Duration,
) DecisionRecord {
	return DecisionRecord{
		Effect:     string(result.Decision),
		Action:     action,
		ResourceID: resourceID,
		Principal:  principal,
		PolicyID:   result.PolicyID(),
		Reason:     result.Reason(role),
		Latency:    latency,
	}
}

func (a *Authorizer) logDecision(ctx context.Context, rec DecisionRecord) {
	if a.logger == nil {
		return
	}

	if rec.PolicyID != "" {
		a.logger.InfoCtx(
			ctx,
			"authz decision",
			log.String("effect", rec.Effect),
			log.String("action", rec.Action),
			log.String("principal_id", rec.Principal.String()),
			log.String("resource_id", rec.ResourceID.String()),
			log.String("policy_id", rec.PolicyID),
			log.String("reason", rec.Reason),
			log.Duration("latency", rec.Latency),
		)

		return
	}

	a.logger.InfoCtx(
		ctx,
		"authz decision",
		log.String("effect", rec.Effect),
		log.String("action", rec.Action),
		log.String("principal_id", rec.Principal.String()),
		log.String("resource_id", rec.ResourceID.String()),
		log.String("reason", rec.Reason),
		log.Duration("latency", rec.Latency),
	)
}
