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

package authz

import (
	"slices"

	"go.probo.inc/probo/pkg/coredata"
)

type (
	Role string

	Action string
)

const (
	RoleOwner    Role = "OWNER"
	RoleAdmin    Role = "ADMIN"
	RoleEmployee Role = "EMPLOYEE"
	RoleViewer   Role = "VIEWER"
	RoleAuditor  Role = "AUDITOR"
	RoleFull     Role = "FULL"
)

const (
	ActionGet Action = "get"

	ActionGetAssetType                      Action = "getAssetType"
	ActionGetAssignedTo                     Action = "getAssignedTo"
	ActionGetAuthMethod                     Action = "getAuthMethod"
	ActionGetBusinessAssociateAgreement     Action = "getBusinessAssociateAgreement"
	ActionGetBusinessOwner                  Action = "getBusinessOwner"
	ActionGetCustomDomain                   Action = "getCustomDomain"
	ActionGetDataPrivacyAgreement           Action = "getDataPrivacyAgreement"
	ActionGetDataProtectionOfficer          Action = "getDataProtectionOfficer"
	ActionGetDataProtectionImpactAssessment Action = "getDataProtectionImpactAssessment"
	ActionGetTransferImpactAssessment       Action = "getTransferImpactAssessment"
	ActionGetDocument                       Action = "getDocument"
	ActionGetReport                         Action = "getReport"
	ActionGetFile                           Action = "getFile"
	ActionGetAudit                          Action = "getAudit"
	ActionGetFileUrl                        Action = "getFileUrl"
	ActionGetFramework                      Action = "getFramework"
	ActionGetHorizontalLogoUrl              Action = "getHorizontalLogoUrl"
	ActionGetLogoUrl                        Action = "getLogoUrl"
	ActionGetMeasure                        Action = "getMeasure"
	ActionGetNdaFileUrl                     Action = "getNdaFileUrl"
	ActionGetOrganization                   Action = "getOrganization"
	ActionGetOwner                          Action = "getOwner"
	ActionGetSecurityOwner                  Action = "getSecurityOwner"
	ActionGetSigned                         Action = "getSigned"
	ActionGetSignableDocument               Action = "getSignableDocument"
	ActionGetSnapshot                       Action = "getSnapshot"
	ActionGetTask                           Action = "getTask"
	ActionGetTrustCenter                    Action = "getTrustCenter"
	ActionGetTrustCenterFile                Action = "getTrustCenterFile"
	ActionGetVendor                         Action = "getVendor"

	ActionActiveCount               Action = "activeCount"
	ActionAudit                     Action = "audit"
	ActionAvailableDocumentAccesses Action = "availableDocumentAccesses"
	ActionDocumentVersion           Action = "documentVersion"
	ActionDownloadUrl               Action = "downloadUrl"
	ActionMemberships               Action = "memberships"
	ActionPendingRequestCount       Action = "pendingRequestCount"
	ActionPeoples                   Action = "peoples"
	ActionReport                    Action = "report"
	ActionReportUrl                 Action = "reportUrl"
	ActionSignatures                Action = "signatures"
	ActionSignedBy                  Action = "signedBy"
	ActionSpMetadataUrl             Action = "spMetadataUrl"
	ActionTestLoginUrl              Action = "testLoginUrl"
	ActionTotalCount                Action = "totalCount"
	ActionTrustCenterFile           Action = "trustCenterFile"

	ActionListAccesses                Action = "listAccesses"
	ActionListAssets                  Action = "listAssets"
	ActionListAudits                  Action = "listAudits"
	ActionListComplianceReports       Action = "listComplianceReports"
	ActionListContacts                Action = "listContacts"
	ActionListContinualImprovements   Action = "listContinualImprovements"
	ActionListControls                Action = "listControls"
	ActionListData                    Action = "listData"
	ActionListDocuments               Action = "listDocuments"
	ActionListEvidences               Action = "listEvidences"
	ActionListFrameworks              Action = "listFrameworks"
	ActionListInvitations             Action = "listInvitations"
	ActionListMeasures                Action = "listMeasures"
	ActionListMeetings                Action = "listMeetings"
	ActionListMembers                 Action = "listMembers"
	ActionListNonconformities         Action = "listNonconformities"
	ActionListObligations             Action = "listObligations"
	ActionListPeople                  Action = "listPeople"
	ActionListProcessingActivities    Action = "listProcessingActivities"
	ActionListReferences              Action = "listReferences"
	ActionListRiskAssessments         Action = "listRiskAssessments"
	ActionListRisks                   Action = "listRisks"
	ActionListSAMLConfigurations      Action = "listSAMLConfigurations"
	ActionListServices                Action = "listServices"
	ActionListSlackConnections        Action = "listSlackConnections"
	ActionListSnapshots               Action = "listSnapshots"
	ActionListTasks                   Action = "listTasks"
	ActionListTrustCenterFiles        Action = "listTrustCenterFiles"
	ActionListVendors                 Action = "listVendors"
	ActionListVersions                Action = "listVersions"
	ActionListSignableDocuments       Action = "listSignableDocuments"
	ActionListSignableDocumentVersion Action = "listSignableDocumentVersion"

	ActionCreateAsset                          Action = "createAsset"
	ActionCreateAudit                          Action = "createAudit"
	ActionCreateContinualImprovement           Action = "createContinualImprovement"
	ActionCreateControl                        Action = "createControl"
	ActionCreateControlAuditMapping            Action = "createControlAuditMapping"
	ActionCreateControlDocumentMapping         Action = "createControlDocumentMapping"
	ActionCreateControlMeasureMapping          Action = "createControlMeasureMapping"
	ActionCreateControlSnapshotMapping         Action = "createControlSnapshotMapping"
	ActionCreateCustomDomain                   Action = "createCustomDomain"
	ActionCreateDatum                          Action = "createDatum"
	ActionCreateDocument                       Action = "createDocument"
	ActionCreateDraftDocumentVersion           Action = "createDraftDocumentVersion"
	ActionCreateFramework                      Action = "createFramework"
	ActionCreateMeasure                        Action = "createMeasure"
	ActionCreateMeeting                        Action = "createMeeting"
	ActionCreateNonconformity                  Action = "createNonconformity"
	ActionCreateObligation                     Action = "createObligation"
	ActionCreatePeople                         Action = "createPeople"
	ActionCreateProcessingActivity             Action = "createProcessingActivity"
	ActionCreateDataProtectionImpactAssessment Action = "createDataProtectionImpactAssessment"
	ActionCreateTransferImpactAssessment       Action = "createTransferImpactAssessment"
	ActionCreateRisk                           Action = "createRisk"
	ActionCreateRiskDocumentMapping            Action = "createRiskDocumentMapping"
	ActionCreateRiskMeasureMapping             Action = "createRiskMeasureMapping"
	ActionCreateRiskObligationMapping          Action = "createRiskObligationMapping"
	ActionCreateSAMLConfiguration              Action = "createSAMLConfiguration"
	ActionCreateSnapshot                       Action = "createSnapshot"
	ActionCreateTask                           Action = "createTask"
	ActionCreateTrustCenter                    Action = "createTrustCenter"
	ActionCreateTrustCenterAccess              Action = "createTrustCenterAccess"
	ActionCreateTrustCenterFile                Action = "createTrustCenterFile"
	ActionCreateTrustCenterReference           Action = "createTrustCenterReference"
	ActionCreateVendor                         Action = "createVendor"
	ActionCreateVendorContact                  Action = "createVendorContact"
	ActionCreateVendorRiskAssessment           Action = "createVendorRiskAssessment"
	ActionCreateVendorService                  Action = "createVendorService"

	ActionUpdateAsset                            Action = "updateAsset"
	ActionUpdateAudit                            Action = "updateAudit"
	ActionUpdateContinualImprovement             Action = "updateContinualImprovement"
	ActionUpdateControl                          Action = "updateControl"
	ActionUpdateDatum                            Action = "updateDatum"
	ActionUpdateDocument                         Action = "updateDocument"
	ActionUpdateDocumentVersion                  Action = "updateDocumentVersion"
	ActionUpdateFramework                        Action = "updateFramework"
	ActionUpdateMeasure                          Action = "updateMeasure"
	ActionUpdateMeeting                          Action = "updateMeeting"
	ActionUpdateMembership                       Action = "updateMembership"
	ActionUpdateNonconformity                    Action = "updateNonconformity"
	ActionUpdateObligation                       Action = "updateObligation"
	ActionUpdateOrganization                     Action = "updateOrganization"
	ActionUpdatePeople                           Action = "updatePeople"
	ActionUpdateProcessingActivity               Action = "updateProcessingActivity"
	ActionUpdateDataProtectionImpactAssessment   Action = "updateDataProtectionImpactAssessment"
	ActionUpdateTransferImpactAssessment         Action = "updateTransferImpactAssessment"
	ActionUpdateRisk                             Action = "updateRisk"
	ActionUpdateSAMLConfiguration                Action = "updateSAMLConfiguration"
	ActionUpdateTask                             Action = "updateTask"
	ActionUpdateTrustCenter                      Action = "updateTrustCenter"
	ActionUpdateTrustCenterAccess                Action = "updateTrustCenterAccess"
	ActionUpdateTrustCenterFile                  Action = "updateTrustCenterFile"
	ActionUpdateTrustCenterReference             Action = "updateTrustCenterReference"
	ActionUpdateVendor                           Action = "updateVendor"
	ActionUpdateVendorBusinessAssociateAgreement Action = "updateVendorBusinessAssociateAgreement"
	ActionUpdateVendorContact                    Action = "updateVendorContact"
	ActionUpdateVendorDataPrivacyAgreement       Action = "updateVendorDataPrivacyAgreement"
	ActionUpdateVendorService                    Action = "updateVendorService"

	ActionDeleteAsset                            Action = "deleteAsset"
	ActionDeleteAudit                            Action = "deleteAudit"
	ActionDeleteAuditReport                      Action = "deleteAuditReport"
	ActionDeleteContinualImprovement             Action = "deleteContinualImprovement"
	ActionDeleteControl                          Action = "deleteControl"
	ActionDeleteControlAuditMapping              Action = "deleteControlAuditMapping"
	ActionDeleteControlDocumentMapping           Action = "deleteControlDocumentMapping"
	ActionDeleteControlMeasureMapping            Action = "deleteControlMeasureMapping"
	ActionDeleteControlSnapshotMapping           Action = "deleteControlSnapshotMapping"
	ActionDeleteCustomDomain                     Action = "deleteCustomDomain"
	ActionDeleteDatum                            Action = "deleteDatum"
	ActionDeleteDocument                         Action = "deleteDocument"
	ActionDeleteDraftDocumentVersion             Action = "deleteDraftDocumentVersion"
	ActionDeleteEvidence                         Action = "deleteEvidence"
	ActionDeleteFramework                        Action = "deleteFramework"
	ActionDeleteInvitation                       Action = "deleteInvitation"
	ActionDeleteMeasure                          Action = "deleteMeasure"
	ActionDeleteMeeting                          Action = "deleteMeeting"
	ActionDeleteNonconformity                    Action = "deleteNonconformity"
	ActionDeleteObligation                       Action = "deleteObligation"
	ActionDeleteOrganization                     Action = "deleteOrganization"
	ActionDeleteOrganizationHorizontalLogo       Action = "deleteOrganizationHorizontalLogo"
	ActionDeletePeople                           Action = "deletePeople"
	ActionDeleteProcessingActivity               Action = "deleteProcessingActivity"
	ActionDeleteDataProtectionImpactAssessment   Action = "deleteDataProtectionImpactAssessment"
	ActionDeleteTransferImpactAssessment         Action = "deleteTransferImpactAssessment"
	ActionDeleteRisk                             Action = "deleteRisk"
	ActionDeleteRiskDocumentMapping              Action = "deleteRiskDocumentMapping"
	ActionDeleteRiskMeasureMapping               Action = "deleteRiskMeasureMapping"
	ActionDeleteRiskObligationMapping            Action = "deleteRiskObligationMapping"
	ActionDeleteSAMLConfiguration                Action = "deleteSAMLConfiguration"
	ActionDeleteSnapshot                         Action = "deleteSnapshot"
	ActionDeleteTask                             Action = "deleteTask"
	ActionDeleteTrustCenterAccess                Action = "deleteTrustCenterAccess"
	ActionDeleteTrustCenterFile                  Action = "deleteTrustCenterFile"
	ActionDeleteTrustCenterNDA                   Action = "deleteTrustCenterNDA"
	ActionDeleteTrustCenterReference             Action = "deleteTrustCenterReference"
	ActionDeleteVendor                           Action = "deleteVendor"
	ActionDeleteVendorBusinessAssociateAgreement Action = "deleteVendorBusinessAssociateAgreement"
	ActionDeleteVendorComplianceReport           Action = "deleteVendorComplianceReport"
	ActionDeleteVendorContact                    Action = "deleteVendorContact"
	ActionDeleteVendorDataPrivacyAgreement       Action = "deleteVendorDataPrivacyAgreement"
	ActionDeleteVendorService                    Action = "deleteVendorService"

	ActionAcceptInvitation                       Action = "acceptInvitation"
	ActionAssessVendor                           Action = "assessVendor"
	ActionAssignTask                             Action = "assignTask"
	ActionBulkDeleteDocuments                    Action = "bulkDeleteDocuments"
	ActionBulkExportDocuments                    Action = "bulkExportDocuments"
	ActionBulkPublishDocumentVersions            Action = "bulkPublishDocumentVersions"
	ActionBulkRequestSignatures                  Action = "bulkRequestSignatures"
	ActionCancelSignatureRequest                 Action = "cancelSignatureRequest"
	ActionSignDocument                           Action = "signDocument"
	ActionConfirmEmail                           Action = "confirmEmail"
	ActionDisableSAML                            Action = "disableSAML"
	ActionEnableSAML                             Action = "enableSAML"
	ActionExportDocumentVersionPDF               Action = "exportDocumentVersionPDF"
	ActionExportSignableVersionDocumentPDF       Action = "exportSignableVersionDocumentPDF"
	ActionExportFramework                        Action = "exportFramework"
	ActionGenerateDocumentChangelog              Action = "generateDocumentChangelog"
	ActionGenerateFrameworkStateOfApplicability  Action = "generateFrameworkStateOfApplicability"
	ActionImportFramework                        Action = "importFramework"
	ActionImportMeasure                          Action = "importMeasure"
	ActionInitiateDomainVerification             Action = "initiateDomainVerification"
	ActionInviteUser                             Action = "inviteUser"
	ActionPublishDocumentVersion                 Action = "publishDocumentVersion"
	ActionRemoveMember                           Action = "removeMember"
	ActionRequestSignature                       Action = "requestSignature"
	ActionSendSigningNotifications               Action = "sendSigningNotifications"
	ActionUnassignTask                           Action = "unassignTask"
	ActionUploadAuditReport                      Action = "uploadAuditReport"
	ActionUploadMeasureEvidence                  Action = "uploadMeasureEvidence"
	ActionUploadTrustCenterNDA                   Action = "uploadTrustCenterNDA"
	ActionUploadVendorBusinessAssociateAgreement Action = "uploadVendorBusinessAssociateAgreement"
	ActionUploadVendorComplianceReport           Action = "uploadVendorComplianceReport"
	ActionUploadVendorDataPrivacyAgreement       Action = "uploadVendorDataPrivacyAgreement"
	ActionVerifyDomain                           Action = "verifyDomain"
)

