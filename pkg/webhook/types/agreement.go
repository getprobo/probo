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

package types

import (
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type (
	AgreementAcceptance struct {
		ID                    gid.GID   `json:"id"`
		OrganizationID        gid.GID   `json:"organizationId"`
		SiteID                gid.GID   `json:"siteId"`
		AgreementID           gid.GID   `json:"agreementId"`
		SubjectID             gid.GID   `json:"subjectId"`
		ExternalUserID        string    `json:"externalUserId"`
		Email                 string    `json:"email"`
		AccountID             *string   `json:"accountId,omitempty"`
		DocumentVersionID     gid.GID   `json:"documentVersionId"`
		ContentSHA256         string    `json:"contentSha256"`
		ElectronicSignatureID gid.GID   `json:"electronicSignatureId"`
		IPAddress             *string   `json:"ipAddress,omitempty"`
		AcceptedAt            time.Time `json:"acceptedAt"`
		CreatedAt             time.Time `json:"createdAt"`
	}
)

func NewAgreementAcceptance(
	acceptance *coredata.AgreementAcceptance,
	subject *coredata.AgreementSubject,
) *AgreementAcceptance {
	return &AgreementAcceptance{
		ID:                    acceptance.ID,
		OrganizationID:        acceptance.OrganizationID,
		SiteID:                acceptance.SiteID,
		AgreementID:           acceptance.AgreementID,
		SubjectID:             acceptance.SubjectID,
		ExternalUserID:        subject.ExternalUserID,
		Email:                 subject.Email,
		AccountID:             subject.AccountID,
		DocumentVersionID:     acceptance.DocumentVersionID,
		ContentSHA256:         acceptance.ContentSHA256,
		ElectronicSignatureID: acceptance.ElectronicSignatureID,
		IPAddress:             acceptance.IPAddress,
		AcceptedAt:            acceptance.AcceptedAt,
		CreatedAt:             acceptance.CreatedAt,
	}
}
