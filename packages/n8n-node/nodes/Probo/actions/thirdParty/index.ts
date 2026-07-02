// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import type { INodeProperties } from 'n8n-workflow';
import * as createOp from './create.operation';
import * as updateOp from './update.operation';
import * as deleteOp from './delete.operation';
import * as getOp from './get.operation';
import * as getAllOp from './getAll.operation';
import * as createContactOp from './createContact.operation';
import * as updateContactOp from './updateContact.operation';
import * as deleteContactOp from './deleteContact.operation';
import * as getContactOp from './getContact.operation';
import * as getAllContactsOp from './getAllContacts.operation';
import * as createServiceOp from './createService.operation';
import * as updateServiceOp from './updateService.operation';
import * as deleteServiceOp from './deleteService.operation';
import * as getServiceOp from './getService.operation';
import * as getAllServicesOp from './getAllServices.operation';
import * as createRiskAssessmentOp from './createRiskAssessment.operation';
import * as getRiskAssessmentOp from './getRiskAssessment.operation';
import * as getAllRiskAssessmentsOp from './getAllRiskAssessments.operation';
import * as getAllComplianceReportsOp from './getAllComplianceReports.operation';
import * as deleteComplianceReportOp from './deleteComplianceReport.operation';
import * as getBusinessAssociateAgreementOp from './getBusinessAssociateAgreement.operation';
import * as deleteBusinessAssociateAgreementOp from './deleteBusinessAssociateAgreement.operation';
import * as updateBusinessAssociateAgreementOp from './updateBusinessAssociateAgreement.operation';
import * as getDataPrivacyAgreementOp from './getDataPrivacyAgreement.operation';
import * as deleteDataPrivacyAgreementOp from './deleteDataPrivacyAgreement.operation';
import * as updateDataPrivacyAgreementOp from './updateDataPrivacyAgreement.operation';
import * as linkThirdPartyOp from './linkThirdParty.operation';
import * as unlinkThirdPartyOp from './unlinkThirdParty.operation';
import * as listChildThirdPartiesOp from './listChildThirdParties.operation';
import * as publishOp from './publish.operation';
import * as vetOp from './vet.operation';

