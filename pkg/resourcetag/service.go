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

package resourcetag

import (
	"context"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

const (
	keyMaxLength    = 100
	valueMaxLength  = 200
	resourceIDLimit = 1000
)

type (
	Service struct {
		pg *pg.Client
	}

	CreateTagRequest struct {
		OrganizationID gid.GID
		Key            string
		Value          string
		Color          *string
	}

	UpdateTagRequest struct {
		ID    gid.GID
		Key   *string
		Value *string
		Color *string
	}

	AttachRequest struct {
		ResourceID gid.GID
		TagID      gid.GID
	}

	DetachRequest struct {
		ResourceID gid.GID
		TagID      gid.GID
	}
)

func NewService(pgClient *pg.Client) *Service {
	return &Service{
		pg: pgClient,
	}
}

func (req *CreateTagRequest) Validate() error {
	v := validator.New()

	v.Check(req.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(req.Key, "key", validator.Required(), validator.Slug(keyMaxLength))
	v.Check(req.Value, "value", validator.Required(), validator.SafeText(valueMaxLength))
	v.Check(req.Color, "color", validator.HexColor())

	return v.Error()
}

func (req *UpdateTagRequest) Validate() error {
	v := validator.New()

	v.Check(req.ID, "id", validator.Required(), validator.GID(coredata.ResourceTagEntityType))
	v.Check(req.Key, "key", validator.NotEmpty(), validator.Slug(keyMaxLength))
	v.Check(req.Value, "value", validator.NotEmpty(), validator.SafeText(valueMaxLength))
	v.Check(req.Color, "color", validator.HexColor())

	return v.Error()
}

func (req *AttachRequest) Validate() error {
	v := validator.New()

	v.Check(req.ResourceID, "resource_id", validator.Required(), validator.GID())
	v.Check(req.TagID, "tag_id", validator.Required(), validator.GID(coredata.ResourceTagEntityType))

	return v.Error()
}

func (req *DetachRequest) Validate() error {
	v := validator.New()

	v.Check(req.ResourceID, "resource_id", validator.Required(), validator.GID())
	v.Check(req.TagID, "tag_id", validator.Required(), validator.GID(coredata.ResourceTagEntityType))

	return v.Error()
}

func (s *Service) Get(
	ctx context.Context,
	scope coredata.Scoper,
	tagID gid.GID,
) (*coredata.ResourceTag, error) {
	tag := &coredata.ResourceTag{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := tag.LoadByID(ctx, conn, scope, tagID); err != nil {
				return fmt.Errorf("cannot load resource tag: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return tag, nil
}

func (s *Service) ListForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.ResourceTagOrderField],
) (*page.Page[*coredata.ResourceTag, coredata.ResourceTagOrderField], error) {
	var tags coredata.ResourceTags

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			if err := tags.LoadByOrganizationID(ctx, conn, scope, organization.ID, cursor); err != nil {
				return fmt.Errorf("cannot load resource tags: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(tags, cursor), nil
}

func (s *Service) CountForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			tags := &coredata.ResourceTags{}

			var err error
			count, err = tags.CountByOrganizationID(ctx, conn, scope, organizationID)
			if err != nil {
				return fmt.Errorf("cannot count resource tags: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) Create(
	ctx context.Context,
	scope coredata.Scoper,
	req CreateTagRequest,
) (*coredata.ResourceTag, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()

	color := req.Color
	if color != nil && *color == "" {
		color = nil
	}

	tag := &coredata.ResourceTag{
		ID:             gid.New(scope.GetTenantID(), coredata.ResourceTagEntityType),
		OrganizationID: req.OrganizationID,
		Key:            req.Key,
		Value:          req.Value,
		Color:          color,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			if err := tag.Insert(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot create resource tag: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return tag, nil
}

func (s *Service) Update(
	ctx context.Context,
	scope coredata.Scoper,
	req UpdateTagRequest,
) (*coredata.ResourceTag, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	tag := &coredata.ResourceTag{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := tag.LoadByID(ctx, conn, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load resource tag: %w", err)
			}

			if req.Key != nil {
				tag.Key = *req.Key
			}

			if req.Value != nil {
				tag.Value = *req.Value
			}

			if req.Color != nil {
				if *req.Color == "" {
					tag.Color = nil
				} else {
					tag.Color = req.Color
				}
			}

			tag.UpdatedAt = time.Now()

			if err := tag.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update resource tag: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return tag, nil
}

func (s *Service) Delete(
	ctx context.Context,
	scope coredata.Scoper,
	tagID gid.GID,
) error {
	return s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			tag := &coredata.ResourceTag{}
			if err := tag.LoadByID(ctx, conn, scope, tagID); err != nil {
				return fmt.Errorf("cannot load resource tag: %w", err)
			}

			if err := tag.Delete(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot delete resource tag: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) Attach(
	ctx context.Context,
	scope coredata.Scoper,
	req AttachRequest,
) error {
	if err := req.Validate(); err != nil {
		return err
	}

	return s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			tag := &coredata.ResourceTag{}
			if err := tag.LoadByID(ctx, conn, scope, req.TagID); err != nil {
				return fmt.Errorf("cannot load resource tag: %w", err)
			}

			assignment := &coredata.ResourceTagAssignment{
				ResourceID: req.ResourceID,
				TagID:      req.TagID,
				CreatedAt:  time.Now(),
			}

			if err := assignment.Insert(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot attach resource tag: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) Detach(
	ctx context.Context,
	scope coredata.Scoper,
	req DetachRequest,
) error {
	if err := req.Validate(); err != nil {
		return err
	}

	return s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			assignment := &coredata.ResourceTagAssignment{
				ResourceID: req.ResourceID,
				TagID:      req.TagID,
			}

			if err := assignment.Delete(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot detach resource tag: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) ListForResourceID(
	ctx context.Context,
	scope coredata.Scoper,
	resourceID gid.GID,
) ([]*coredata.ResourceTag, error) {
	var tags coredata.ResourceTags

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := tags.LoadByResourceID(ctx, conn, scope, resourceID); err != nil {
				return fmt.Errorf("cannot load resource tags for resource: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return tags, nil
}

// LoadByResourceIDs returns tags keyed by resource ID for dataloader / field
// resolvers. Ready for the follow-up PR that wires tags onto entity types.
func (s *Service) LoadByResourceIDs(
	ctx context.Context,
	scope coredata.Scoper,
	resourceIDs []gid.GID,
) (map[gid.GID][]*coredata.ResourceTag, error) {
	result := make(map[gid.GID][]*coredata.ResourceTag, len(resourceIDs))
	if len(resourceIDs) == 0 {
		return result, nil
	}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var assignments coredata.ResourceTagAssignments
			if err := assignments.LoadByResourceIDs(ctx, conn, scope, resourceIDs); err != nil {
				return fmt.Errorf("cannot load resource tag assignments: %w", err)
			}

			tagIDs := make([]gid.GID, 0, len(assignments))
			seen := make(map[gid.GID]struct{}, len(assignments))
			for _, assignment := range assignments {
				if _, ok := seen[assignment.TagID]; ok {
					continue
				}

				seen[assignment.TagID] = struct{}{}
				tagIDs = append(tagIDs, assignment.TagID)
			}

			var tags coredata.ResourceTags
			if err := tags.LoadByIDs(ctx, conn, scope, tagIDs); err != nil {
				return fmt.Errorf("cannot load resource tags: %w", err)
			}

			tagsByID := make(map[gid.GID]*coredata.ResourceTag, len(tags))
			for _, tag := range tags {
				tagsByID[tag.ID] = tag
			}

			for _, assignment := range assignments {
				tag, ok := tagsByID[assignment.TagID]
				if !ok {
					continue
				}

				result[assignment.ResourceID] = append(result[assignment.ResourceID], tag)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// FilterResourceIDs returns candidates that have all of the given tags (AND).
func (s *Service) FilterResourceIDs(
	ctx context.Context,
	scope coredata.Scoper,
	candidateIDs []gid.GID,
	tagIDs []gid.GID,
) ([]gid.GID, error) {
	var filtered []gid.GID

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			assignments := &coredata.ResourceTagAssignments{}

			var err error
			filtered, err = assignments.FilterResourceIDs(ctx, conn, scope, candidateIDs, tagIDs)
			if err != nil {
				return fmt.Errorf("cannot filter resource ids by tags: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return filtered, nil
}

// ListResourceIDsByTagID returns up to resourceIDLimit resource IDs that have
// the given tag. Prefer FilterResourceIDs inside entity list queries once tags
// are wired onto specific resources.
func (s *Service) ListResourceIDsByTagID(
	ctx context.Context,
	scope coredata.Scoper,
	tagID gid.GID,
) ([]gid.GID, error) {
	var ids []gid.GID

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			tag := &coredata.ResourceTag{}
			if err := tag.LoadByID(ctx, conn, scope, tagID); err != nil {
				return fmt.Errorf("cannot load resource tag: %w", err)
			}

			assignments := &coredata.ResourceTagAssignments{}

			var err error
			ids, err = assignments.LoadResourceIDsByTagID(ctx, conn, scope, tagID, resourceIDLimit)
			if err != nil {
				return fmt.Errorf("cannot list resource ids by tag: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return ids, nil
}
