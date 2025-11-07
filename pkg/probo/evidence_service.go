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
		svc           *TenantService
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
	ctx context.Context,
	evidenceID gid.GID,
) (*coredata.Evidence, error) {
	evidence := &coredata.Evidence{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := evidence.LoadByID(ctx, conn, s.svc.scope, evidenceID); err != nil {
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
	ctx context.Context,
	req UploadMeasureEvidenceRequest,
) (*coredata.Evidence, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	now := time.Now()
	evidenceID := gid.New(s.svc.scope.GetTenantID(), coredata.EvidenceEntityType)

	referenceID, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("cannot generate reference id: %w", err)
	}

	evidence := &coredata.Evidence{
		ID:          evidenceID,
		MeasureID:   req.MeasureID,
		State:       coredata.EvidenceStateFulfilled,
		ReferenceID: "custom-evidence-" + referenceID.String(),
		Type:        coredata.EvidenceTypeFile,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err = s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			measure := &coredata.Measure{}
			var file *coredata.File
			var err error

			if err := measure.LoadByID(ctx, conn, s.svc.scope, req.MeasureID); err != nil {
				return fmt.Errorf("cannot load measure %q: %w", req.MeasureID, err)
			}

			file, err = s.svc.Files.UploadAndSaveFile(
				ctx,
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

			if err := evidence.Insert(ctx, conn, s.svc.scope); err != nil {
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
	ctx context.Context,
	measureID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			evidences := coredata.Evidences{}
			count, err = evidences.CountByMeasureID(ctx, conn, s.svc.scope, measureID)
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
	ctx context.Context,
	measureID gid.GID,
	cursor *page.Cursor[coredata.EvidenceOrderField],
) (*page.Page[*coredata.Evidence, coredata.EvidenceOrderField], error) {
	var evidences coredata.Evidences

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return evidences.LoadByMeasureID(
				ctx,
				conn,
				s.svc.scope,
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
	ctx context.Context,
	taskID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			evidences := coredata.Evidences{}
			count, err = evidences.CountByTaskID(ctx, conn, s.svc.scope, taskID)
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
	ctx context.Context,
	taskID gid.GID,
	cursor *page.Cursor[coredata.EvidenceOrderField],
) (*page.Page[*coredata.Evidence, coredata.EvidenceOrderField], error) {
	var evidences coredata.Evidences

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return evidences.LoadByTaskID(
				ctx,
				conn,
				s.svc.scope,
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
	ctx context.Context,
	evidenceID gid.GID,
) error {
	evidence := &coredata.Evidence{ID: evidenceID}

	return s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := evidence.Delete(ctx, conn, s.svc.scope)
			if err != nil {
				return fmt.Errorf("cannot delete evidence: %w", err)
			}

			return nil
		},
	)
}
