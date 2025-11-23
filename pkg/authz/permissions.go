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
	RoleOwner  Role = "OWNER"
	RoleAdmin  Role = "ADMIN"
	RoleViewer Role = "VIEWER"
	RoleFull   Role = "FULL"
)

const (
	ActionGet Action = "get"

	ActionGetAssetType                  Action = "getAssetType"
	ActionGetAssignedTo                 Action = "getAssignedTo"
	ActionGetAuthMethod                 Action = "getAuthMethod"
	ActionGetBusinessAssociateAgreement Action = "getBusinessAssociateAgreement"
	ActionGetBusinessOwner              Action = "getBusinessOwner"
	ActionGetCustomDomain               Action = "getCustomDomain"
	ActionGetDataPrivacyAgreement       Action = "getDataPrivacyAgreement"
	ActionGetFile                       Action = "getFile"
	ActionGetFileUrl                    Action = "getFileUrl"
	ActionGetFramework                  Action = "getFramework"
	ActionGetHorizontalLogoUrl          Action = "getHorizontalLogoUrl"
	ActionGetLogoUrl                    Action = "getLogoUrl"
	ActionGetMeasure                    Action = "getMeasure"
	ActionGetNdaFileUrl                 Action = "getNdaFileUrl"
	ActionGetOrganization               Action = "getOrganization"
	ActionGetOwner                      Action = "getOwner"
	ActionGetSecurityOwner              Action = "getSecurityOwner"
	ActionGetSnapshot                   Action = "getSnapshot"
	ActionGetTask                       Action = "getTask"
	ActionGetTrustCenter                Action = "getTrustCenter"
	ActionGetTrustCenterFile            Action = "getTrustCenterFile"
	ActionGetVendor                     Action = "getVendor"

	ActionActiveCount               Action = "activeCount"
	ActionAudit                     Action = "audit"
	ActionAvailableDocumentAccesses Action = "availableDocumentAccesses"
	ActionDocument                  Action = "document"
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

	ActionListAccesses              Action = "listAccesses"
	ActionListAssets                Action = "listAssets"
	ActionListAudits                Action = "listAudits"
	ActionListComplianceReports     Action = "listComplianceReports"
	ActionListContacts              Action = "listContacts"
	ActionListContinualImprovements Action = "listContinualImprovements"
	ActionListControls              Action = "listControls"
	ActionListData                  Action = "listData"
	ActionListDocuments             Action = "listDocuments"
	ActionListEvidences             Action = "listEvidences"
	ActionListFrameworks            Action = "listFrameworks"
	ActionListInvitations           Action = "listInvitations"
	ActionListMeasures              Action = "listMeasures"
	ActionListMeetings              Action = "listMeetings"
	ActionListMembers               Action = "listMembers"
	ActionListNonconformities       Action = "listNonconformities"
	ActionListObligations           Action = "listObligations"
	ActionListPeople                Action = "listPeople"
	ActionListProcessingActivities  Action = "listProcessingActivities"
	ActionListReferences            Action = "listReferences"
	ActionListRiskAssessments       Action = "listRiskAssessments"
	ActionListRisks                 Action = "listRisks"
	ActionListSAMLConfigurations    Action = "listSAMLConfigurations"
	ActionListServices              Action = "listServices"
	ActionListSlackConnections      Action = "listSlackConnections"
	ActionListSnapshots             Action = "listSnapshots"
	ActionListTasks                 Action = "listTasks"
	ActionListTrustCenterFiles      Action = "listTrustCenterFiles"
	ActionListVendors               Action = "listVendors"
	ActionListVersions              Action = "listVersions"

	ActionCreateAsset                  Action = "createAsset"
	ActionCreateAudit                  Action = "createAudit"
	ActionCreateContinualImprovement   Action = "createContinualImprovement"
	ActionCreateControl                Action = "createControl"
	ActionCreateControlAuditMapping    Action = "createControlAuditMapping"
	ActionCreateControlDocumentMapping Action = "createControlDocumentMapping"
	ActionCreateControlMeasureMapping  Action = "createControlMeasureMapping"
	ActionCreateControlSnapshotMapping Action = "createControlSnapshotMapping"
	ActionCreateCustomDomain           Action = "createCustomDomain"
	ActionCreateDatum                  Action = "createDatum"
	ActionCreateDocument               Action = "createDocument"
	ActionCreateDraftDocumentVersion   Action = "createDraftDocumentVersion"
	ActionCreateFramework              Action = "createFramework"
	ActionCreateMeasure                Action = "createMeasure"
	ActionCreateMeeting                Action = "createMeeting"
	ActionCreateNonconformity          Action = "createNonconformity"
	ActionCreateObligation             Action = "createObligation"
	ActionCreatePeople                 Action = "createPeople"
	ActionCreateProcessingActivity     Action = "createProcessingActivity"
	ActionCreateRisk                   Action = "createRisk"
	ActionCreateRiskDocumentMapping    Action = "createRiskDocumentMapping"
	ActionCreateRiskMeasureMapping     Action = "createRiskMeasureMapping"
	ActionCreateRiskObligationMapping  Action = "createRiskObligationMapping"
	ActionCreateSAMLConfiguration      Action = "createSAMLConfiguration"
	ActionCreateSnapshot               Action = "createSnapshot"
	ActionCreateTask                   Action = "createTask"
	ActionCreateTrustCenter            Action = "createTrustCenter"
	ActionCreateTrustCenterAccess      Action = "createTrustCenterAccess"
	ActionCreateTrustCenterFile        Action = "createTrustCenterFile"
	ActionCreateTrustCenterReference   Action = "createTrustCenterReference"
	ActionCreateVendor                 Action = "createVendor"
	ActionCreateVendorContact          Action = "createVendorContact"
	ActionCreateVendorRiskAssessment   Action = "createVendorRiskAssessment"
	ActionCreateVendorService          Action = "createVendorService"

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
	ActionConfirmEmail                           Action = "confirmEmail"
	ActionDisableSAML                            Action = "disableSAML"
	ActionEnableSAML                             Action = "enableSAML"
	ActionExportDocumentVersionPDF               Action = "exportDocumentVersionPDF"
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
	AllRoles  = []Role{RoleOwner, RoleAdmin, RoleViewer, RoleFull}
	EditRoles = []Role{RoleOwner, RoleAdmin, RoleFull}
)