export const description: INodeProperties[] = [
	{
		displayName: 'Operation',
		name: 'operation',
		type: 'options',
		noDataExpression: true,
		displayOptions: {
			show: {
				resource: ['thirdParty'],
			},
		},
		options: [
			{
				name: 'Create',
				value: 'create',
				description: 'Create a new third party',
				action: 'Create a third party',
			},
			{
				name: 'Create Contact',
				value: 'createContact',
				description: 'Create a new third party contact',
				action: 'Create a third party contact',
			},
			{
				name: 'Create Risk Assessment',
				value: 'createRiskAssessment',
				description: 'Create a new third party risk assessment',
				action: 'Create a third party risk assessment',
			},
			{
				name: 'Create Service',
				value: 'createService',
				description: 'Create a new third party service',
				action: 'Create a third party service',
			},
			{
				name: 'Delete',
				value: 'delete',
				description: 'Delete a third party',
				action: 'Delete a third party',
			},
			{
				name: 'Delete Business Associate Agreement',
				value: 'deleteBusinessAssociateAgreement',
				description: 'Delete a third party business associate agreement',
				action: 'Delete a third party business associate agreement',
			},
			{
				name: 'Delete Compliance Report',
				value: 'deleteComplianceReport',
				description: 'Delete a third party compliance report',
				action: 'Delete a third party compliance report',
			},
			{
				name: 'Delete Contact',
				value: 'deleteContact',
				description: 'Delete a third party contact',
				action: 'Delete a third party contact',
			},
			{
				name: 'Delete Data Privacy Agreement',
				value: 'deleteDataPrivacyAgreement',
				description: 'Delete a third party data privacy agreement',
				action: 'Delete a third party data privacy agreement',
			},
			{
				name: 'Delete Service',
				value: 'deleteService',
				description: 'Delete a third party service',
				action: 'Delete a third party service',
			},
			{
				name: 'Get',
				value: 'get',
				description: 'Get a third party',
				action: 'Get a third party',
			},
			{
				name: 'Get Business Associate Agreement',
				value: 'getBusinessAssociateAgreement',
				description: 'Get a third party business associate agreement',
				action: 'Get a third party business associate agreement',
			},
			{
				name: 'Get Contact',
				value: 'getContact',
				description: 'Get a third party contact',
				action: 'Get a third party contact',
			},
			{
				name: 'Get Data Privacy Agreement',
				value: 'getDataPrivacyAgreement',
				description: 'Get a third party data privacy agreement',
				action: 'Get a third party data privacy agreement',
			},
			{
				name: 'Get Many',
				value: 'getAll',
				description: 'Get many third parties',
				action: 'Get many third parties',
			},
			{
				name: 'Get Many Child Third Parties',
				value: 'listChildThirdParties',
				description: 'Get child third parties linked to a parent',
				action: 'Get many child third parties',
			},
			{
				name: 'Get Many Compliance Reports',
				value: 'getAllComplianceReports',
				description: 'Get many third party compliance reports',
				action: 'Get many third party compliance reports',
			},
			{
				name: 'Get Many Contacts',
				value: 'getAllContacts',
				description: 'Get many third party contacts',
				action: 'Get many third party contacts',
			},
			{
				name: 'Get Many Risk Assessments',
				value: 'getAllRiskAssessments',
				description: 'Get many third party risk assessments',
				action: 'Get many third party risk assessments',
			},
			{
				name: 'Get Many Services',
				value: 'getAllServices',
				description: 'Get many third party services',
				action: 'Get many third party services',
			},
			{
				name: 'Get Risk Assessment',
				value: 'getRiskAssessment',
				description: 'Get a third party risk assessment',
				action: 'Get a third party risk assessment',
			},
			{
				name: 'Get Service',
				value: 'getService',
				description: 'Get a third party service',
				action: 'Get a third party service',
			},
			{
				name: 'Link Third Party',
				value: 'linkThirdParty',
				description: 'Link a child third party to a parent third party',
				action: 'Link a child third party',
			},
			{
				name: 'Publish List',
				value: 'publish',
				description: 'Publish the third party register as a document version',
				action: 'Publish the third party register',
			},
			{
				name: 'Unlink Third Party',
				value: 'unlinkThirdParty',
				description: 'Unlink a child third party from a parent third party',
				action: 'Unlink a child third party',
			},
			{
				name: 'Update',
				value: 'update',
				description: 'Update an existing third party',
				action: 'Update a third party',
			},
			{
				name: 'Update Business Associate Agreement',
				value: 'updateBusinessAssociateAgreement',
				description: 'Update a third party business associate agreement validity',
				action: 'Update a third party business associate agreement',
			},
			{
				name: 'Update Contact',
				value: 'updateContact',
				description: 'Update an existing third party contact',
				action: 'Update a third party contact',
			},
			{
				name: 'Update Data Privacy Agreement',
				value: 'updateDataPrivacyAgreement',
				description: 'Update a third party data privacy agreement validity',
				action: 'Update a third party data privacy agreement',
			},
			{
				name: 'Update Service',
				value: 'updateService',
				description: 'Update an existing third party service',
				action: 'Update a third party service',
			},
			{
				name: 'Vet',
				value: 'vet',
				description: 'Start AI-powered vetting of a third party from its website',
				action: 'Vet a third party',
			},
		],
		default: 'create',
	},
	...createOp.description,
	...updateOp.description,
	...deleteOp.description,
	...getOp.description,
	...getAllOp.description,
	...createContactOp.description,
	...updateContactOp.description,
	...deleteContactOp.description,
	...getContactOp.description,
	...getAllContactsOp.description,
	...createServiceOp.description,
	...updateServiceOp.description,
	...deleteServiceOp.description,
	...getServiceOp.description,
	...getAllServicesOp.description,
	...createRiskAssessmentOp.description,
	...getRiskAssessmentOp.description,
	...getAllRiskAssessmentsOp.description,
	...getAllComplianceReportsOp.description,
	...deleteComplianceReportOp.description,
	...getBusinessAssociateAgreementOp.description,
	...deleteBusinessAssociateAgreementOp.description,
	...updateBusinessAssociateAgreementOp.description,
	...getDataPrivacyAgreementOp.description,
	...deleteDataPrivacyAgreementOp.description,
	...updateDataPrivacyAgreementOp.description,
	...linkThirdPartyOp.description,
	...unlinkThirdPartyOp.description,
	...listChildThirdPartiesOp.description,
	...publishOp.description,
	...vetOp.description,
];

export {
	createOp as create,
	updateOp as update,
	deleteOp as delete,
	getOp as get,
	getAllOp as getAll,
	createContactOp as createContact,
	updateContactOp as updateContact,
	deleteContactOp as deleteContact,
	getContactOp as getContact,
	getAllContactsOp as getAllContacts,
	createServiceOp as createService,
	updateServiceOp as updateService,
	deleteServiceOp as deleteService,
	getServiceOp as getService,
	getAllServicesOp as getAllServices,
	createRiskAssessmentOp as createRiskAssessment,
	getRiskAssessmentOp as getRiskAssessment,
	getAllRiskAssessmentsOp as getAllRiskAssessments,
	getAllComplianceReportsOp as getAllComplianceReports,
	deleteComplianceReportOp as deleteComplianceReport,
	getBusinessAssociateAgreementOp as getBusinessAssociateAgreement,
	deleteBusinessAssociateAgreementOp as deleteBusinessAssociateAgreement,
	updateBusinessAssociateAgreementOp as updateBusinessAssociateAgreement,
	getDataPrivacyAgreementOp as getDataPrivacyAgreement,
	deleteDataPrivacyAgreementOp as deleteDataPrivacyAgreement,
	updateDataPrivacyAgreementOp as updateDataPrivacyAgreement,
	linkThirdPartyOp as linkThirdParty,
	unlinkThirdPartyOp as unlinkThirdParty,
	listChildThirdPartiesOp as listChildThirdParties,
	publishOp as publish,
	vetOp as vet,
};
