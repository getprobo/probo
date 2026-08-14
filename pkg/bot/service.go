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

package bot

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type (
	ServiceConfig struct {
		Disabled bool
	}

	Service struct {
		config ServiceConfig
		now    func() time.Time
	}

	MessageParams struct {
		OrganizationID   gid.GID
		Capability       string
		MessageType      string
		Attributes       map[string]any
		SubjectNamespace string
		SubjectKey       string
		EventKey         string
		Purpose          coredata.BotMessagePurpose
	}
)

func NewService(config ServiceConfig) *Service {
	return &Service{
		config: config,
		now:    time.Now,
	}
}

func (s *Service) EnqueueMessage(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	params MessageParams,
) (*coredata.BotMessage, error) {
	if s.config.Disabled {
		return nil, nil
	}

	if params.OrganizationID == gid.Nil ||
		params.Capability == "" ||
		params.MessageType == "" ||
		params.SubjectNamespace == "" ||
		params.SubjectKey == "" ||
		params.EventKey == "" ||
		!params.Purpose.IsValid() {
		return nil, fmt.Errorf("cannot enqueue bot message: parameters are incomplete")
	}

	attributes, err := json.Marshal(cloneAttributes(params.Attributes))
	if err != nil {
		return nil, fmt.Errorf("cannot encode bot message attributes: %w", err)
	}

	now := s.now()
	message := &coredata.BotMessage{
		ID:               gid.New(scope.GetTenantID(), coredata.BotMessageEntityType),
		OrganizationID:   params.OrganizationID,
		Capability:       params.Capability,
		MessageType:      params.MessageType,
		Attributes:       attributes,
		SubjectNamespace: params.SubjectNamespace,
		SubjectKey:       params.SubjectKey,
		EventKey:         params.EventKey,
		Purpose:          params.Purpose,
		MaxAttempts:      coredata.BotMessageDefaultMaxAttempts,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if _, err := message.EnqueueIdempotently(ctx, tx, scope); err != nil {
		return nil, fmt.Errorf("cannot enqueue bot message: %w", err)
	}

	return message, nil
}

func StableEventKey(action string, components ...string) string {
	normalized := slices.Clone(components)
	slices.Sort(normalized)

	hash := sha256.New()

	_, _ = fmt.Fprintf(hash, "%s\x00", action)
	for _, component := range normalized {
		_, _ = fmt.Fprintf(hash, "%s\x00", component)
	}

	return fmt.Sprintf("%x", hash.Sum(nil))
}

func cloneAttributes(attributes map[string]any) map[string]any {
	cloned := make(map[string]any, len(attributes))
	for key, value := range attributes {
		cloned[key] = value
	}

	return cloned
}
