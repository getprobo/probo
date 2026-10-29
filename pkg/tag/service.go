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

package tag

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

const (
	keyMaxLength   = 100
	valueMaxLength = 200
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
		// Color is left unchanged when nil and cleared when it points to nil
		// or to an empty string.
		Color **string
	}

	AttachRequest struct {
		ResourceID gid.GID
		TagID      gid.GID
	}

	DetachRequest struct {
		ResourceID gid.GID
		TagID      gid.GID
	}

	AttachResult struct {
		Tag        *coredata.Tag
		Assignment *coredata.TagAssignment
		// ReplacedAssignmentID is set when the resource already carried another
		// value of the same key, which this attach removed.
		ReplacedAssignmentID *gid.GID
	}

	DetachResult struct {
		Tag *coredata.Tag
		// DeletedAssignmentID is nil when the tag was not attached.
		DeletedAssignmentID *gid.GID
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

	v.Check(req.ID, "id", validator.Required(), validator.GID(coredata.TagEntityType))
	v.Check(req.Key, "key", validator.NotEmpty(), validator.Slug(keyMaxLength))
	v.Check(req.Value, "value", validator.NotEmpty(), validator.SafeText(valueMaxLength))
	v.Check(req.Color, "color", validator.HexColor())

	return v.Error()
}

func (req *AttachRequest) Validate() error {
	v := validator.New()

	v.Check(req.ResourceID, "resource_id", validator.Required(), validator.GID(TaggableEntityTypes...))
	v.Check(req.TagID, "tag_id", validator.Required(), validator.GID(coredata.TagEntityType))

	return v.Error()
}

func (req *DetachRequest) Validate() error {
	v := validator.New()

	v.Check(req.ResourceID, "resource_id", validator.Required(), validator.GID(TaggableEntityTypes...))
	v.Check(req.TagID, "tag_id", validator.Required(), validator.GID(coredata.TagEntityType))

	return v.Error()
}

