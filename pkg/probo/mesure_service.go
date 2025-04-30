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

	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
)

type (
	MeasureService struct {
		svc *TenantService
	}

	CreateMeasureRequest struct {
		OrganizationID gid.GID
		Name           string
		Description    string
		Category       string
	}

	UpdateMeasureRequest struct {
		ID          gid.GID
		Name        *string
		Description *string
		Category    *string
		State       *coredata.MeasureState
	}

	ImportMeasureRequest struct {
		Measures []struct {
			Name        string `json:"name"`
			Category    string `json:"category"`
			ReferenceID string `json:"reference-id"`
			Standards   []struct {
				Framework string `json:"framework"`
				Control   string `json:"control"`
			} `json:"standards"`
			Tasks []struct {
				Name               string `json:"name"`
				Description        string `json:"description"`
				ReferenceID        string `json:"reference-id"`
				RequestedEvidences []struct {
					ReferenceID string                `json:"reference-id"`
					Type        coredata.EvidenceType `json:"type"`
					Name        string                `json:"name"`
				} `json:"requested-evidences"`
			} `json:"tasks"`
		} `json:"measures"`
	}
)

func (s MeasureService) ListForRiskID(
	ctx context.Context,
	riskID gid.GID,
	cursor *page.Cursor[coredata.MeasureOrderField],
) (*page.Page[*coredata.Measure, coredata.MeasureOrderField], error) {
	var measures coredata.Measures

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return measures.LoadByRiskID(ctx, conn, s.svc.scope, riskID, cursor)
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(measures, cursor), nil
}

func (s MeasureService) ListForControlID(
	ctx context.Context,
	controlID gid.GID,
	cursor *page.Cursor[coredata.MeasureOrderField],
) (*page.Page[*coredata.Measure, coredata.MeasureOrderField], error) {
	var measures coredata.Measures

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return measures.LoadByControlID(ctx, conn, s.svc.scope, controlID, cursor)
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(measures, cursor), nil
}

func (s MeasureService) Get(
	ctx context.Context,
	measureID gid.GID,
) (*coredata.Measure, error) {
	measure := &coredata.Measure{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return measure.LoadByID(ctx, conn, s.svc.scope, measureID)
		},
	)

	if err != nil {
		return nil, err
	}

	return measure, nil
}

