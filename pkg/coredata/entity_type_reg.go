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

package coredata

type ctxKey struct{ name string }

var (
	ContextKeyIPAddress = &ctxKey{name: "ip_address"}
)

const (
	OrganizationEntityType                     uint16 = 0
	FrameworkEntityType                        uint16 = 1
	MeasureEntityType                          uint16 = 2
	TaskEntityType                             uint16 = 3
	EvidenceEntityType                         uint16 = 4
	ConnectorEntityType                        uint16 = 5
	VendorRiskAssessmentEntityType             uint16 = 6
	VendorEntityType                           uint16 = 7
	PeopleEntityType                           uint16 = 8
	VendorComplianceReportEntityType           uint16 = 9
	DocumentEntityType                         uint16 = 10
	UserEntityType                             uint16 = 11
	SessionEntityType                          uint16 = 12
	EmailEntityType                            uint16 = 13
	ControlEntityType                          uint16 = 14
	RiskEntityType                             uint16 = 15
	DocumentVersionEntityType                  uint16 = 16
	DocumentVersionSignatureEntityType         uint16 = 17
	AssetEntityType                            uint16 = 18
	DatumEntityType                            uint16 = 19
	AuditEntityType                            uint16 = 20
	ReportEntityType                           uint16 = 21
	TrustCenterEntityType                      uint16 = 22
	TrustCenterAccessEntityType                uint16 = 23
	VendorBusinessAssociateAgreementEntityType uint16 = 24
	FileEntityType                             uint16 = 25
	VendorContactEntityType                    uint16 = 26
	VendorDataPrivacyAgreementEntityType       uint16 = 27
	NonconformityEntityType                    uint16 = 28
	ObligationEntityType                       uint16 = 29
	VendorServiceEntityType                    uint16 = 30
	SnapshotEntityType                         uint16 = 31
	ContinualImprovementEntityType             uint16 = 32
	ProcessingActivityEntityType               uint16 = 33
	ExportJobEntityType                        uint16 = 34
	TrustCenterReferenceEntityType             uint16 = 35
	TrustCenterDocumentAccessEntityType        uint16 = 36
	CustomDomainEntityType                     uint16 = 37
	InvitationEntityType                       uint16 = 38
	MembershipEntityType                       uint16 = 39
	SlackMessageEntityType                     uint16 = 40
	TrustCenterFileEntityType                  uint16 = 41
	SAMLConfigurationEntityType                uint16 = 42
	UserAPIKeyEntityType                       uint16 = 43
	UserAPIKeyMembershipEntityType             uint16 = 44
	MeetingEntityType                          uint16 = 45
)