func (s *Service) Get(
	ctx context.Context,
	scope coredata.Scoper,
	tagID gid.GID,
) (*coredata.Tag, error) {
	tag := &coredata.Tag{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := tag.LoadByID(ctx, conn, scope, tagID); err != nil {
				return fmt.Errorf("cannot load tag: %w", err)
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
	cursor *page.Cursor[coredata.TagOrderField],
) (*page.Page[*coredata.Tag, coredata.TagOrderField], error) {
	var tags coredata.Tags

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			if err := tags.LoadByOrganizationID(ctx, conn, scope, organization.ID, cursor); err != nil {
				return fmt.Errorf("cannot load tags: %w", err)
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
			tags := &coredata.Tags{}

			var err error

			count, err = tags.CountByOrganizationID(ctx, conn, scope, organizationID)
			if err != nil {
				return fmt.Errorf("cannot count tags: %w", err)
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
) (*coredata.Tag, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()

	color := req.Color
	if color != nil && *color == "" {
		color = nil
	}

	tag := &coredata.Tag{
		ID:             gid.New(scope.GetTenantID(), coredata.TagEntityType),
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
				return fmt.Errorf("cannot create tag: %w", err)
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
) (*coredata.Tag, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	tag := &coredata.Tag{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := tag.LoadByID(ctx, conn, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load tag: %w", err)
			}

			if req.Key != nil {
				tag.Key = *req.Key
			}

			if req.Value != nil {
				tag.Value = *req.Value
			}

			if req.Color != nil {
				color := *req.Color
				if color != nil && *color == "" {
					color = nil
				}

				tag.Color = color
			}

			tag.UpdatedAt = time.Now()

			if err := tag.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update tag: %w", err)
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
			tag := &coredata.Tag{}
			if err := tag.LoadByID(ctx, conn, scope, tagID); err != nil {
				return fmt.Errorf("cannot load tag: %w", err)
			}

			if err := tag.Delete(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot delete tag: %w", err)
			}

			return nil
		},
	)
}

// Attach puts the tag on the resource. A resource holds one value per key, so
// attaching env=staging replaces env=prod. Attaching a tag that is already
// present returns the existing assignment.
func (s *Service) Attach(
	ctx context.Context,
	scope coredata.Scoper,
	req AttachRequest,
) (*AttachResult, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	result := &AttachResult{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			tag := &coredata.Tag{}
			if err := tag.LoadByID(ctx, tx, scope, req.TagID); err != nil {
				return fmt.Errorf("cannot load tag: %w", err)
			}

			result.Tag = tag

			existing := &coredata.TagAssignment{}

			err := existing.LoadByResourceIDAndKey(ctx, tx, scope, req.ResourceID, tag.Key)
			switch {
			case err == nil && existing.TagID == tag.ID:
				result.Assignment = existing

				return nil
			case err == nil:
				if err := existing.Delete(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot remove previous tag value: %w", err)
				}

				result.ReplacedAssignmentID = &existing.ID
			case !errors.Is(err, coredata.ErrResourceNotFound):
				return fmt.Errorf("cannot load tag assignment: %w", err)
			}

			assignment := &coredata.TagAssignment{
				ID:             gid.New(scope.GetTenantID(), coredata.TagAssignmentEntityType),
				OrganizationID: tag.OrganizationID,
				ResourceID:     req.ResourceID,
				TagID:          tag.ID,
				Key:            tag.Key,
				CreatedAt:      time.Now(),
			}

			if err := assignment.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot attach tag: %w", err)
			}

			result.Assignment = assignment

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Detach removes the tag from the resource. Detaching a tag that is not
// attached succeeds and reports no deleted assignment.
func (s *Service) Detach(
	ctx context.Context,
	scope coredata.Scoper,
	req DetachRequest,
) (*DetachResult, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	result := &DetachResult{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			tag := &coredata.Tag{}
			if err := tag.LoadByID(ctx, tx, scope, req.TagID); err != nil {
				return fmt.Errorf("cannot load tag: %w", err)
			}

			result.Tag = tag

			assignment := &coredata.TagAssignment{}

			err := assignment.DeleteByResourceIDAndTagID(ctx, tx, scope, req.ResourceID, req.TagID)
			switch {
			case err == nil:
				result.DeletedAssignmentID = &assignment.ID
			case !errors.Is(err, coredata.ErrResourceNotFound):
				return fmt.Errorf("cannot detach tag: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Service) ListForResourceID(
	ctx context.Context,
	scope coredata.Scoper,
	resourceID gid.GID,
) ([]*coredata.Tag, error) {
	var tags coredata.Tags

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := tags.LoadByResourceID(ctx, conn, scope, resourceID); err != nil {
				return fmt.Errorf("cannot load tags for resource: %w", err)
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
) (map[gid.GID][]*coredata.Tag, error) {
	result := make(map[gid.GID][]*coredata.Tag, len(resourceIDs))
	if len(resourceIDs) == 0 {
		return result, nil
	}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var assignments coredata.TagAssignments
			if err := assignments.LoadByResourceIDs(ctx, conn, scope, resourceIDs); err != nil {
				return fmt.Errorf("cannot load tag assignments: %w", err)
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

			var tags coredata.Tags
			if err := tags.LoadByIDs(ctx, conn, scope, tagIDs); err != nil {
				return fmt.Errorf("cannot load tags: %w", err)
			}

			tagsByID := make(map[gid.GID]*coredata.Tag, len(tags))
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
			assignments := &coredata.TagAssignments{}

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

// ListAssignmentsForTagID returns a page of assignments for the given tag.
func (s *Service) ListAssignmentsForTagID(
	ctx context.Context,
	scope coredata.Scoper,
	tagID gid.GID,
	cursor *page.Cursor[coredata.TagAssignmentOrderField],
) (*page.Page[*coredata.TagAssignment, coredata.TagAssignmentOrderField], error) {
	var assignments coredata.TagAssignments

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			tag := &coredata.Tag{}
			if err := tag.LoadByID(ctx, conn, scope, tagID); err != nil {
				return fmt.Errorf("cannot load tag: %w", err)
			}

			if err := assignments.LoadByTagID(ctx, conn, scope, tagID, cursor); err != nil {
				return fmt.Errorf("cannot list tag assignments: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(assignments, cursor), nil
}

// CountAssignmentsForTagID returns how many resources are assigned the tag.
func (s *Service) CountAssignmentsForTagID(
	ctx context.Context,
	scope coredata.Scoper,
	tagID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			assignments := &coredata.TagAssignments{}

			var err error

			count, err = assignments.CountByTagID(ctx, conn, scope, tagID)
			if err != nil {
				return fmt.Errorf("cannot count tag assignments: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}