var (
	AllRoles = []Role{RoleOwner, RoleAdmin, RoleEmployee, RoleViewer, RoleAuditor, RoleFull}

	EditRoles = []Role{RoleOwner, RoleAdmin, RoleFull}

	CoreRoles        = []Role{RoleOwner, RoleAdmin, RoleViewer, RoleFull}
	NonEmployeeRoles = []Role{RoleOwner, RoleAdmin, RoleViewer, RoleAuditor, RoleFull}
	InternalRoles    = []Role{RoleOwner, RoleAdmin, RoleViewer, RoleEmployee, RoleFull}
)

var Permissions = map[uint16]map[Action][]Role{
	coredata.OrganizationEntityType: {
		ActionGet:        AllRoles,
		ActionGetLogoUrl: AllRoles,

		ActionListSignableDocuments: InternalRoles,

		ActionListDocuments:             NonEmployeeRoles,
		ActionGetHorizontalLogoUrl:      NonEmployeeRoles,
		ActionPeoples:                   NonEmployeeRoles,
		ActionTotalCount:                NonEmployeeRoles,
		ActionListFrameworks:            NonEmployeeRoles,
		ActionListControls:              NonEmployeeRoles,
		ActionListVendors:               NonEmployeeRoles,
		ActionListPeople:                NonEmployeeRoles,
		ActionListMeasures:              NonEmployeeRoles,
		ActionListRisks:                 NonEmployeeRoles,
		ActionListAssets:                NonEmployeeRoles,
		ActionListData:                  NonEmployeeRoles,
		ActionListAudits:                NonEmployeeRoles,
		ActionListNonconformities:       NonEmployeeRoles,
		ActionListObligations:           NonEmployeeRoles,
		ActionListContinualImprovements: NonEmployeeRoles,
		ActionListProcessingActivities:  NonEmployeeRoles,
		ActionListSnapshots:             NonEmployeeRoles,
		ActionConfirmEmail:              NonEmployeeRoles,
		ActionAcceptInvitation:          NonEmployeeRoles,

		ActionListTrustCenterFiles:   CoreRoles,
		ActionGetTrustCenter:         CoreRoles,
		ActionMemberships:            CoreRoles,
		ActionListMembers:            CoreRoles,
		ActionListInvitations:        CoreRoles,
		ActionListSlackConnections:   CoreRoles,
		ActionGetCustomDomain:        CoreRoles,
		ActionListSAMLConfigurations: CoreRoles,
		ActionListMeetings:           CoreRoles,
		ActionListTasks:              CoreRoles,

		ActionUpdateOrganization:               EditRoles,
		ActionDeleteOrganizationHorizontalLogo: EditRoles,
		ActionCreateTrustCenter:                EditRoles,
		ActionInviteUser:                       EditRoles,
		ActionUpdateMembership:                 EditRoles,
		ActionCreatePeople:                     EditRoles,
		ActionCreateVendor:                     EditRoles,
		ActionCreateFramework:                  EditRoles,
		ActionImportFramework:                  EditRoles,
		ActionCreateControl:                    EditRoles,
		ActionCreateMeasure:                    EditRoles,
		ActionImportMeasure:                    EditRoles,
		ActionCreateMeeting:                    EditRoles,
		ActionCreateTask:                       EditRoles,
		ActionCreateRisk:                       EditRoles,
		ActionCreateDocument:                   EditRoles,
		ActionCreateAsset:                      EditRoles,
		ActionCreateDatum:                      EditRoles,
		ActionCreateAudit:                      EditRoles,
		ActionCreateNonconformity:              EditRoles,
		ActionCreateObligation:                 EditRoles,
		ActionCreateContinualImprovement:       EditRoles,
		ActionCreateProcessingActivity:         EditRoles,
		ActionCreateSnapshot:                   EditRoles,
		ActionCreateTrustCenterFile:            EditRoles,
		ActionSendSigningNotifications:         EditRoles,

		ActionRemoveMember: {RoleOwner, RoleFull},

		ActionCreateCustomDomain:         {RoleOwner},
		ActionDeleteCustomDomain:         {RoleOwner},
		ActionInitiateDomainVerification: {RoleOwner},
		ActionVerifyDomain:               {RoleOwner},
		ActionCreateSAMLConfiguration:    {RoleOwner},
		ActionDeleteOrganization:         {RoleOwner},
	},
	coredata.TrustCenterEntityType: {
		ActionGet:             CoreRoles,
		ActionGetNdaFileUrl:   CoreRoles,
		ActionGetOrganization: CoreRoles,
		ActionListAccesses:    CoreRoles,
		ActionListReferences:  CoreRoles,

		ActionUpdateTrustCenter:          EditRoles,
		ActionUploadTrustCenterNDA:       EditRoles,
		ActionDeleteTrustCenterNDA:       EditRoles,
		ActionCreateTrustCenterAccess:    EditRoles,
		ActionCreateTrustCenterReference: EditRoles,
	},
	coredata.TrustCenterAccessEntityType: {
		ActionGet:                       CoreRoles,
		ActionActiveCount:               CoreRoles,
		ActionPendingRequestCount:       CoreRoles,
		ActionAvailableDocumentAccesses: CoreRoles,
		ActionGetTrustCenterFile:        CoreRoles,
		ActionGetReport:                 CoreRoles,

		ActionUpdateTrustCenterAccess: EditRoles,
		ActionDeleteTrustCenterAccess: EditRoles,
	},
	coredata.TrustCenterReferenceEntityType: {
		ActionGet:        CoreRoles,
		ActionGetLogoUrl: CoreRoles,

		ActionUpdateTrustCenterReference: EditRoles,
		ActionDeleteTrustCenterReference: EditRoles,
	},
	coredata.TrustCenterFileEntityType: {
		ActionGet:        CoreRoles,
		ActionGetFileUrl: CoreRoles,

		ActionUpdateTrustCenterFile: EditRoles,
		ActionGetTrustCenterFile:    EditRoles,
		ActionDeleteTrustCenterFile: EditRoles,
	},
	coredata.UserEntityType: {
		ActionGet: NonEmployeeRoles,
	},
	coredata.MembershipEntityType: {
		ActionGet:           NonEmployeeRoles,
		ActionGetAuthMethod: NonEmployeeRoles,
	},
	coredata.InvitationEntityType: {
		ActionGet:             NonEmployeeRoles,
		ActionGetOrganization: NonEmployeeRoles,

		ActionDeleteInvitation: EditRoles,
	},
	coredata.PeopleEntityType: {
		ActionGet: NonEmployeeRoles,

		ActionUpdatePeople: EditRoles,
		ActionDeletePeople: EditRoles,
	},
	coredata.VendorEntityType: {
		ActionGet:                           NonEmployeeRoles,
		ActionGetOrganization:               NonEmployeeRoles,
		ActionListComplianceReports:         NonEmployeeRoles,
		ActionGetBusinessAssociateAgreement: NonEmployeeRoles,
		ActionGetDataPrivacyAgreement:       NonEmployeeRoles,
		ActionListContacts:                  NonEmployeeRoles,
		ActionListServices:                  NonEmployeeRoles,
		ActionListRiskAssessments:           NonEmployeeRoles,
		ActionGetBusinessOwner:              NonEmployeeRoles,
		ActionGetSecurityOwner:              NonEmployeeRoles,

		ActionUpdateVendor:                           EditRoles,
		ActionDeleteVendor:                           EditRoles,
		ActionCreateVendorContact:                    EditRoles,
		ActionCreateVendorService:                    EditRoles,
		ActionUploadVendorComplianceReport:           EditRoles,
		ActionUploadVendorBusinessAssociateAgreement: EditRoles,
		ActionDeleteVendorBusinessAssociateAgreement: EditRoles,
		ActionUploadVendorDataPrivacyAgreement:       EditRoles,
		ActionCreateVendorRiskAssessment:             EditRoles,
		ActionAssessVendor:                           EditRoles,
	},
	coredata.VendorComplianceReportEntityType: {
		ActionGet:       NonEmployeeRoles,
		ActionGetVendor: NonEmployeeRoles,
		ActionGetFile:   NonEmployeeRoles,

		ActionDeleteVendorComplianceReport: EditRoles,
	},
	coredata.VendorBusinessAssociateAgreementEntityType: {
		ActionGet:        NonEmployeeRoles,
		ActionGetVendor:  NonEmployeeRoles,
		ActionGetFileUrl: NonEmployeeRoles,

		ActionUpdateVendorBusinessAssociateAgreement: EditRoles,
		ActionDeleteVendorBusinessAssociateAgreement: EditRoles,
	},
	coredata.VendorContactEntityType: {
		ActionGet:       NonEmployeeRoles,
		ActionGetVendor: NonEmployeeRoles,

		ActionUpdateVendorContact: EditRoles,
		ActionDeleteVendorContact: EditRoles,
	},
	coredata.VendorServiceEntityType: {
		ActionGet:       NonEmployeeRoles,
		ActionGetVendor: NonEmployeeRoles,

		ActionUpdateVendorService: EditRoles,
		ActionDeleteVendorService: EditRoles,
	},
	coredata.VendorDataPrivacyAgreementEntityType: {
		ActionGet:        NonEmployeeRoles,
		ActionGetVendor:  NonEmployeeRoles,
		ActionGetFileUrl: NonEmployeeRoles,

		ActionUpdateVendorDataPrivacyAgreement: EditRoles,
		ActionDeleteVendorDataPrivacyAgreement: EditRoles,
	},
	coredata.VendorRiskAssessmentEntityType: {
		ActionGet:       NonEmployeeRoles,
		ActionGetVendor: NonEmployeeRoles,
	},
	coredata.FrameworkEntityType: {
		ActionGet:             NonEmployeeRoles,
		ActionGetOrganization: NonEmployeeRoles,
		ActionListControls:    NonEmployeeRoles,
		ActionGetLogoUrl:      NonEmployeeRoles,

		ActionCreateControl:                         EditRoles,
		ActionUpdateFramework:                       EditRoles,
		ActionDeleteFramework:                       EditRoles,
		ActionGenerateFrameworkStateOfApplicability: EditRoles,
		ActionExportFramework:                       EditRoles,
	},
	coredata.ControlEntityType: {
		ActionGet:           NonEmployeeRoles,
		ActionGetFramework:  NonEmployeeRoles,
		ActionListMeasures:  NonEmployeeRoles,
		ActionListDocuments: NonEmployeeRoles,
		ActionListAudits:    NonEmployeeRoles,
		ActionListSnapshots: NonEmployeeRoles,

		ActionUpdateControl:                EditRoles,
		ActionDeleteControl:                EditRoles,
		ActionCreateControlMeasureMapping:  EditRoles,
		ActionCreateControlDocumentMapping: EditRoles,
		ActionDeleteControlMeasureMapping:  EditRoles,
		ActionDeleteControlDocumentMapping: EditRoles,
		ActionCreateControlAuditMapping:    EditRoles,
		ActionDeleteControlAuditMapping:    EditRoles,
		ActionCreateControlSnapshotMapping: EditRoles,
		ActionDeleteControlSnapshotMapping: EditRoles,
	},
	coredata.MeasureEntityType: {
		ActionListTasks: CoreRoles,

		ActionGet:           NonEmployeeRoles,
		ActionListEvidences: NonEmployeeRoles,
		ActionListRisks:     NonEmployeeRoles,
		ActionListControls:  NonEmployeeRoles,
		ActionTotalCount:    NonEmployeeRoles,

		ActionUpdateMeasure:         EditRoles,
		ActionDeleteMeasure:         EditRoles,
		ActionUploadMeasureEvidence: EditRoles,
	},
	coredata.TaskEntityType: {
		ActionGet:             CoreRoles,
		ActionGetAssignedTo:   CoreRoles,
		ActionGetOrganization: CoreRoles,
		ActionGetMeasure:      CoreRoles,
		ActionListEvidences:   CoreRoles,

		ActionUpdateTask:   EditRoles,
		ActionDeleteTask:   EditRoles,
		ActionAssignTask:   EditRoles,
		ActionUnassignTask: EditRoles,
	},
	coredata.EvidenceEntityType: {
		ActionGet:        NonEmployeeRoles,
		ActionGetFile:    NonEmployeeRoles,
		ActionGetTask:    CoreRoles,
		ActionGetMeasure: NonEmployeeRoles,

		ActionDeleteEvidence: EditRoles,
	},
	coredata.DocumentEntityType: {
		ActionListSignableDocumentVersion: AllRoles,
		ActionGetSigned:                   AllRoles,

		ActionGetSignableDocument: InternalRoles,

		ActionGet:                 NonEmployeeRoles,
		ActionGetOwner:            NonEmployeeRoles,
		ActionGetOrganization:     NonEmployeeRoles,
		ActionBulkExportDocuments: NonEmployeeRoles,
		ActionTotalCount:          NonEmployeeRoles,
		ActionListControls:        NonEmployeeRoles,
		ActionListVersions:        NonEmployeeRoles,

		ActionUpdateDocument:              EditRoles,
		ActionDeleteDocument:              EditRoles,
		ActionPublishDocumentVersion:      EditRoles,
		ActionBulkPublishDocumentVersions: EditRoles,
		ActionBulkDeleteDocuments:         EditRoles,
		ActionGenerateDocumentChangelog:   EditRoles,
		ActionCreateDraftDocumentVersion:  EditRoles,
		ActionDeleteDraftDocumentVersion:  EditRoles,
		ActionUpdateDocumentVersion:       EditRoles,
		ActionRequestSignature:            EditRoles,
		ActionBulkRequestSignatures:       EditRoles,
		ActionSendSigningNotifications:    EditRoles,
	},
	coredata.DocumentVersionEntityType: {
		ActionSignDocument: AllRoles,

		ActionExportSignableVersionDocumentPDF: AllRoles,
		ActionGetSigned:                        AllRoles,

		ActionGet:                      NonEmployeeRoles,
		ActionGetFile:                  NonEmployeeRoles,
		ActionGetOwner:                 NonEmployeeRoles,
		ActionGetDocument:              NonEmployeeRoles,
		ActionSignatures:               NonEmployeeRoles,
		ActionExportDocumentVersionPDF: NonEmployeeRoles,

		ActionUpdateDocumentVersion:      EditRoles,
		ActionRequestSignature:           EditRoles,
		ActionDeleteDraftDocumentVersion: EditRoles,
	},
	coredata.DocumentVersionSignatureEntityType: {
		ActionGet:             NonEmployeeRoles,
		ActionDocumentVersion: NonEmployeeRoles,
		ActionSignedBy:        NonEmployeeRoles,

		ActionCancelSignatureRequest: EditRoles,
	},
	coredata.RiskEntityType: {
		ActionGet:             NonEmployeeRoles,
		ActionGetOwner:        NonEmployeeRoles,
		ActionGetOrganization: NonEmployeeRoles,
		ActionTotalCount:      NonEmployeeRoles,
		ActionListControls:    NonEmployeeRoles,
		ActionListMeasures:    NonEmployeeRoles,
		ActionListDocuments:   NonEmployeeRoles,
		ActionListObligations: NonEmployeeRoles,

		ActionUpdateRisk:                  EditRoles,
		ActionDeleteRisk:                  EditRoles,
		ActionCreateRiskMeasureMapping:    EditRoles,
		ActionDeleteRiskMeasureMapping:    EditRoles,
		ActionCreateRiskDocumentMapping:   EditRoles,
		ActionDeleteRiskDocumentMapping:   EditRoles,
		ActionCreateRiskObligationMapping: EditRoles,
		ActionDeleteRiskObligationMapping: EditRoles,
	},
	coredata.AssetEntityType: {
		ActionGet:             NonEmployeeRoles,
		ActionGetOwner:        NonEmployeeRoles,
		ActionListVendors:     NonEmployeeRoles,
		ActionGetAssetType:    NonEmployeeRoles,
		ActionGetOrganization: NonEmployeeRoles,

		ActionUpdateAsset: EditRoles,
		ActionDeleteAsset: EditRoles,
	},
	coredata.DatumEntityType: {
		ActionGet:             NonEmployeeRoles,
		ActionGetOwner:        NonEmployeeRoles,
		ActionGetOrganization: NonEmployeeRoles,
		ActionListVendors:     NonEmployeeRoles,

		ActionUpdateDatum: EditRoles,
		ActionDeleteDatum: EditRoles,
	},
	coredata.AuditEntityType: {
		ActionGet:             NonEmployeeRoles,
		ActionGetFile:         NonEmployeeRoles,
		ActionGetFramework:    NonEmployeeRoles,
		ActionGetOrganization: NonEmployeeRoles,
		ActionReport:          NonEmployeeRoles,
		ActionReportUrl:       NonEmployeeRoles,
		ActionListControls:    NonEmployeeRoles,

		ActionUpdateAudit:       EditRoles,
		ActionDeleteAudit:       EditRoles,
		ActionUploadAuditReport: EditRoles,
		ActionDeleteAuditReport: EditRoles,
	},
	coredata.ReportEntityType: {
		ActionGet:             NonEmployeeRoles,
		ActionGetAudit:        NonEmployeeRoles,
		ActionGetFile:         NonEmployeeRoles,
		ActionGetOrganization: NonEmployeeRoles,
		ActionGetSnapshot:     NonEmployeeRoles,
		ActionDownloadUrl:     NonEmployeeRoles,
	},
	coredata.NonconformityEntityType: {
		ActionGet:             NonEmployeeRoles,
		ActionGetOwner:        NonEmployeeRoles,
		ActionGetOrganization: NonEmployeeRoles,
		ActionAudit:           NonEmployeeRoles,

		ActionUpdateNonconformity: EditRoles,
		ActionDeleteNonconformity: EditRoles,
	},
	coredata.ObligationEntityType: {
		ActionGet:             NonEmployeeRoles,
		ActionGetOrganization: NonEmployeeRoles,
		ActionGetOwner:        NonEmployeeRoles,
		ActionListRisks:       NonEmployeeRoles,

		ActionUpdateObligation: EditRoles,
		ActionDeleteObligation: EditRoles,
	},
	coredata.ContinualImprovementEntityType: {
		ActionGet:             NonEmployeeRoles,
		ActionGetOwner:        NonEmployeeRoles,
		ActionGetOrganization: NonEmployeeRoles,

		ActionUpdateContinualImprovement: EditRoles,
		ActionDeleteContinualImprovement: EditRoles,
	},
	coredata.ProcessingActivityEntityType: {
		ActionGet:                               NonEmployeeRoles,
		ActionGetOrganization:                   NonEmployeeRoles,
		ActionListVendors:                       NonEmployeeRoles,
		ActionGetDataProtectionOfficer:          NonEmployeeRoles,
		ActionGetDataProtectionImpactAssessment: NonEmployeeRoles,
		ActionGetTransferImpactAssessment:       NonEmployeeRoles,

		ActionUpdateProcessingActivity:             EditRoles,
		ActionDeleteProcessingActivity:             EditRoles,
		ActionCreateDataProtectionImpactAssessment: EditRoles,
		ActionCreateTransferImpactAssessment:       EditRoles,
	},
	coredata.DataProtectionImpactAssessmentEntityType: {
		ActionGet:             NonEmployeeRoles,
		ActionGetOrganization: NonEmployeeRoles,

		ActionCreateDataProtectionImpactAssessment: EditRoles,
		ActionUpdateDataProtectionImpactAssessment: EditRoles,
		ActionDeleteDataProtectionImpactAssessment: EditRoles,
	},
	coredata.TransferImpactAssessmentEntityType: {
		ActionGet:             NonEmployeeRoles,
		ActionGetOrganization: NonEmployeeRoles,

		ActionCreateTransferImpactAssessment: EditRoles,
		ActionUpdateTransferImpactAssessment: EditRoles,
		ActionDeleteTransferImpactAssessment: EditRoles,
	},
	coredata.SnapshotEntityType: {
		ActionGet:             NonEmployeeRoles,
		ActionGetOrganization: NonEmployeeRoles,
		ActionListControls:    NonEmployeeRoles,

		ActionDeleteSnapshot: EditRoles,
	},
	coredata.CustomDomainEntityType: {
		ActionGet: {RoleOwner, RoleAdmin},

		ActionDeleteCustomDomain: {RoleOwner},
	},
	coredata.SAMLConfigurationEntityType: {
		ActionGet:           {RoleOwner, RoleAdmin},
		ActionSpMetadataUrl: {RoleOwner, RoleAdmin},
		ActionTestLoginUrl:  {RoleOwner, RoleAdmin},

		ActionUpdateSAMLConfiguration: {RoleOwner, RoleAdmin},
		ActionDeleteSAMLConfiguration: {RoleOwner, RoleAdmin},
		ActionEnableSAML:              {RoleOwner, RoleAdmin},
		ActionDisableSAML:             {RoleOwner, RoleAdmin},
		ActionVerifyDomain:            {RoleOwner},
	},
	coredata.FileEntityType: {
		ActionGet:         NonEmployeeRoles,
		ActionDownloadUrl: NonEmployeeRoles,
	},
	coredata.TrustCenterDocumentAccessEntityType: {
		ActionGet:             CoreRoles,
		ActionReport:          CoreRoles,
		ActionTrustCenterFile: CoreRoles,
	},
	coredata.MeetingEntityType: {
		ActionGet:             CoreRoles,
		ActionGetOrganization: CoreRoles,
		ActionTotalCount:      CoreRoles,

		ActionUpdateMeeting: EditRoles,
		ActionDeleteMeeting: EditRoles,
	},
}

func GetPermissionsForAction(entityType uint16, action Action) []Role {
	if entityActions, ok := Permissions[entityType]; ok {
		if roles, ok := entityActions[action]; ok {
			return roles
		}
	}
	return nil
}

func GetPermissionsByRole(userRole Role) map[string]map[Action]bool {
	permissions := make(map[string]map[Action]bool)

	for entityType, actions := range Permissions {
		entityTypeName, ok := coredata.EntityModel(entityType)
		if !ok {
			continue
		}

		if permissions[entityTypeName] == nil {
			permissions[entityTypeName] = make(map[Action]bool)
		}

		for action, allowedRoles := range actions {
			if slices.Contains(allowedRoles, userRole) {
				permissions[entityTypeName][action] = true
			}
		}
	}

	return permissions
}
