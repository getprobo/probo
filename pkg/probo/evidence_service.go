// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package probo

import (
	"context"
	"fmt"
	"time"

	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filevalidation"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

type (
	EvidenceService struct {
		svc           *Service
		fileValidator *filevalidation.FileValidator
	}

	UploadMeasureEvidenceRequest struct {
		MeasureID gid.GID
		URL       *string
		File      FileUpload
	}
)

func (umer *UploadMeasureEvidenceRequest) Validate() error {
	v := validator.New()

	v.Check(umer.MeasureID, "measure_id", validator.Required(), validator.GID(coredata.MeasureEntityType))
	v.Check(umer.URL, "url", validator.URL())
	v.Check(umer.File, "file", validator.Required())

	return v.Error()
}

func (s EvidenceService) Get(
	ctx context.Context, scope coredata.Scoper,
	evidenceID gid.GID,
) (*coredata.Evidence, error) {
	evidence := &coredata.Evidence{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := evidence.LoadByID(ctx, conn, scope, evidenceID); err != nil {
				return fmt.Errorf("cannot load evidence %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load evidence: %w", err)
	}

	return evidence, nil
}

func (s EvidenceService) UploadMeasureEvidence(
	ctx context.Context, scope coredata.Scoper,
	req UploadMeasureEvidenceRequest,
) (*coredata.Evidence, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	now := time.Now()
	evidenceID := gid.New(scope.GetTenantID(), coredata.EvidenceEntityType)

	referenceID, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("cannot generate reference id: %w", err)
	}

	evidence := &coredata.Evidence{
		ID:                evidenceID,
		MeasureID:         req.MeasureID,
		State:             coredata.EvidenceStateFulfilled,
		ReferenceID:       "custom-evidence-" + referenceID.String(),
		Type:              coredata.EvidenceTypeFile,
		DescriptionStatus: coredata.EvidenceDescriptionStatusPending,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	err = s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			measure := &coredata.Measure{}

			var (
				file *coredata.File
				err  error
			)

			if err := measure.LoadByID(ctx, conn, scope, req.MeasureID); err != nil {
				return fmt.Errorf("cannot load measure %q: %w", req.MeasureID, err)
			}

			file, err = s.svc.Files.UploadAndSaveFile(
				ctx,
				scope,
				s.fileValidator,
				map[string]string{
					"type":            "evidence",
					"evidence-id":     evidenceID.String(),
					"organization-id": measure.OrganizationID.String(),
				},
				&req.File)
			if err != nil {
				return fmt.Errorf("cannot upload or file: %w", err)
			}

			evidence.OrganizationID = measure.OrganizationID
			evidence.EvidenceFileId = &file.ID
			evidence.MeasureID = req.MeasureID

			if err := evidence.Insert(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot insert evidence: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		// TODO try do delete file from s3 if it's a file type
		return nil, err
	}

	return evidence, nil
}

func (s EvidenceService) CountForMeasureID(
	ctx context.Context, scope coredata.Scoper,
	measureID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			evidences := coredata.Evidences{}

			count, err = evidences.CountByMeasureID(ctx, conn, scope, measureID)
			if err != nil {
				return fmt.Errorf("cannot count evidences: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s EvidenceService) ListForMeasureID(
	ctx context.Context, scope coredata.Scoper,
	measureID gid.GID,
	cursor *page.Cursor[coredata.EvidenceOrderField],
) (*page.Page[*coredata.Evidence, coredata.EvidenceOrderField], error) {
	var evidences coredata.Evidences

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return evidences.LoadByMeasureID(
				ctx,
				conn,
				scope,
				measureID,
				cursor,
			)
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(evidences, cursor), nil
}

func (s EvidenceService) CountForTaskID(
	ctx context.Context, scope coredata.Scoper,
	taskID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			evidences := coredata.Evidences{}

			count, err = evidences.CountByTaskID(ctx, conn, scope, taskID)
			if err != nil {
				return fmt.Errorf("cannot count evidences: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s EvidenceService) ListForTaskID(
	ctx context.Context, scope coredata.Scoper,
	taskID gid.GID,
	cursor *page.Cursor[coredata.EvidenceOrderField],
) (*page.Page[*coredata.Evidence, coredata.EvidenceOrderField], error) {
	var evidences coredata.Evidences

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return evidences.LoadByTaskID(
				ctx,
				conn,
				scope,
				taskID,
				cursor,
			)
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(evidences, cursor), nil
}

func (s *EvidenceService) Delete(
	ctx context.Context, scope coredata.Scoper,
	evidenceID gid.GID,
) error {
	evidence := &coredata.Evidence{ID: evidenceID}

	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			err := evidence.Delete(ctx, tx, scope)
			if err != nil {
				return fmt.Errorf("cannot delete evidence: %w", err)
			}

			return nil
		},
	)
}
