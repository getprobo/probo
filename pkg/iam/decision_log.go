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
