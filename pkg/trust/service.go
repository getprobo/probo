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

package trust

import (
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"go.gearno.de/kit/pg"
)

type (
	Service struct {
		pg     *pg.Client
		bucket string
	}

	TenantService struct {
		pg           *pg.Client
		bucket       string
		scope        coredata.Scoper
		TrustCenters *TrustCenterService
	}
)

func NewService(
	pgClient *pg.Client,
	bucket string,
) *Service {
	return &Service{
		pg:     pgClient,
		bucket: bucket,
	}
}

func (s *Service) WithTenant(tenantID gid.TenantID) *TenantService {
	tenantService := &TenantService{
		pg:     s.pg,
		bucket: s.bucket,
		scope:  coredata.NewScope(tenantID),
	}

	tenantService.TrustCenters = &TrustCenterService{svc: tenantService}

	return tenantService
}