var Permissions = map[uint16]map[Action][]Role{
	coredata.OrganizationEntityType: {
		ActionGet:                       AllRoles,
		ActionGetLogoUrl:                AllRoles,
		ActionGetHorizontalLogoUrl:      AllRoles,
		ActionMemberships:               AllRoles,
		ActionPeoples:                   AllRoles,
		ActionTotalCount:                AllRoles,
		ActionListMembers:               AllRoles,
		ActionListInvitations:           AllRoles,
		ActionListSlackConnections:      AllRoles,
		ActionListFrameworks:            AllRoles,
		ActionListControls:              AllRoles,
		ActionListVendors:               AllRoles,
		ActionListPeople:                AllRoles,
		ActionListDocuments:             AllRoles,
		ActionListMeetings:              AllRoles,
		ActionListMeasures:              AllRoles,
		ActionListRisks:                 AllRoles,
		ActionListTasks:                 AllRoles,
		ActionListAssets:                AllRoles,
		ActionListData:                  AllRoles,
		ActionListAudits:                AllRoles,
		ActionListNonconformities:       AllRoles,
		ActionListObligations:           AllRoles,
		ActionListContinualImprovements: AllRoles,
		ActionListProcessingActivities:  AllRoles,
		ActionListSnapshots:             AllRoles,
		ActionListTrustCenterFiles:      AllRoles,
		ActionGetTrustCenter:            AllRoles,
		ActionGetCustomDomain:           AllRoles,
		ActionListSAMLConfigurations:    AllRoles,
		ActionConfirmEmail:              AllRoles,
		ActionAcceptInvitation:          AllRoles,

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
		ActionGet:             AllRoles,
		ActionGetNdaFileUrl:   AllRoles,
		ActionGetOrganization: AllRoles,
		ActionListAccesses:    AllRoles,
		ActionListReferences:  AllRoles,

		ActionUpdateTrustCenter:          EditRoles,
		ActionUploadTrustCenterNDA:       EditRoles,
		ActionDeleteTrustCenterNDA:       EditRoles,
		ActionCreateTrustCenterAccess:    EditRoles,
		ActionCreateTrustCenterReference: EditRoles,
	},
	coredata.TrustCenterAccessEntityType: {
		ActionGet:                       AllRoles,
		ActionActiveCount:               AllRoles,
		ActionPendingRequestCount:       AllRoles,
		ActionAvailableDocumentAccesses: AllRoles,

		ActionUpdateTrustCenterAccess: EditRoles,
		ActionDeleteTrustCenterAccess: EditRoles,
	},
	coredata.TrustCenterReferenceEntityType: {
		ActionGet:        AllRoles,
		ActionGetLogoUrl: AllRoles,

		ActionUpdateTrustCenterReference: EditRoles,
		ActionDeleteTrustCenterReference: EditRoles,
	},
	coredata.TrustCenterFileEntityType: {
		ActionGet:        AllRoles,
		ActionGetFileUrl: AllRoles,

		ActionUpdateTrustCenterFile: EditRoles,
		ActionGetTrustCenterFile:    EditRoles,
		ActionDeleteTrustCenterFile: EditRoles,
	},
	coredata.UserEntityType: {
		ActionGet: AllRoles,
	},
	coredata.MembershipEntityType: {
		ActionGet:           AllRoles,
		ActionGetAuthMethod: AllRoles,
	},
	coredata.InvitationEntityType: {
		ActionGet:             AllRoles,
		ActionGetOrganization: AllRoles,

		ActionDeleteInvitation: EditRoles,
	},
	coredata.PeopleEntityType: {
		ActionGet: AllRoles,

		ActionUpdatePeople: EditRoles,
		ActionDeletePeople: EditRoles,
	},
	coredata.VendorEntityType: {
		ActionGet:                           AllRoles,
		ActionGetOrganization:               AllRoles,
		ActionListComplianceReports:         AllRoles,
		ActionGetBusinessAssociateAgreement: AllRoles,
		ActionGetDataPrivacyAgreement:       AllRoles,
		ActionListContacts:                  AllRoles,
		ActionListServices:                  AllRoles,
		ActionListRiskAssessments:           AllRoles,
		ActionGetBusinessOwner:              AllRoles,
		ActionGetSecurityOwner:              AllRoles,

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
		ActionGet:       AllRoles,
		ActionGetVendor: AllRoles,
		ActionGetFile:   AllRoles,

		ActionDeleteVendorComplianceReport: EditRoles,
	},
	coredata.VendorBusinessAssociateAgreementEntityType: {
		ActionGet:        AllRoles,
		ActionGetVendor:  AllRoles,
		ActionGetFileUrl: AllRoles,

		ActionUpdateVendorBusinessAssociateAgreement: EditRoles,
		ActionDeleteVendorBusinessAssociateAgreement: EditRoles,
	},
	coredata.VendorContactEntityType: {
		ActionGet:       AllRoles,
		ActionGetVendor: AllRoles,

		ActionUpdateVendorContact: EditRoles,
		ActionDeleteVendorContact: EditRoles,
	},
	coredata.VendorServiceEntityType: {
		ActionGet:       AllRoles,
		ActionGetVendor: AllRoles,

		ActionUpdateVendorService: EditRoles,
		ActionDeleteVendorService: EditRoles,
	},
	coredata.VendorDataPrivacyAgreementEntityType: {
		ActionGet:        AllRoles,
		ActionGetVendor:  AllRoles,
		ActionGetFileUrl: AllRoles,

		ActionUpdateVendorDataPrivacyAgreement: EditRoles,
		ActionDeleteVendorDataPrivacyAgreement: EditRoles,
	},
	coredata.VendorRiskAssessmentEntityType: {
		ActionGet: AllRoles,
	},
	coredata.FrameworkEntityType: {
		ActionGet:             AllRoles,
		ActionGetOrganization: AllRoles,
		ActionListControls:    AllRoles,

		ActionCreateControl:                         EditRoles,
		ActionUpdateFramework:                       EditRoles,
		ActionDeleteFramework:                       EditRoles,
		ActionGenerateFrameworkStateOfApplicability: EditRoles,
		ActionExportFramework:                       EditRoles,
	},
	coredata.ControlEntityType: {
		ActionGet:           AllRoles,
		ActionGetFramework:  AllRoles,
		ActionListMeasures:  AllRoles,
		ActionListDocuments: AllRoles,
		ActionListAudits:    AllRoles,
		ActionListSnapshots: AllRoles,

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
		ActionGet:           AllRoles,
		ActionListEvidences: AllRoles,
		ActionListTasks:     AllRoles,
		ActionListRisks:     AllRoles,
		ActionListControls:  AllRoles,
		ActionTotalCount:    AllRoles,

		ActionUpdateMeasure:         EditRoles,
		ActionDeleteMeasure:         EditRoles,
		ActionUploadMeasureEvidence: EditRoles,
	},
	coredata.TaskEntityType: {
		ActionGet:             AllRoles,
		ActionGetAssignedTo:   AllRoles,
		ActionGetOrganization: AllRoles,
		ActionGetMeasure:      AllRoles,
		ActionListEvidences:   AllRoles,

		ActionUpdateTask:   EditRoles,
		ActionDeleteTask:   EditRoles,
		ActionAssignTask:   EditRoles,
		ActionUnassignTask: EditRoles,
	},
	coredata.EvidenceEntityType: {
		ActionGet:        AllRoles,
		ActionGetFile:    AllRoles,
		ActionGetTask:    AllRoles,
		ActionGetMeasure: AllRoles,

		ActionDeleteEvidence: EditRoles,
	},
	coredata.DocumentEntityType: {
		ActionGet:                      AllRoles,
		ActionExportDocumentVersionPDF: AllRoles,
		ActionGetOwner:                 AllRoles,
		ActionGetOrganization:          AllRoles,
		ActionListVersions:             AllRoles,
		ActionListControls:             AllRoles,
		ActionTotalCount:               AllRoles,

		ActionUpdateDocument:              EditRoles,
		ActionDeleteDocument:              EditRoles,
		ActionPublishDocumentVersion:      EditRoles,
		ActionBulkPublishDocumentVersions: EditRoles,
		ActionBulkDeleteDocuments:         EditRoles,
		ActionBulkExportDocuments:         EditRoles,
		ActionGenerateDocumentChangelog:   EditRoles,
		ActionCreateDraftDocumentVersion:  EditRoles,
		ActionDeleteDraftDocumentVersion:  EditRoles,
		ActionUpdateDocumentVersion:       EditRoles,
		ActionRequestSignature:            EditRoles,
		ActionBulkRequestSignatures:       EditRoles,
		ActionSendSigningNotifications:    EditRoles,
		ActionCancelSignatureRequest:      EditRoles,
	},
	coredata.DocumentVersionEntityType: {
		ActionGet:                      AllRoles,
		ActionGetFile:                  AllRoles,
		ActionGetOwner:                 AllRoles,
		ActionDocument:                 AllRoles,
		ActionSignatures:               AllRoles,
		ActionExportDocumentVersionPDF: AllRoles,

		ActionUpdateDocumentVersion:    EditRoles,
		ActionRequestSignature:         EditRoles,
		ActionBulkRequestSignatures:    EditRoles,
		ActionSendSigningNotifications: EditRoles,
		ActionCancelSignatureRequest:   EditRoles,
	},
	coredata.DocumentVersionSignatureEntityType: {
		ActionGet:             AllRoles,
		ActionDocumentVersion: AllRoles,
		ActionSignedBy:        AllRoles,
	},
	coredata.RiskEntityType: {
		ActionGet:             AllRoles,
		ActionGetOwner:        AllRoles,
		ActionGetOrganization: AllRoles,
		ActionTotalCount:      AllRoles,
		ActionListControls:    AllRoles,
		ActionListMeasures:    AllRoles,
		ActionListDocuments:   AllRoles,
		ActionListObligations: AllRoles,

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
		ActionGet:             AllRoles,
		ActionGetOwner:        AllRoles,
		ActionListVendors:     AllRoles,
		ActionGetAssetType:    AllRoles,
		ActionGetOrganization: AllRoles,

		ActionUpdateAsset: EditRoles,
		ActionDeleteAsset: EditRoles,
	},
	coredata.DatumEntityType: {
		ActionGet:             AllRoles,
		ActionGetOwner:        AllRoles,
		ActionGetOrganization: AllRoles,
		ActionListVendors:     AllRoles,

		ActionUpdateDatum: EditRoles,
		ActionDeleteDatum: EditRoles,
	},
	coredata.AuditEntityType: {
		ActionGet:             AllRoles,
		ActionGetFile:         AllRoles,
		ActionGetFramework:    AllRoles,
		ActionGetOrganization: AllRoles,
		ActionReport:          AllRoles,
		ActionReportUrl:       AllRoles,
		ActionListControls:    AllRoles,

		ActionUpdateAudit:       EditRoles,
		ActionDeleteAudit:       EditRoles,
		ActionUploadAuditReport: EditRoles,
		ActionDeleteAuditReport: EditRoles,
	},
	coredata.ReportEntityType: {
		ActionGet:             AllRoles,
		ActionGetFile:         AllRoles,
		ActionGetOrganization: AllRoles,
		ActionGetSnapshot:     AllRoles,
		ActionDownloadUrl:     AllRoles,
	},
	coredata.NonconformityEntityType: {
		ActionGet:             AllRoles,
		ActionGetOwner:        AllRoles,
		ActionGetOrganization: AllRoles,
		ActionAudit:           AllRoles,

		ActionUpdateNonconformity: EditRoles,
		ActionDeleteNonconformity: EditRoles,
	},
	coredata.ObligationEntityType: {
		ActionGet:             AllRoles,
		ActionGetOrganization: AllRoles,
		ActionGetOwner:        AllRoles,
		ActionListRisks:       AllRoles,

		ActionUpdateObligation: EditRoles,
		ActionDeleteObligation: EditRoles,
	},
	coredata.ContinualImprovementEntityType: {
		ActionGet:             AllRoles,
		ActionGetOwner:        AllRoles,
		ActionGetOrganization: AllRoles,

		ActionUpdateContinualImprovement: EditRoles,
		ActionDeleteContinualImprovement: EditRoles,
	},
	coredata.ProcessingActivityEntityType: {
		ActionGet:             AllRoles,
		ActionGetOrganization: AllRoles,
		ActionListVendors:     AllRoles,

		ActionUpdateProcessingActivity: EditRoles,
		ActionDeleteProcessingActivity: EditRoles,
	},
	coredata.SnapshotEntityType: {
		ActionGet:             AllRoles,
		ActionGetOrganization: AllRoles,
		ActionListControls:    AllRoles,

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
		ActionGet:         AllRoles,
		ActionDownloadUrl: AllRoles,
	},
	coredata.TrustCenterDocumentAccessEntityType: {
		ActionGet:             AllRoles,
		ActionReport:          AllRoles,
		ActionTrustCenterFile: AllRoles,
	},
	coredata.MeetingEntityType: {
		ActionGet:             AllRoles,
		ActionGetOrganization: AllRoles,
		ActionTotalCount:      AllRoles,

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
