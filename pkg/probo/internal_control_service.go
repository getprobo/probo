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
	"errors"
	"fmt"
	"strings"
	"time"

	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/prosemirror"
	taskpkg "go.probo.inc/probo/pkg/task"
	"go.probo.inc/probo/pkg/timespan"
	"go.probo.inc/probo/pkg/validator"
)

type (
	InternalControlService struct {
		svc *Service
	}

	CreateInternalControlRequest struct {
		OrganizationID       gid.GID
		Name                 string
		Description          *string
		Category             string
		Code                 *string
		ControlType          *coredata.InternalControlType
		Nature               *coredata.InternalControlNature
		OperatingFrequency   *coredata.InternalControlOperatingFrequency
		EvidenceCadence      *timespan.TimeSpan
		TestingCadence       *timespan.TimeSpan
		ImplementationStatus *coredata.InternalControlImplementationStatus
		OwnerID              *gid.GID
		ReviewerID           *gid.GID
	}

	UpdateInternalControlRequest struct {
		ID                   gid.GID
		Name                 *string
		Description          **string
		Category             *string
		State                *coredata.InternalControlState
		Code                 **string
		ControlType          **coredata.InternalControlType
		Nature               **coredata.InternalControlNature
		OperatingFrequency   **coredata.InternalControlOperatingFrequency
		EvidenceCadence      **timespan.TimeSpan
		TestingCadence       **timespan.TimeSpan
		ImplementationStatus *coredata.InternalControlImplementationStatus
		OwnerID              **gid.GID
		ReviewerID           **gid.GID
	}

	ImportInternalControlRequest struct {
		IdentityID       *gid.GID
		InternalControls []struct {
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
		} `json:"internalControls"`
	}
)

