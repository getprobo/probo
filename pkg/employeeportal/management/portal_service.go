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

package management

import (
	"context"
	"fmt"
	"mime"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filevalidation"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

type (
	CreateEmployeePortalRequest struct {
		OrganizationID gid.GID
		Name           string
	}

	UpdateRequest struct {
		ID           gid.GID
		Name         *string
		Active       *bool
		Capabilities *coredata.EmployeePortalCapabilitiesPatch
	}

	UpdateBrandRequest struct {
		EmployeePortalID gid.GID
		LogoFile         **FileUpload
		DarkLogoFile     **FileUpload
	}
)

func (r *CreateEmployeePortalRequest) Validate() error {
	v := validator.New()

	v.Check(r.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(r.Name, "name", validator.Required(), validator.SafeTextNoNewLine(NameMaxLength))

	return v.Error()
}

func (r *UpdateRequest) Validate() error {
	v := validator.New()

	v.Check(r.ID, "id", validator.Required(), validator.GID(coredata.EmployeePortalEntityType))

	if r.Name != nil {
		v.Check(*r.Name, "name", validator.Required(), validator.SafeTextNoNewLine(NameMaxLength))
	}

	return v.Error()
}

func (req *UpdateBrandRequest) Validate() error {
	v := validator.New()

	v.Check(req.EmployeePortalID, "employee_portal_id", validator.Required(), validator.GID(coredata.EmployeePortalEntityType))

	if err := v.Error(); err != nil {
		return err
	}

	fv := filevalidation.NewValidator(
		filevalidation.WithCategories(filevalidation.CategoryImage),
		filevalidation.WithMaxFileSize(maxBrandFileSize),
	)

	if req.LogoFile != nil && *req.LogoFile != nil {
		logoFile := *req.LogoFile
		if err := fv.Validate(logoFile.Filename, logoFile.ContentType, logoFile.Size); err != nil {
			return fmt.Errorf("invalid logo file: %w", err)
		}
	}

	if req.DarkLogoFile != nil && *req.DarkLogoFile != nil {
		darkLogoFile := *req.DarkLogoFile
		if err := fv.Validate(darkLogoFile.Filename, darkLogoFile.ContentType, darkLogoFile.Size); err != nil {
			return fmt.Errorf("invalid dark logo file: %w", err)
		}
	}

	return nil
}

func (s *Service) OldestIDForOrganization(
	ctx context.Context,
	organizationID gid.GID,
) (gid.GID, error) {
	var portalID gid.GID

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			scope := coredata.NewScopeFromObjectID(organizationID)
			portal := &coredata.EmployeePortal{}
			if err := portal.LoadOldestByOrganizationID(ctx, conn, scope, organizationID); err != nil {
				return fmt.Errorf("cannot load employee portal: %w", err)
			}

			portalID = portal.ID

			return nil
		},
	)
	if err != nil {
		return gid.Nil, err
	}

	return portalID, nil
}

func (s *Service) Get(
	ctx context.Context,
	scope coredata.Scoper,
	portalID gid.GID,
) (*coredata.EmployeePortal, error) {
	var portal *coredata.EmployeePortal

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			portal = &coredata.EmployeePortal{}
			if err := portal.LoadByID(ctx, conn, scope, portalID); err != nil {
				return fmt.Errorf("cannot load employee portal: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load employee portal: %w", err)
	}

	return portal, nil
}