func (s MeasureService) Import(
	ctx context.Context,
	organizationID gid.GID,
	req ImportMeasureRequest,
) (*page.Page[*coredata.Measure, coredata.MeasureOrderField], error) {
	importedMeasures := coredata.Measures{}

	err := s.svc.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			for i := range req.Measures {
				now := time.Now()

				measureID, err := gid.NewGID(organizationID.TenantID(), coredata.MeasureEntityType)
				if err != nil {
					return fmt.Errorf("cannot create global id: %w", err)
				}

				measure := &coredata.Measure{
					ID:             measureID,
					OrganizationID: organizationID,
					Name:           req.Measures[i].Name,
					Description:    "",
					Category:       req.Measures[i].Category,
					State:          coredata.MeasureStateNotStarted,
					ReferenceID:    req.Measures[i].ReferenceID,
					CreatedAt:      now,
					UpdatedAt:      now,
				}

				importedMeasures = append(importedMeasures, measure)

				if err := measure.Upsert(ctx, tx, s.svc.scope); err != nil {
					return fmt.Errorf("cannot upsert measure: %w", err)
				}

				for j := range req.Measures[i].Tasks {
					taskID, err := gid.NewGID(organizationID.TenantID(), coredata.TaskEntityType)
					if err != nil {
						return fmt.Errorf("cannot create global id: %w", err)
					}

					task := &coredata.Task{
						ID:          taskID,
						MeasureID:   measure.ID,
						Name:        req.Measures[i].Tasks[j].Name,
						Description: req.Measures[i].Tasks[j].Description,
						ReferenceID: req.Measures[i].Tasks[j].ReferenceID,
						State:       coredata.TaskStateTodo,
						CreatedAt:   now,
						UpdatedAt:   now,
					}

					if err := task.Upsert(ctx, tx, s.svc.scope); err != nil {
						return fmt.Errorf("cannot upsert task: %w", err)
					}

					for k := range req.Measures[i].Tasks[j].RequestedEvidences {
						evidenceID, err := gid.NewGID(organizationID.TenantID(), coredata.EvidenceEntityType)
						if err != nil {
							return fmt.Errorf("cannot create global id: %w", err)
						}

						evidence := &coredata.Evidence{
							State:       coredata.EvidenceStateRequested,
							ID:          evidenceID,
							TaskID:      task.ID,
							ReferenceID: req.Measures[i].Tasks[j].RequestedEvidences[k].ReferenceID,
							Type:        req.Measures[i].Tasks[j].RequestedEvidences[k].Type,
							Description: req.Measures[i].Tasks[j].RequestedEvidences[k].Name,
							CreatedAt:   now,
							UpdatedAt:   now,
						}

						if err := evidence.Upsert(ctx, tx, s.svc.scope); err != nil {
							return fmt.Errorf("cannot upsert evidence: %w", err)
						}
					}
				}

				for _, standard := range req.Measures[i].Standards {
					framework := &coredata.Framework{}
					if err := framework.LoadByReferenceID(ctx, tx, s.svc.scope, standard.Framework); err != nil {
						continue
					}

					control := &coredata.Control{}
					if err := control.LoadByFrameworkIDAndReferenceID(ctx, tx, s.svc.scope, framework.ID, standard.Control); err != nil {
						continue
					}

					controlMeasure := &coredata.ControlMeasure{
						ControlID: control.ID,
						MeasureID: measure.ID,
						CreatedAt: now,
					}

					if err := controlMeasure.Upsert(ctx, tx, s.svc.scope); err != nil {
						return fmt.Errorf("cannot insert control measure: %w", err)
					}
				}
			}

			return nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("cannot import measures: %w", err)
	}

	cursor := page.NewCursor(
		len(importedMeasures),
		nil,
		page.Head,
		page.OrderBy[coredata.MeasureOrderField]{
			Field:     coredata.MeasureOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
	)

	return page.NewPage(importedMeasures, cursor), nil
}

func (s MeasureService) Update(
	ctx context.Context,
	req UpdateMeasureRequest,
) (*coredata.Measure, error) {
	measure := &coredata.Measure{ID: req.ID}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := measure.LoadByID(ctx, conn, s.svc.scope, req.ID); err != nil {
				return fmt.Errorf("cannot load measure: %w", err)
			}

			if req.Name != nil {
				measure.Name = *req.Name
			}

			if req.Description != nil {
				measure.Description = *req.Description
			}

			if req.Category != nil {
				measure.Category = *req.Category
			}

			if req.State != nil {
				measure.State = *req.State
			}

			measure.UpdatedAt = time.Now()

			if err := measure.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update measure: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return measure, nil
}

func (s MeasureService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.MeasureOrderField],
) (*page.Page[*coredata.Measure, coredata.MeasureOrderField], error) {
	var measures coredata.Measures

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return measures.LoadByOrganizationID(
				ctx,
				conn,
				s.svc.scope,
				organizationID,
				cursor,
			)
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(measures, cursor), nil
}

func (s MeasureService) Create(
	ctx context.Context,
	req CreateMeasureRequest,
) (*coredata.Measure, error) {
	now := time.Now()
	measureID, err := gid.NewGID(s.svc.scope.GetTenantID(), coredata.MeasureEntityType)
	if err != nil {
		return nil, fmt.Errorf("cannot create measure global id: %w", err)
	}

	referenceID, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("cannot generate reference id: %w", err)
	}

	measure := &coredata.Measure{
		ID:             measureID,
		OrganizationID: req.OrganizationID,
		Name:           req.Name,
		Description:    req.Description,
		Category:       req.Category,
		ReferenceID:    "custom-measure-" + referenceID.String(),
		State:          coredata.MeasureStateNotStarted,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err = s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := measure.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert measure: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return measure, nil
}

func (s MeasureService) Delete(
	ctx context.Context,
	measureID gid.GID,
) error {
	return s.svc.pg.WithTx(ctx, func(conn pg.Conn) error {
		measure := &coredata.Measure{ID: measureID}

		if err := measure.Delete(ctx, conn, s.svc.scope); err != nil {
			return fmt.Errorf("cannot delete measure: %w", err)
		}

		return nil
	})
}