func (cmr *CreateInternalControlRequest) Validate() error {
	v := validator.New()

	cmr.Code = normalizeInternalControlCode(cmr.Code)
	normalizeOperatingFrequency(cmr.OperatingFrequency)
	cmr.EvidenceCadence = normalizeNonPositiveCadence(cmr.EvidenceCadence)
	cmr.TestingCadence = normalizeNonPositiveCadence(cmr.TestingCadence)

	v.Check(cmr.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(cmr.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(cmr.Description, "description", validator.SafeText(ContentMaxLength))
	v.Check(cmr.Category, "category", validator.Required(), validator.SafeText(TitleMaxLength))
	v.Check(cmr.Code, "code", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(cmr.ControlType, "control_type", validator.OneOfSlice(coredata.InternalControlTypes()))
	v.Check(cmr.Nature, "nature", validator.OneOfSlice(coredata.InternalControlNatures()))
	v.Check(cmr.OperatingFrequency, "operating_frequency", validOperatingFrequency())
	v.Check(cmr.EvidenceCadence, "evidence_cadence", positiveTimeSpan())
	v.Check(cmr.TestingCadence, "testing_cadence", positiveTimeSpan())
	v.Check(cmr.ImplementationStatus, "implementation_status", validator.OneOfSlice(coredata.InternalControlImplementationStatuses()))
	v.Check(cmr.OwnerID, "owner_id", validator.GID(coredata.MembershipProfileEntityType))
	v.Check(cmr.ReviewerID, "reviewer_id", validator.GID(coredata.MembershipProfileEntityType))
	v.Check(
		internalControlPeople{ownerID: cmr.OwnerID, reviewerID: cmr.ReviewerID},
		"reviewer_id",
		distinctInternalControlPeople(),
	)

	return v.Error()
}

func (umr *UpdateInternalControlRequest) Validate() error {
	v := validator.New()

	normalizeOmittableInternalControlCode(umr.Code)

	if umr.OperatingFrequency != nil {
		normalizeOperatingFrequency(*umr.OperatingFrequency)
	}

	normalizeOmittableCadence(umr.EvidenceCadence)
	normalizeOmittableCadence(umr.TestingCadence)

	v.Check(umr.ID, "id", validator.Required(), validator.GID(coredata.InternalControlEntityType))
	v.Check(umr.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(umr.Description, "description", validator.SafeText(ContentMaxLength))
	v.Check(umr.Category, "category", validator.SafeText(TitleMaxLength))
	v.Check(umr.State, "state", validator.OneOfSlice(coredata.InternalControlStates()))
	v.Check(umr.Code, "code", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(umr.ControlType, "control_type", validator.OneOfSlice(coredata.InternalControlTypes()))
	v.Check(umr.Nature, "nature", validator.OneOfSlice(coredata.InternalControlNatures()))
	v.Check(umr.OperatingFrequency, "operating_frequency", validOperatingFrequency())
	v.Check(umr.EvidenceCadence, "evidence_cadence", positiveTimeSpan())
	v.Check(umr.TestingCadence, "testing_cadence", positiveTimeSpan())
	v.Check(umr.ImplementationStatus, "implementation_status", validator.OneOfSlice(coredata.InternalControlImplementationStatuses()))
	v.Check(umr.OwnerID, "owner_id", validator.GID(coredata.MembershipProfileEntityType))
	v.Check(umr.ReviewerID, "reviewer_id", validator.GID(coredata.MembershipProfileEntityType))

	return v.Error()
}

func (s InternalControlService) CountForRiskID(
	ctx context.Context, scope coredata.Scoper,
	riskID gid.GID,
	filter *coredata.InternalControlFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			internalControls := &coredata.InternalControls{}

			count, err = internalControls.CountByRiskID(ctx, conn, scope, riskID, filter)
			if err != nil {
				return fmt.Errorf("cannot count internalControls: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}
func (s InternalControlService) ListForRiskID(
	ctx context.Context, scope coredata.Scoper,
	riskID gid.GID,
	cursor *page.Cursor[coredata.InternalControlOrderField],
	filter *coredata.InternalControlFilter,
) (*page.Page[*coredata.InternalControl, coredata.InternalControlOrderField], error) {
	var internalControls coredata.InternalControls

	risk := &coredata.Risk{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := risk.LoadByID(ctx, conn, scope, riskID); err != nil {
				return fmt.Errorf("cannot load risk: %w", err)
			}

			err := internalControls.LoadByRiskID(ctx, conn, scope, risk.ID, cursor, filter)
			if err != nil {
				return fmt.Errorf("cannot load internalControls: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(internalControls, cursor), nil
}

func (s InternalControlService) CountForTreatmentPlanID(
	ctx context.Context,
	scope coredata.Scoper,
	treatmentPlanID gid.GID,
	filter *coredata.InternalControlFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			internalControls := &coredata.InternalControls{}

			count, err = internalControls.CountByTreatmentPlanID(ctx, conn, scope, treatmentPlanID, filter)
			if err != nil {
				return fmt.Errorf("cannot count internalControls: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s InternalControlService) ListForTreatmentPlanID(
	ctx context.Context,
	scope coredata.Scoper,
	treatmentPlanID gid.GID,
	cursor *page.Cursor[coredata.InternalControlOrderField],
	filter *coredata.InternalControlFilter,
) (*page.Page[*coredata.InternalControl, coredata.InternalControlOrderField], error) {
	var internalControls coredata.InternalControls

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			tp := &coredata.TreatmentPlan{}
			if err := tp.LoadByID(ctx, conn, scope, treatmentPlanID); err != nil {
				return fmt.Errorf("cannot load treatment plan: %w", err)
			}

			err := internalControls.LoadByTreatmentPlanID(ctx, conn, scope, tp.ID, cursor, filter)
			if err != nil {
				return fmt.Errorf("cannot load internalControls: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(internalControls, cursor), nil
}

func (s InternalControlService) CountForControlID(
	ctx context.Context, scope coredata.Scoper,
	controlID gid.GID,
	filter *coredata.InternalControlFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			internalControls := &coredata.InternalControls{}

			count, err = internalControls.CountByControlID(ctx, conn, scope, controlID, filter)
			if err != nil {
				return fmt.Errorf("cannot count internalControls: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s InternalControlService) ListForControlID(
	ctx context.Context, scope coredata.Scoper,
	controlID gid.GID,
	cursor *page.Cursor[coredata.InternalControlOrderField],
	filter *coredata.InternalControlFilter,
) (*page.Page[*coredata.InternalControl, coredata.InternalControlOrderField], error) {
	var internalControls coredata.InternalControls

	control := &coredata.Control{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := control.LoadByID(ctx, conn, scope, controlID); err != nil {
				return fmt.Errorf("cannot load control: %w", err)
			}

			err := internalControls.LoadByControlID(ctx, conn, scope, control.ID, cursor, filter)
			if err != nil {
				return fmt.Errorf("cannot load internalControls: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(internalControls, cursor), nil
}

func (s InternalControlService) CountForOrganizationID(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
	filter *coredata.InternalControlFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			internalControls := &coredata.InternalControls{}

			count, err = internalControls.CountByOrganizationID(ctx, conn, scope, organizationID, filter)
			if err != nil {
				return fmt.Errorf("cannot count internalControls: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s InternalControlService) ListDistinctCategoriesForOrganizationID(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
) ([]string, error) {
	var categories []string

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			var (
				internalControls coredata.InternalControls
				err              error
			)

			categories, err = internalControls.LoadDistinctCategoriesByOrganizationID(
				ctx,
				conn,
				scope,
				organization.ID,
			)
			if err != nil {
				return fmt.Errorf("cannot load internal control categories: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return categories, nil
}

func (s InternalControlService) ListForOrganizationID(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.InternalControlOrderField],
	filter *coredata.InternalControlFilter,
) (*page.Page[*coredata.InternalControl, coredata.InternalControlOrderField], error) {
	var internalControls coredata.InternalControls

	organization := &coredata.Organization{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := organization.LoadByID(ctx, conn, scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			err := internalControls.LoadByOrganizationID(
				ctx,
				conn,
				scope,
				organization.ID,
				cursor,
				filter,
			)
			if err != nil {
				return fmt.Errorf("cannot load internalControls: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(internalControls, cursor), nil
}

func (s InternalControlService) Get(
	ctx context.Context, scope coredata.Scoper,
	internalControlID gid.GID,
) (*coredata.InternalControl, error) {
	internalControl := &coredata.InternalControl{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return internalControl.LoadByID(ctx, conn, scope, internalControlID)
		},
	)
	if err != nil {
		return nil, err
	}

	return internalControl, nil
}

func (s InternalControlService) GetByIDs(
	ctx context.Context, scope coredata.Scoper,
	internalControlIDs ...gid.GID,
) (coredata.InternalControls, error) {
	var internalControls coredata.InternalControls

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := internalControls.LoadByIDs(
				ctx,
				conn,
				scope,
				internalControlIDs,
			); err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
				return fmt.Errorf("cannot load internal controls by ids: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return internalControls, nil
}

func (s InternalControlService) Import(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
	req ImportInternalControlRequest,
) (*page.Page[*coredata.InternalControl, coredata.InternalControlOrderField], error) {
	importedInternalControls := coredata.InternalControls{}
	organization := &coredata.Organization{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := organization.LoadByID(ctx, tx, scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			actorID, err := taskpkg.ResolveActivityActorID(
				ctx,
				tx,
				scope,
				req.IdentityID,
				organization.ID,
			)
			if err != nil {
				return fmt.Errorf("cannot resolve task activity actor: %w", err)
			}

			for i := range req.InternalControls {
				now := time.Now()

				internalControlID := gid.New(organization.ID.TenantID(), coredata.InternalControlEntityType)

				internalControl := &coredata.InternalControl{
					ID:                   internalControlID,
					OrganizationID:       organization.ID,
					Name:                 req.InternalControls[i].Name,
					Description:          nil,
					Category:             req.InternalControls[i].Category,
					State:                coredata.InternalControlStateNotStarted,
					ReferenceID:          req.InternalControls[i].ReferenceID,
					ImplementationStatus: coredata.InternalControlImplementationStatusNotImplemented,
					CreatedAt:            now,
					UpdatedAt:            now,
				}

				importedInternalControls = append(importedInternalControls, internalControl)

				originalID := internalControl.ID
				if err := internalControl.Upsert(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot upsert internalControl: %w", err)
				}

				eventType := coredata.InternalControlEventTypeCreated
				if originalID != internalControl.ID {
					eventType = coredata.InternalControlEventTypeUpdated
				}

				if err := insertInternalControlEvent(
					ctx,
					tx,
					scope,
					internalControl,
					eventType,
					now,
				); err != nil {
					return fmt.Errorf("cannot record internal control event: %w", err)
				}

				for j := range req.InternalControls[i].Tasks {
					taskID := gid.New(organization.ID.TenantID(), coredata.TaskEntityType)

					taskDescription := req.InternalControls[i].Tasks[j].Description
					task := &coredata.Task{
						ID:             taskID,
						OrganizationID: organizationID,
						Name:           req.InternalControls[i].Tasks[j].Name,
						Content:        prosemirror.FromPlainText(taskDescription),
						ReferenceID:    req.InternalControls[i].Tasks[j].ReferenceID,
						State:          coredata.TaskStateTodo,
						Priority:       coredata.TaskPriorityMedium,
						CreatedAt:      now,
						UpdatedAt:      now,
					}

					existingTask := &coredata.Task{}

					existingErr := existingTask.LoadByInternalControlIDAndReferenceID(
						ctx,
						tx,
						scope,
						internalControl.ID,
						req.InternalControls[i].Tasks[j].ReferenceID,
					)
					if existingErr != nil && !errors.Is(existingErr, coredata.ErrResourceNotFound) {
						return fmt.Errorf("cannot load task: %w", existingErr)
					}

					if errors.Is(existingErr, coredata.ErrResourceNotFound) {
						if err := task.Insert(ctx, tx, scope); err != nil {
							return fmt.Errorf("cannot insert task: %w", err)
						}

						link := coredata.InternalControlTask{
							InternalControlID: internalControl.ID,
							TaskID:            task.ID,
							OrganizationID:    internalControl.OrganizationID,
							ReferenceID:       task.ReferenceID,
							CreatedAt:         task.CreatedAt,
						}
						if err := link.Upsert(ctx, tx, scope); err != nil {
							return fmt.Errorf("cannot link task to internal control: %w", err)
						}

						if err := taskpkg.InsertCreatedActivity(
							ctx,
							tx,
							scope,
							task,
							actorID,
							now,
						); err != nil {
							return fmt.Errorf("cannot record task created event: %w", err)
						}
					} else {
						previous := *existingTask
						existingTask.Name = task.Name
						existingTask.Content = task.Content
						existingTask.UpdatedAt = task.UpdatedAt

						if err := existingTask.UpdateNameAndContent(ctx, tx, scope); err != nil {
							return fmt.Errorf("cannot update imported task: %w", err)
						}

						if err := taskpkg.InsertUpdateActivities(
							ctx,
							tx,
							scope,
							&previous,
							existingTask,
							actorID,
							now,
						); err != nil {
							return fmt.Errorf("cannot record task update events: %w", err)
						}

						task = existingTask
					}

					for k := range req.InternalControls[i].Tasks[j].RequestedEvidences {
						evidenceID := gid.New(organizationID.TenantID(), coredata.EvidenceEntityType)

						evidenceDescription := req.InternalControls[i].Tasks[j].RequestedEvidences[k].Name
						evidence := &coredata.Evidence{
							State:             coredata.EvidenceStateRequested,
							ID:                evidenceID,
							TaskID:            &task.ID,
							ReferenceID:       req.InternalControls[i].Tasks[j].RequestedEvidences[k].ReferenceID,
							Type:              req.InternalControls[i].Tasks[j].RequestedEvidences[k].Type,
							Description:       &evidenceDescription,
							DescriptionStatus: coredata.EvidenceDescriptionStatusPending,
							CreatedAt:         now,
							UpdatedAt:         now,
						}

						if err := evidence.Upsert(ctx, tx, scope); err != nil {
							return fmt.Errorf("cannot upsert evidence: %w", err)
						}
					}
				}

				for _, standard := range req.InternalControls[i].Standards {
					framework := &coredata.Framework{}
					if err := framework.LoadByReferenceID(ctx, tx, scope, standard.Framework); err != nil {
						continue
					}

					control := &coredata.Control{}
					if err := control.LoadByFrameworkIDAndSectionTitle(ctx, tx, scope, framework.ID, standard.Control); err != nil {
						continue
					}

					controlInternalControl := &coredata.ControlInternalControl{
						ControlID:         control.ID,
						InternalControlID: internalControl.ID,
						OrganizationID:    internalControl.OrganizationID,
						CreatedAt:         now,
					}

					if err := controlInternalControl.Upsert(ctx, tx, scope); err != nil {
						return fmt.Errorf("cannot insert control internalControl: %w", err)
					}
				}
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot import internalControls: %w", err)
	}

	cursor := page.NewCursor(
		len(importedInternalControls),
		nil,
		page.Head,
		page.OrderBy[coredata.InternalControlOrderField]{
			Field:     coredata.InternalControlOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
	)

	return page.NewPage(importedInternalControls, cursor), nil
}

func (s InternalControlService) Update(
	ctx context.Context, scope coredata.Scoper,
	req UpdateInternalControlRequest,
) (*coredata.InternalControl, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	internalControl := &coredata.InternalControl{ID: req.ID}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := internalControl.LoadByID(ctx, conn, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load internalControl: %w", err)
			}

			previousName := internalControl.Name
			previousState := internalControl.State
			previousCategory := internalControl.Category

			if req.Name != nil {
				internalControl.Name = *req.Name
			}

			if req.Description != nil {
				internalControl.Description = *req.Description
			}

			if req.Category != nil {
				internalControl.Category = *req.Category
			}

			now := time.Now()
			applyInternalControlUpdate(&req, internalControl, now)

			if err := distinctInternalControlPeopleError(internalControl.OwnerID, internalControl.ReviewerID); err != nil {
				return err
			}

			if internalControl.OwnerID != nil {
				if err := loadInternalControlProfile(ctx, conn, scope, *internalControl.OwnerID, internalControl.OrganizationID); err != nil {
					return err
				}
			}

			if internalControl.ReviewerID != nil {
				if err := loadInternalControlProfile(ctx, conn, scope, *internalControl.ReviewerID, internalControl.OrganizationID); err != nil {
					return err
				}
			}

			internalControl.UpdatedAt = now

			if err := internalControl.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update internalControl: %w", err)
			}

			if internalControl.Name != previousName || internalControl.State != previousState || internalControl.Category != previousCategory {
				if err := insertInternalControlEvent(
					ctx,
					conn,
					scope,
					internalControl,
					coredata.InternalControlEventTypeUpdated,
					internalControl.UpdatedAt,
				); err != nil {
					return fmt.Errorf("cannot record internal control event: %w", err)
				}
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return internalControl, nil
}

func (s InternalControlService) Create(
	ctx context.Context, scope coredata.Scoper,
	req CreateInternalControlRequest,
) (*coredata.InternalControl, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	now := time.Now()

	var internalControl *coredata.InternalControl

	organization := &coredata.Organization{}

	referenceID, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("cannot generate reference id: %w", err)
	}

	err = s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := organization.LoadByID(ctx, conn, scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			if req.OwnerID != nil {
				if err := loadInternalControlProfile(ctx, conn, scope, *req.OwnerID, organization.ID); err != nil {
					return err
				}
			}

			if req.ReviewerID != nil {
				if err := loadInternalControlProfile(ctx, conn, scope, *req.ReviewerID, organization.ID); err != nil {
					return err
				}
			}

			status := coredata.InternalControlImplementationStatusNotImplemented
			state := coredata.InternalControlStateNotStarted

			if req.ImplementationStatus != nil {
				status = *req.ImplementationStatus
				if mapped, ok := coredata.InternalControlStateForImplementationStatus(status); ok {
					state = mapped
				}
			}

			operatingMode, operatingInterval, operatingEvent := req.OperatingFrequency.Columns()

			internalControl = &coredata.InternalControl{
				ID:                   gid.New(organization.ID.TenantID(), coredata.InternalControlEntityType),
				OrganizationID:       organization.ID,
				Name:                 req.Name,
				Description:          req.Description,
				Category:             req.Category,
				ReferenceID:          "custom-internal-control-" + referenceID.String(),
				Code:                 req.Code,
				ControlType:          req.ControlType,
				Nature:               req.Nature,
				OperatingMode:        operatingMode,
				OperatingInterval:    operatingInterval,
				OperatingEvent:       operatingEvent,
				EvidenceCadence:      req.EvidenceCadence,
				TestingCadence:       req.TestingCadence,
				NextEvidenceDue:      coredata.NextInternalControlDue(now, req.EvidenceCadence),
				NextTestDue:          coredata.NextInternalControlDue(now, req.TestingCadence),
				ImplementationStatus: status,
				OwnerID:              req.OwnerID,
				ReviewerID:           req.ReviewerID,
				State:                state,
				CreatedAt:            now,
				UpdatedAt:            now,
			}

			if err := internalControl.Insert(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot insert internalControl: %w", err)
			}

			if err := insertInternalControlEvent(
				ctx,
				conn,
				scope,
				internalControl,
				coredata.InternalControlEventTypeCreated,
				now,
			); err != nil {
				return fmt.Errorf("cannot record internal control event: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return internalControl, nil
}

func (s InternalControlService) Delete(
	ctx context.Context, scope coredata.Scoper,
	internalControlID gid.GID,
) error {
	return s.svc.pg.WithTx(ctx, func(ctx context.Context, conn pg.Tx) error {
		internalControl := &coredata.InternalControl{}
		if err := internalControl.LoadByID(ctx, conn, scope, internalControlID); err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return nil
			}

			return fmt.Errorf("cannot load internalControl: %w", err)
		}

		now := time.Now()
		if err := insertInternalControlEvent(
			ctx,
			conn,
			scope,
			internalControl,
			coredata.InternalControlEventTypeDeleted,
			now,
		); err != nil {
			return fmt.Errorf("cannot record internal control event: %w", err)
		}

		if err := internalControl.Delete(ctx, conn, scope, internalControlID); err != nil {
			return fmt.Errorf("cannot delete internalControl: %w", err)
		}

		return nil
	})
}

func (s InternalControlService) CreateDocumentMapping(
	ctx context.Context, scope coredata.Scoper,
	internalControlID gid.GID,
	documentID gid.GID,
) (*coredata.InternalControl, *coredata.Document, error) {
	internalControl := &coredata.InternalControl{}
	document := &coredata.Document{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := internalControl.LoadByID(ctx, tx, scope, internalControlID); err != nil {
				return fmt.Errorf("cannot load internalControl: %w", err)
			}

			if err := document.LoadByID(ctx, tx, scope, documentID); err != nil {
				return fmt.Errorf("cannot load document: %w", err)
			}

			internalControlDocument := &coredata.InternalControlDocument{
				InternalControlID: internalControl.ID,
				DocumentID:        document.ID,
				OrganizationID:    internalControl.OrganizationID,
				TenantID:          scope.GetTenantID(),
				CreatedAt:         time.Now(),
			}

			if err := internalControlDocument.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert internal control document: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return internalControl, document, nil
}

func (s InternalControlService) CountForThirdPartyID(
	ctx context.Context, scope coredata.Scoper,
	thirdPartyID gid.GID,
	filter *coredata.InternalControlFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			internalControls := &coredata.InternalControls{}

			count, err = internalControls.CountByThirdPartyID(ctx, conn, scope, thirdPartyID, filter)
			if err != nil {
				return fmt.Errorf("cannot count internalControls: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s InternalControlService) ListForThirdPartyID(
	ctx context.Context, scope coredata.Scoper,
	thirdPartyID gid.GID,
	cursor *page.Cursor[coredata.InternalControlOrderField],
	filter *coredata.InternalControlFilter,
) (*page.Page[*coredata.InternalControl, coredata.InternalControlOrderField], error) {
	var internalControls coredata.InternalControls

	thirdParty := &coredata.ThirdParty{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := thirdParty.LoadByID(ctx, conn, scope, thirdPartyID); err != nil {
				return fmt.Errorf("cannot load third party: %w", err)
			}

			err := internalControls.LoadByThirdPartyID(ctx, conn, scope, thirdParty.ID, cursor, filter)
			if err != nil {
				return fmt.Errorf("cannot load internalControls: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(internalControls, cursor), nil
}

func (s InternalControlService) CountForTaskID(
	ctx context.Context, scope coredata.Scoper,
	taskID gid.GID,
	filter *coredata.InternalControlFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			internalControls := &coredata.InternalControls{}

			count, err = internalControls.CountByTaskID(ctx, conn, scope, taskID, filter)
			if err != nil {
				return fmt.Errorf("cannot count internal controls: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s InternalControlService) ListForTaskID(
	ctx context.Context, scope coredata.Scoper,
	taskID gid.GID,
	cursor *page.Cursor[coredata.InternalControlOrderField],
	filter *coredata.InternalControlFilter,
) (*page.Page[*coredata.InternalControl, coredata.InternalControlOrderField], error) {
	var internalControls coredata.InternalControls

	task := &coredata.Task{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := task.LoadByID(ctx, conn, scope, taskID); err != nil {
				return fmt.Errorf("cannot load task: %w", err)
			}

			err := internalControls.LoadByTaskID(ctx, conn, scope, task.ID, cursor, filter)
			if err != nil {
				return fmt.Errorf("cannot load internal controls: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(internalControls, cursor), nil
}

func (s InternalControlService) CreateThirdPartyMapping(
	ctx context.Context, scope coredata.Scoper,
	internalControlID gid.GID,
	thirdPartyID gid.GID,
) (*coredata.InternalControl, *coredata.ThirdParty, error) {
	internalControl := &coredata.InternalControl{}
	thirdParty := &coredata.ThirdParty{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := internalControl.LoadByID(ctx, tx, scope, internalControlID); err != nil {
				return fmt.Errorf("cannot load internalControl: %w", err)
			}

			if err := thirdParty.LoadByID(ctx, tx, scope, thirdPartyID); err != nil {
				return fmt.Errorf("cannot load third party: %w", err)
			}

			internalControlThirdParty := &coredata.InternalControlThirdParty{
				InternalControlID: internalControl.ID,
				ThirdPartyID:      thirdParty.ID,
				CreatedAt:         time.Now(),
			}

			if err := internalControlThirdParty.Upsert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot upsert internal control third party: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return internalControl, thirdParty, nil
}

func (s InternalControlService) DeleteThirdPartyMapping(
	ctx context.Context, scope coredata.Scoper,
	internalControlID gid.GID,
	thirdPartyID gid.GID,
) (*coredata.InternalControl, *coredata.ThirdParty, error) {
	internalControl := &coredata.InternalControl{}
	thirdParty := &coredata.ThirdParty{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := internalControl.LoadByID(ctx, tx, scope, internalControlID); err != nil {
				return fmt.Errorf("cannot load internalControl: %w", err)
			}

			if err := thirdParty.LoadByID(ctx, tx, scope, thirdPartyID); err != nil {
				return fmt.Errorf("cannot load third party: %w", err)
			}

			internalControlThirdParty := &coredata.InternalControlThirdParty{}
			if err := internalControlThirdParty.Delete(ctx, tx, scope, internalControl.ID, thirdParty.ID); err != nil {
				return fmt.Errorf("cannot delete internal control third party mapping: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return internalControl, thirdParty, nil
}

func (s InternalControlService) DeleteDocumentMapping(
	ctx context.Context, scope coredata.Scoper,
	internalControlID gid.GID,
	documentID gid.GID,
) (*coredata.InternalControl, *coredata.Document, error) {
	internalControl := &coredata.InternalControl{}
	document := &coredata.Document{ID: documentID}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := internalControl.LoadByID(ctx, tx, scope, internalControlID); err != nil {
				return fmt.Errorf("cannot load internalControl: %w", err)
			}

			if err := document.LoadByID(ctx, tx, scope, documentID); err != nil {
				if !errors.Is(err, coredata.ErrResourceNotFound) {
					return fmt.Errorf("cannot load document: %w", err)
				}
			}

			internalControlDocument := &coredata.InternalControlDocument{}
			if err := internalControlDocument.Delete(ctx, tx, scope, internalControl.ID, documentID); err != nil {
				return fmt.Errorf("cannot delete internal control document mapping: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return internalControl, document, nil
}

func insertInternalControlEvent(
	ctx context.Context,
	conn pg.Tx,
	scope coredata.Scoper,
	internalControl *coredata.InternalControl,
	eventType coredata.InternalControlEventType,
	now time.Time,
) error {
	event := coredata.NewInternalControlEvent(internalControl, eventType, now)
	if err := event.Insert(ctx, conn, scope); err != nil {
		return fmt.Errorf("cannot insert internal control event: %w", err)
	}

	return nil
}

type internalControlPeople struct {
	ownerID    *gid.GID
	reviewerID *gid.GID
}

func normalizeInternalControlCode(code *string) *string {
	if code == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*code)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}

func normalizeOmittableInternalControlCode(code **string) {
	if code == nil || *code == nil {
		return
	}

	*code = normalizeInternalControlCode(*code)
}

func distinctInternalControlPeople() validator.ValidatorFunc {
	return func(value any) *validator.ValidationError {
		pair, ok := value.(internalControlPeople)
		if !ok || pair.ownerID == nil || pair.reviewerID == nil || *pair.ownerID != *pair.reviewerID {
			return nil
		}

		return &validator.ValidationError{
			Code:    validator.ErrorCodeCustom,
			Message: "must be a different person from the owner",
		}
	}
}

func distinctInternalControlPeopleError(ownerID, reviewerID *gid.GID) error {
	validationError := distinctInternalControlPeople()(internalControlPeople{ownerID: ownerID, reviewerID: reviewerID})
	if validationError == nil {
		return nil
	}

	validationError.Field = "reviewer_id"

	return validator.ValidationErrors{validationError}
}

func normalizeNonPositiveCadence(cadence *timespan.TimeSpan) *timespan.TimeSpan {
	if cadence == nil || !cadenceAdvances(*cadence) {
		return nil
	}

	return cadence
}

func normalizeOmittableCadence(cadence **timespan.TimeSpan) {
	if cadence == nil || *cadence == nil {
		return
	}

	*cadence = normalizeNonPositiveCadence(*cadence)
}

func cadenceAdvances(span timespan.TimeSpan) bool {
	if span.IsZero() {
		return false
	}

	anchor := time.Unix(0, 0).UTC()

	return span.AddTo(anchor).After(anchor)
}

func setOmittable[T any](dest **T, src **T) {
	if src == nil {
		return
	}

	*dest = *src
}

func assignCadence(
	current **timespan.TimeSpan,
	due **time.Time,
	next **timespan.TimeSpan,
	now time.Time,
) {
	if next == nil {
		return
	}

	if *next == nil {
		*current = nil

		if due != nil {
			*due = nil
		}

		return
	}

	changed := *current == nil || **current != **next
	*current = *next

	if changed && due != nil {
		*due = coredata.NextInternalControlDue(now, *next)
	}
}

func applyInternalControlUpdate(req *UpdateInternalControlRequest, internalControl *coredata.InternalControl, now time.Time) {
	setOmittable(&internalControl.Code, req.Code)
	setOmittable(&internalControl.ControlType, req.ControlType)
	setOmittable(&internalControl.Nature, req.Nature)
	assignOperatingFrequency(internalControl, req.OperatingFrequency)
	assignCadence(&internalControl.EvidenceCadence, &internalControl.NextEvidenceDue, req.EvidenceCadence, now)
	assignCadence(&internalControl.TestingCadence, &internalControl.NextTestDue, req.TestingCadence, now)
	setOmittable(&internalControl.OwnerID, req.OwnerID)
	setOmittable(&internalControl.ReviewerID, req.ReviewerID)

	// Resending the current status must not rewrite a legacy state such as
	// NOT_STARTED. A real status change wins over a state sent in the same
	// request, because operating controls share the implemented state.
	if req.ImplementationStatus != nil && internalControl.ImplementationStatus != *req.ImplementationStatus {
		internalControl.ImplementationStatus = *req.ImplementationStatus
		if state, ok := coredata.InternalControlStateForImplementationStatus(*req.ImplementationStatus); ok {
			internalControl.State = state
		}

		return
	}

	if req.State != nil && internalControl.State != *req.State {
		internalControl.State = *req.State
		internalControl.ImplementationStatus = coredata.ImplementationStatusForInternalControlState(*req.State)
	}
}

func normalizeOperatingFrequency(freq *coredata.InternalControlOperatingFrequency) {
	if freq == nil {
		return
	}

	freq.Event = normalizeOperatingEvent(freq.Event)
}

func normalizeOperatingEvent(event *string) *string {
	if event == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*event)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}

func assignOperatingFrequency(internalControl *coredata.InternalControl, next **coredata.InternalControlOperatingFrequency) {
	if next == nil {
		return
	}

	mode, interval, event := (*next).Columns()
	internalControl.OperatingMode = mode
	internalControl.OperatingInterval = interval
	internalControl.OperatingEvent = event
}

func validOperatingFrequency() validator.ValidatorFunc {
	return func(value any) *validator.ValidationError {
		freq, ok := value.(coredata.InternalControlOperatingFrequency)
		if !ok {
			return nil
		}

		if !freq.Mode.IsValid() {
			return &validator.ValidationError{
				Code:    validator.ErrorCodeInvalidEnum,
				Message: "must be CONTINUOUS, EVENT, or PERIODIC",
			}
		}

		if freq.Event != nil {
			if err := validator.SafeTextNoNewLine(TitleMaxLength)(*freq.Event); err != nil {
				return err
			}
		}

		switch freq.Mode {
		case coredata.InternalControlOperatingModeContinuous:
			if freq.Interval != nil || freq.Event != nil {
				return &validator.ValidationError{
					Code:    validator.ErrorCodeCustom,
					Message: "continuous frequency has no interval or event",
				}
			}
		case coredata.InternalControlOperatingModeEvent:
			if freq.Interval != nil {
				return &validator.ValidationError{
					Code:    validator.ErrorCodeCustom,
					Message: "event frequency has no interval",
				}
			}
		case coredata.InternalControlOperatingModePeriodic:
			if freq.Event != nil {
				return &validator.ValidationError{
					Code:    validator.ErrorCodeCustom,
					Message: "periodic frequency has no event",
				}
			}

			if freq.Interval == nil || freq.Interval.IsZero() {
				return &validator.ValidationError{
					Code:    validator.ErrorCodeRequired,
					Message: "periodic frequency requires a duration",
				}
			}

			if err := positiveTimeSpan()(*freq.Interval); err != nil {
				return err
			}
		}

		return nil
	}
}

func positiveTimeSpan() validator.ValidatorFunc {
	return func(value any) *validator.ValidationError {
		span, ok := value.(timespan.TimeSpan)
		if !ok || span.IsZero() || cadenceAdvances(span) {
			return nil
		}

		return &validator.ValidationError{
			Code:    validator.ErrorCodeCustom,
			Message: "must be a positive duration",
		}
	}
}

func loadInternalControlProfile(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	profileID gid.GID,
	organizationID gid.GID,
) error {
	profile := &coredata.MembershipProfile{}
	if err := profile.LoadByID(ctx, conn, scope, profileID); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return coredata.ErrResourceNotFound
		}

		return fmt.Errorf("cannot load membership profile: %w", err)
	}

	if profile.OrganizationID != organizationID {
		return coredata.ErrResourceNotFound
	}

	return nil
}