func (s *Service) Create(
	ctx context.Context,
	scope coredata.Scoper,
	req *CreateEmployeePortalRequest,
) (*coredata.EmployeePortal, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	var portal *coredata.EmployeePortal

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, tx, scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			now := time.Now()

			portal = &coredata.EmployeePortal{
				ID:             gid.New(scope.GetTenantID(), coredata.EmployeePortalEntityType),
				OrganizationID: organization.ID,
				Name:           req.Name,
				Active:         true,
				Capabilities:   coredata.DefaultEmployeePortalCapabilities(),
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			if err := portal.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert employee portal: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return portal, nil
}

func (s *Service) ListForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.EmployeePortalOrderField],
) (*page.Page[*coredata.EmployeePortal, coredata.EmployeePortalOrderField], error) {
	var portals coredata.EmployeePortals

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := portals.LoadByOrganizationID(ctx, conn, scope, organizationID, cursor); err != nil {
				return fmt.Errorf("cannot list employee portals: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(portals, cursor), nil
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
			var err error

			count, err = (&coredata.EmployeePortals{}).CountByOrganizationID(
				ctx,
				conn,
				scope,
				organizationID,
			)
			if err != nil {
				return fmt.Errorf("cannot count employee portals: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) Update(
	ctx context.Context,
	scope coredata.Scoper,
	req *UpdateRequest,
) (*coredata.EmployeePortal, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	var portal *coredata.EmployeePortal

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			portal = &coredata.EmployeePortal{}
			if err := portal.LoadByID(ctx, conn, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load employee portal: %w", err)
			}

			if req.Name != nil {
				portal.Name = *req.Name
			}

			if req.Active != nil {
				portal.Active = *req.Active
			}

			if req.Capabilities != nil {
				portal.Capabilities = req.Capabilities.Apply(portal.Capabilities)
			}

			portal.UpdatedAt = time.Now()

			if err := portal.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update employee portal: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return portal, nil
}

func (s *Service) UpdateBrand(
	ctx context.Context,
	scope coredata.Scoper,
	req *UpdateBrandRequest,
) (*coredata.EmployeePortal, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	var portal *coredata.EmployeePortal

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			portal = &coredata.EmployeePortal{}
			if err := portal.LoadByID(ctx, conn, scope, req.EmployeePortalID); err != nil {
				return fmt.Errorf("cannot load employee portal: %w", err)
			}

			now := time.Now()

			if req.LogoFile != nil {
				if *req.LogoFile == nil {
					portal.LogoFileID = nil
				} else {
					file, err := s.uploadBrandFile(ctx, scope, conn, *req.LogoFile, "employee-portal-logo", portal)
					if err != nil {
						return fmt.Errorf("cannot upload logo file: %w", err)
					}

					portal.LogoFileID = &file.ID
				}
			}

			if req.DarkLogoFile != nil {
				if *req.DarkLogoFile == nil {
					portal.DarkLogoFileID = nil
				} else {
					file, err := s.uploadBrandFile(ctx, scope, conn, *req.DarkLogoFile, "employee-portal-dark-logo", portal)
					if err != nil {
						return fmt.Errorf("cannot upload dark logo file: %w", err)
					}

					portal.DarkLogoFileID = &file.ID
				}
			}

			portal.UpdatedAt = now

			if err := portal.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update employee portal: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return portal, nil
}

func (s *Service) uploadBrandFile(
	ctx context.Context,
	scope coredata.Scoper,
	conn pg.Tx,
	fileUpload *FileUpload,
	fileType string,
	portal *coredata.EmployeePortal,
) (*coredata.File, error) {
	objectKey, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("cannot generate object key: %w", err)
	}

	mimeType := fileUpload.ContentType
	if mimeType == "" {
		mimeType = mime.TypeByExtension(filepath.Ext(fileUpload.Filename))
	}

	_, err = s.s3.PutObject(
		ctx,
		&s3.PutObjectInput{
			Bucket:       &s.bucket,
			Key:          new(objectKey.String()),
			Body:         fileUpload.Content,
			ContentType:  &mimeType,
			CacheControl: new("max-age=3600, public"),
			Metadata: map[string]string{
				"type":               fileType,
				"employee-portal-id": portal.ID.String(),
				"organization-id":    portal.OrganizationID.String(),
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot upload file to S3: %w", err)
	}

	headOutput, err := s.s3.HeadObject(
		ctx,
		&s3.HeadObjectInput{
			Bucket: new(s.bucket),
			Key:    new(objectKey.String()),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot get object metadata: %w", err)
	}

	now := time.Now()
	fileID := gid.New(scope.GetTenantID(), coredata.FileEntityType)

	file := &coredata.File{
		ID:             fileID,
		OrganizationID: portal.OrganizationID,
		BucketName:     s.bucket,
		MimeType:       mimeType,
		FileName:       fileUpload.Filename,
		FileKey:        objectKey.String(),
		FileSize:       *headOutput.ContentLength,
		Visibility:     coredata.FileVisibilityPublic,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := file.Insert(ctx, conn, scope); err != nil {
		return nil, fmt.Errorf("cannot insert file: %w", err)
	}

	return file, nil
}

func (s *Service) Delete(
	ctx context.Context,
	scope coredata.Scoper,
	portalID gid.GID,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			portal := &coredata.EmployeePortal{}
			if err := portal.LoadByID(ctx, tx, scope, portalID); err != nil {
				return fmt.Errorf("cannot load employee portal: %w", err)
			}

			if err := portal.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot delete employee portal: %w", err)
			}

			return nil
		},
	)
}
