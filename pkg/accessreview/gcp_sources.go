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

package accessreview

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	cloudgcp "go.probo.inc/probo/pkg/cloud/gcp"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

const (
	MaxGCPSourceProjects = 100
)

type (
	GCPSourceFailureReason string

	GCPSourceProject struct {
		ProjectID                string
		WorkloadIdentityProvider string
		ServiceAccountEmail      string
	}

	GCPSourceFailure struct {
		Index     int
		ProjectID *string
		Reason    GCPSourceFailureReason
	}
)

const (
	GCPSourceFailureInvalid      GCPSourceFailureReason = "INVALID"
	GCPSourceFailureDisconnected GCPSourceFailureReason = "DISCONNECTED"
	GCPSourceFailureFailed       GCPSourceFailureReason = "FAILED"
)

var (
	ErrGCPSourcesEmpty   = errors.New("projects is required")
	ErrGCPSourcesTooMany = errors.New("cannot create more than 100 GCP access sources at once")
)

// CreateGCPSources creates one workload-identity connector and access
// source per project. Invalid settings and failed probes are recorded
// in failures; remaining projects still succeed.
func (s *Service) CreateGCPSources(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	projects []GCPSourceProject,
) ([]*coredata.AccessReviewSource, []GCPSourceFailure, error) {
	if len(projects) == 0 {
		return nil, nil, ErrGCPSourcesEmpty
	}

	if len(projects) > MaxGCPSourceProjects {
		return nil, nil, ErrGCPSourcesTooMany
	}

	sources := make([]*coredata.AccessReviewSource, 0, len(projects))
	failures := make([]GCPSourceFailure, 0)

	for i, project := range projects {
		source, reason, err := s.createGCPSource(ctx, scope, organizationID, project)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot create gcp access source: %w", err)
		}

		if reason != "" {
			failures = append(failures, GCPSourceFailure{
				Index:     i,
				ProjectID: optionalProjectID(project.ProjectID),
				Reason:    reason,
			})

			continue
		}

		sources = append(sources, source)
	}

	return sources, failures, nil
}

func (s *Service) createGCPSource(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	project GCPSourceProject,
) (*coredata.AccessReviewSource, GCPSourceFailureReason, error) {
	raw, providerResource, reason, err := marshalGCPWorkloadIdentitySettings(project)
	if err != nil {
		return nil, "", err
	}

	if reason != "" {
		return nil, reason, nil
	}

	cnnctr, err := s.insertGCPConnector(ctx, scope, organizationID, raw)
	if err != nil {
		s.logger.ErrorCtx(ctx, "cannot create gcp workload identity connector", log.Error(err))

		return nil, GCPSourceFailureFailed, nil
	}

	if err := s.ProbeConnector(ctx, scope, cnnctr.ID); err != nil {
		s.abandonGCPConnector(ctx, scope, cnnctr.ID)

		if _, ok := errors.AsType[*ProbeError](err); ok {
			return nil, GCPSourceFailureDisconnected, nil
		}

		s.logger.ErrorCtx(
			ctx,
			"cannot probe gcp connector",
			log.String("connector_id", cnnctr.ID.String()),
			log.Error(err),
		)

		return nil, GCPSourceFailureFailed, nil
	}

	source, _, err := s.EnsureSource(
		ctx,
		scope,
		CreateAccessReviewSourceRequest{
			OrganizationID: organizationID,
			ConnectorID:    &cnnctr.ID,
			Name:           gcpAccessReviewSourceName(project.ProjectID, providerResource),
		},
	)
	if err != nil {
		s.abandonGCPConnector(ctx, scope, cnnctr.ID)
		s.logger.ErrorCtx(ctx, "cannot create gcp access source", log.Error(err))

		return nil, GCPSourceFailureFailed, nil
	}

	s.AutoSelectDefaultOrganization(ctx, scope, source)

	return source, "", nil
}

func marshalGCPWorkloadIdentitySettings(
	project GCPSourceProject,
) ([]byte, string, GCPSourceFailureReason, error) {
	if project.WorkloadIdentityProvider == "" || project.ServiceAccountEmail == "" {
		return nil, "", GCPSourceFailureInvalid, nil
	}

	validated, err := cloudgcp.NewConnectorSettings(
		project.WorkloadIdentityProvider,
		project.ServiceAccountEmail,
	)
	if err != nil {
		return nil, "", GCPSourceFailureInvalid, nil
	}

	raw, err := json.Marshal(
		coredata.GCPConnectorSettings{
			WorkloadIdentityProvider: validated.WorkloadIdentityProvider,
			ServiceAccountEmail:      validated.ServiceAccountEmail,
		},
	)
	if err != nil {
		return nil, "", "", fmt.Errorf("cannot marshal gcp connector settings: %w", err)
	}

	return raw, validated.WorkloadIdentityProvider, "", nil
}

func (s *Service) insertGCPConnector(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	rawSettings []byte,
) (*coredata.Connector, error) {
	now := time.Now()
	cnnctr := &coredata.Connector{
		ID:             gid.New(scope.GetTenantID(), coredata.ConnectorEntityType),
		OrganizationID: organizationID,
		Provider:       coredata.ConnectorProviderGCP,
		Protocol:       coredata.ConnectorProtocolWorkloadIdentity,
		Connection:     &connector.WorkloadIdentityConnection{},
		RawSettings:    rawSettings,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := cnnctr.Insert(ctx, tx, scope, s.encryptionKey); err != nil {
				return fmt.Errorf("cannot insert connector: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return cnnctr, nil
}

func (s *Service) abandonGCPConnector(
	ctx context.Context,
	scope coredata.Scoper,
	connectorID gid.GID,
) {
	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			cnnctr := &coredata.Connector{ID: connectorID}
			if err := cnnctr.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete connector: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		s.logger.ErrorCtx(
			ctx,
			"cannot delete leftover gcp connector",
			log.String("connector_id", connectorID.String()),
			log.Error(err),
		)
	}
}

func gcpAccessReviewSourceName(projectID, providerResource string) string {
	if id := strings.TrimSpace(projectID); id != "" {
		return "GCP / " + id
	}

	if number := gcpProjectNumber(providerResource); number != "" {
		return "GCP / " + number
	}

	return "GCP"
}

func gcpProjectNumber(providerResource string) string {
	parts := strings.Split(providerResource, "/")
	if len(parts) >= 2 && parts[0] == "projects" {
		return parts[1]
	}

	return ""
}

func optionalProjectID(projectID string) *string {
	if id := strings.TrimSpace(projectID); id != "" {
		return &id
	}

	return nil
}
