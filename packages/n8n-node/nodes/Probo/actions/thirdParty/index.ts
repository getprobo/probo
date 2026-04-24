// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
				description: 'Create a new thirdParty',
				action: 'Create a thirdParty',
			},
			{
				name: 'Create Contact',
				value: 'createContact',
				description: 'Create a new thirdParty contact',
				action: 'Create a thirdParty contact',
			},
			{
				name: 'Create Risk Assessment',
				value: 'createRiskAssessment',
				description: 'Create a new thirdParty risk assessment',
				action: 'Create a thirdParty risk assessment',
			},
			{
				name: 'Create Service',
				value: 'createService',
				description: 'Create a new thirdParty service',
				action: 'Create a thirdParty service',
			},
			{
				name: 'Delete',
				value: 'delete',
				description: 'Delete a thirdParty',
				action: 'Delete a thirdParty',
			},
			{
				name: 'Delete Business Associate Agreement',
				value: 'deleteBusinessAssociateAgreement',
				description: 'Delete a thirdParty business associate agreement',
				action: 'Delete a thirdParty business associate agreement',
			},
			{
				name: 'Delete Compliance Report',
				value: 'deleteComplianceReport',
				description: 'Delete a thirdParty compliance report',
				action: 'Delete a thirdParty compliance report',
			},
			{
				name: 'Delete Contact',
				value: 'deleteContact',
				description: 'Delete a thirdParty contact',
				action: 'Delete a thirdParty contact',
			},
			{
				name: 'Delete Data Privacy Agreement',
				value: 'deleteDataPrivacyAgreement',
				description: 'Delete a thirdParty data privacy agreement',
				action: 'Delete a thirdParty data privacy agreement',
			},
			{
				name: 'Delete Service',
				value: 'deleteService',
				description: 'Delete a thirdParty service',
				action: 'Delete a thirdParty service',
			},
			{
				name: 'Get',
				value: 'get',
				description: 'Get a thirdParty',
				action: 'Get a thirdParty',
			},
			{
				name: 'Get Business Associate Agreement',
				value: 'getBusinessAssociateAgreement',
				description: 'Get a thirdParty business associate agreement',
				action: 'Get a thirdParty business associate agreement',
			},
			{
				name: 'Get Contact',
				value: 'getContact',
				description: 'Get a thirdParty contact',
				action: 'Get a thirdParty contact',
			},
			{
				name: 'Get Data Privacy Agreement',
				value: 'getDataPrivacyAgreement',
				description: 'Get a thirdParty data privacy agreement',
				action: 'Get a thirdParty data privacy agreement',
			},
			{
				name: 'Get Many',
				value: 'getAll',
				description: 'Get many thirdParties',
				action: 'Get many thirdParties',
			},
			{
				name: 'Get Many Compliance Reports',
				value: 'getAllComplianceReports',
				description: 'Get many thirdParty compliance reports',
				action: 'Get many thirdParty compliance reports',
			},
			{
				name: 'Get Many Contacts',
				value: 'getAllContacts',
				description: 'Get many thirdParty contacts',
				action: 'Get many thirdParty contacts',
			},
			{
				name: 'Get Many Risk Assessments',
				value: 'getAllRiskAssessments',
				description: 'Get many thirdParty risk assessments',
				action: 'Get many thirdParty risk assessments',
			},
			{
				name: 'Get Many Services',
				value: 'getAllServices',
				description: 'Get many thirdParty services',
				action: 'Get many thirdParty services',
			},
			{
				name: 'Get Risk Assessment',
				value: 'getRiskAssessment',
				description: 'Get a thirdParty risk assessment',
				action: 'Get a thirdParty risk assessment',
			},
			{
				name: 'Get Service',
				value: 'getService',
				description: 'Get a thirdParty service',
				action: 'Get a thirdParty service',
			},
			{
				name: 'Update',
				value: 'update',
				description: 'Update an existing thirdParty',
				action: 'Update a thirdParty',
			},
			{
				name: 'Update Business Associate Agreement',
				value: 'updateBusinessAssociateAgreement',
				description: 'Update a thirdParty business associate agreement validity',
				action: 'Update a thirdParty business associate agreement',
			},
			{
				name: 'Update Contact',
				value: 'updateContact',
				description: 'Update an existing thirdParty contact',
				action: 'Update a thirdParty contact',
			},
			{
				name: 'Update Data Privacy Agreement',
				value: 'updateDataPrivacyAgreement',
				description: 'Update a thirdParty data privacy agreement validity',
				action: 'Update a thirdParty data privacy agreement',
			},
			{
				name: 'Update Service',
				value: 'updateService',
				description: 'Update an existing thirdParty service',
				action: 'Update a thirdParty service',
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
};
