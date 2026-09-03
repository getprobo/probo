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

package drivers

import (
	"context"
	"fmt"
	"strings"

	"go.gearno.de/kit/log"
	cloudgcp "go.probo.inc/probo/pkg/cloud/gcp"
	cloudresourcemanager "google.golang.org/api/cloudresourcemanager/v1"
)

var gcpServiceAccountEmailSuffixes = []string{
	".s3ns.iam.gserviceaccount.com",
	".iam.gserviceaccount.com",
}

// gcpNameResolver names the connected project for the source-name worker.
// The official display name is preferred, then the connected project's
// project ID, then the project ID from the impersonated service-account
// email, then the project number so two sources in one organization still
// differ.
type gcpNameResolver struct {
	session             *cloudgcp.Session
	serviceAccountEmail string
	logger              *log.Logger
}

func NewGCPNameResolver(
	session *cloudgcp.Session,
	serviceAccountEmail string,
	logger *log.Logger,
) NameResolver {
	return &gcpNameResolver{
		session:             session,
		serviceAccountEmail: serviceAccountEmail,
		logger:              logger,
	}
}

func (r *gcpNameResolver) ResolveInstanceName(ctx context.Context) (string, error) {
	project, err := r.getProject(ctx)
	if err != nil {
		if ctx.Err() != nil {
			return "", err
		}

		r.logger.WarnCtx(
			ctx,
			"cannot read gcp project name, trying project id",
			cloudgcp.SafeLogFields(err)...,
		)
	}

	if project != nil {
		if name := strings.TrimSpace(project.Name); name != "" {
			return name, nil
		}

		if id := strings.TrimSpace(project.ProjectId); id != "" {
			return id, nil
		}
	}

	if id := projectIDFromServiceAccountEmail(r.serviceAccountEmail); id != "" {
		return id, nil
	}

	return r.session.AccountID(), nil
}

func (r *gcpNameResolver) getProject(ctx context.Context) (*cloudresourcemanager.Project, error) {
	svc, err := cloudresourcemanager.NewService(ctx, r.session.ServiceOptions()...)
	if err != nil {
		return nil, fmt.Errorf("cannot create gcp resource manager client: %w", err)
	}

	project, err := svc.Projects.Get(r.session.AccountID()).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("cannot get gcp project: %w", err)
	}

	return project, nil
}

func projectIDFromServiceAccountEmail(email string) string {
	_, host, ok := strings.Cut(strings.TrimSpace(email), "@")
	if !ok {
		return ""
	}

	for _, suffix := range gcpServiceAccountEmailSuffixes {
		if before, ok0 := strings.CutSuffix(host, suffix); ok0 {
			return before
		}
	}

	return ""
}
