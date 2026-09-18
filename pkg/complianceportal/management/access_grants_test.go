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
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestCreateAccessGrantsTargets(t *testing.T) {
	t.Parallel()

	documentID := gid.New(gid.NilTenant, coredata.DocumentEntityType)

	t.Run(
		"no grant ids",
		func(t *testing.T) {
			t.Parallel()

			assert.False(t, createAccessGrantsTargets(&CreateAccessRequest{}))
		},
	)

	t.Run(
		"document ids",
		func(t *testing.T) {
			t.Parallel()

			assert.True(
				t,
				createAccessGrantsTargets(
					&CreateAccessRequest{DocumentIDs: []gid.GID{documentID}},
				),
			)
		},
	)
}

func TestUpdateAccessNewlyGrantsTargets(t *testing.T) {
	t.Parallel()

	documentID := gid.New(gid.NilTenant, coredata.DocumentEntityType)
	reportID := gid.New(gid.NilTenant, coredata.FileEntityType)
	fileID := gid.New(gid.NilTenant, coredata.CompliancePortalFileEntityType)

	t.Run(
		"new document grant",
		func(t *testing.T) {
			t.Parallel()

			assert.True(
				t,
				updateAccessNewlyGrantsTargets(
					&UpdateAccessRequest{
						DocumentAccesses: []UpdateDocumentAccessRequest{
							{
								ID:     documentID,
								Status: coredata.CompliancePortalDocumentAccessStatusGranted,
							},
						},
					},
					nil,
					nil,
					nil,
				),
			)
		},
	)

	t.Run(
		"already granted document",
		func(t *testing.T) {
			t.Parallel()

			assert.False(
				t,
				updateAccessNewlyGrantsTargets(
					&UpdateAccessRequest{
						DocumentAccesses: []UpdateDocumentAccessRequest{
							{
								ID:     documentID,
								Status: coredata.CompliancePortalDocumentAccessStatusGranted,
							},
						},
					},
					map[gid.GID]struct{}{documentID: {}},
					nil,
					nil,
				),
			)
		},
	)

	t.Run(
		"new report grant",
		func(t *testing.T) {
			t.Parallel()

			assert.True(
				t,
				updateAccessNewlyGrantsTargets(
					&UpdateAccessRequest{
						ReportAccesses: []UpdateDocumentAccessRequest{
							{
								ID:     reportID,
								Status: coredata.CompliancePortalDocumentAccessStatusGranted,
							},
						},
					},
					nil,
					nil,
					nil,
				),
			)
		},
	)

	t.Run(
		"already granted report",
		func(t *testing.T) {
			t.Parallel()

			assert.False(
				t,
				updateAccessNewlyGrantsTargets(
					&UpdateAccessRequest{
						ReportAccesses: []UpdateDocumentAccessRequest{
							{
								ID:     reportID,
								Status: coredata.CompliancePortalDocumentAccessStatusGranted,
							},
						},
					},
					nil,
					map[gid.GID]struct{}{reportID: {}},
					nil,
				),
			)
		},
	)

	t.Run(
		"new file grant",
		func(t *testing.T) {
			t.Parallel()

			assert.True(
				t,
				updateAccessNewlyGrantsTargets(
					&UpdateAccessRequest{
						CompliancePortalFileAccesses: []UpdateDocumentAccessRequest{
							{
								ID:     fileID,
								Status: coredata.CompliancePortalDocumentAccessStatusGranted,
							},
						},
					},
					nil,
					nil,
					nil,
				),
			)
		},
	)

	t.Run(
		"already granted file",
		func(t *testing.T) {
			t.Parallel()

			assert.False(
				t,
				updateAccessNewlyGrantsTargets(
					&UpdateAccessRequest{
						CompliancePortalFileAccesses: []UpdateDocumentAccessRequest{
							{
								ID:     fileID,
								Status: coredata.CompliancePortalDocumentAccessStatusGranted,
							},
						},
					},
					nil,
					nil,
					map[gid.GID]struct{}{fileID: {}},
				),
			)
		},
	)

	t.Run(
		"revoke only",
		func(t *testing.T) {
			t.Parallel()

			assert.False(
				t,
				updateAccessNewlyGrantsTargets(
					&UpdateAccessRequest{
						DocumentAccesses: []UpdateDocumentAccessRequest{
							{
								ID:     documentID,
								Status: coredata.CompliancePortalDocumentAccessStatusRevoked,
							},
						},
					},
					nil,
					nil,
					nil,
				),
			)
		},
	)
}
