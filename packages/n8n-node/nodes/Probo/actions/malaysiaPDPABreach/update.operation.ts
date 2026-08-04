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

import type {
	IDataObject,
	IExecuteFunctions,
	INodeExecutionData,
	INodeProperties,
} from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';
import { decisionOptions, incidentFields, incidentIdField } from './fields';

export const description: INodeProperties[] = [
	incidentIdField('update'),
	{
		displayName: 'Update Fields',
		name: 'updateFields',
		type: 'collection',
		placeholder: 'Add Field',
		default: {},
		displayOptions: {
			show: { resource: ['malaysiaPDPABreach'], operation: ['update'] },
		},
		options: [
			{
				displayName: 'Affected Data Records',
				name: 'affectedDataRecords',
				type: 'number',
				typeOptions: { minValue: 0, numberPrecision: 0 },
				default: 0,
				description: 'Updated number of affected data records',
			},
			{
				displayName: 'Affected Data Subjects',
				name: 'affectedDataSubjects',
				type: 'number',
				typeOptions: { minValue: 0, numberPrecision: 0 },
				default: 0,
				description: 'Updated number of affected data subjects',
			},
			{
				displayName: 'Awareness At',
				name: 'awarenessAt',
				type: 'dateTime',
				default: '',
				description: 'Updated regulatory awareness time',
			},
			{
				displayName: 'Commissioner Confirmation Received At',
				name: 'commissionerConfirmationReceivedAt',
				type: 'dateTime',
				default: '',
				description: 'When Commissioner confirmation was received',
			},
			{
				displayName: 'Commissioner Confirmation Reference',
				name: 'commissionerConfirmationReference',
				type: 'string',
				default: '',
				description: 'Reference or evidence for Commissioner confirmation',
			},
			{
				displayName: 'Commissioner Notification Reference',
				name: 'commissionerNotificationReference',
				type: 'string',
				default: '',
				description: 'Reference or evidence for the Commissioner notification',
			},
			{
				displayName: 'Commissioner Notified At',
				name: 'commissionerNotifiedAt',
				type: 'dateTime',
				default: '',
				description: 'When the Commissioner was notified',
			},
			{
				displayName: 'Containment Actions',
				name: 'containmentActions',
				type: 'string',
				typeOptions: { rows: 3 },
				default: '',
				description: 'Updated actions taken to contain the incident',
			},
			{
				displayName: 'Data Subjects Notification Evidence',
				name: 'dataSubjectsNotificationEvidence',
				type: 'string',
				default: '',
				description: 'Evidence that affected data subjects were notified',
			},
			{
				displayName: 'Data Subjects Notified At',
				name: 'dataSubjectsNotifiedAt',
				type: 'dateTime',
				default: '',
				description: 'When affected data subjects were notified',
			},
			{
				displayName: 'Decision Evidence',
				name: 'decisionEvidence',
				type: 'string',
				default: '',
				description: 'Evidence supporting the human notification decision',
			},
			{
				displayName: 'Decision Rationale',
				name: 'decisionRationale',
				type: 'string',
				typeOptions: { rows: 3 },
				default: '',
				description: 'Reason for the human notification decision',
			},
			{
				displayName: 'Delayed Notification Evidence',
				name: 'delayedNotificationEvidence',
				type: 'string',
				default: '',
				description: 'Evidence supporting a late-notification reason',
			},
			{
				displayName: 'Delayed Notification Reason',
				name: 'delayedNotificationReason',
				type: 'string',
				typeOptions: { rows: 3 },
				default: '',
				description: 'Reason the Commissioner notification was late',
			},
			{
				displayName: 'Description',
				name: 'description',
				type: 'string',
				typeOptions: { rows: 3 },
				default: '',
				description: 'Updated incident description',
			},
			{
				displayName: 'Discovered At',
				name: 'discoveredAt',
				type: 'dateTime',
				default: '',
				description: 'Updated incident discovery time',
			},
			{
				displayName: 'Human Notification Decision',
				name: 'notificationDecision',
				type: 'options',
				options: decisionOptions,
				default: 'PENDING',
				description: 'Human decision recorded separately from the server recommendation',
			},
			{
				displayName: 'Likely Consequences',
				name: 'likelyConsequences',
				type: 'string',
				typeOptions: { rows: 3 },
				default: '',
				description: 'Updated likely consequences for affected data subjects',
			},
			{
				displayName: 'Personal Data Types',
				name: 'personalDataTypes',
				type: 'string',
				typeOptions: { rows: 3 },
				default: '',
				description: 'Updated categories of affected personal data',
			},
			{
				displayName: 'Potential Credit or Property Damage',
				name: 'potentialCreditOrPropertyDamage',
				type: 'boolean',
				default: false,
				description: 'Whether credit or property damage is now considered possible',
			},
			{
				displayName: 'Potential Financial Loss',
				name: 'potentialFinancialLoss',
				type: 'boolean',
				default: false,
				description: 'Whether financial loss is now considered possible',
			},
			{
				displayName: 'Potential Identity Fraud',
				name: 'potentialIdentityFraud',
				type: 'boolean',
				default: false,
				description: 'Whether identity fraud is now considered possible',
			},
			{
				displayName: 'Potential Illegal Use',
				name: 'potentialIllegalUse',
				type: 'boolean',
				default: false,
				description: 'Whether illegal use is now considered possible',
			},
			{
				displayName: 'Potential Physical Harm',
				name: 'potentialPhysicalHarm',
				type: 'boolean',
				default: false,
				description: 'Whether physical harm is now considered possible',
			},
			{
				displayName: 'Sensitive Personal Data',
				name: 'sensitivePersonalData',
				type: 'boolean',
				default: false,
				description: 'Whether sensitive personal data is involved',
			},
			{
				displayName: 'Title',
				name: 'title',
				type: 'string',
				default: '',
				description: 'Updated incident title',
			},
		],
	},
	{
		displayName: 'Clear Fields',
		name: 'clearFields',
		type: 'multiOptions',
		displayOptions: {
			show: { resource: ['malaysiaPDPABreach'], operation: ['update'] },
		},
		options: [
			{ name: 'Commissioner Confirmation', value: 'commissionerConfirmation' },
			{ name: 'Commissioner Notification', value: 'commissionerNotification' },
			{ name: 'Data Subjects Notification', value: 'dataSubjectsNotification' },
		],
		default: [],
		description: 'Optional recorded notification evidence to clear',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const incidentId = this.getNodeParameter('incidentId', itemIndex) as string;
	const updateFields = this.getNodeParameter('updateFields', itemIndex, {}) as IDataObject;
	const clearFields = this.getNodeParameter('clearFields', itemIndex, []) as string[];
	const input: IDataObject = { id: incidentId, ...updateFields };

	if (clearFields.includes('commissionerNotification')) {
		input.commissionerNotifiedAt = null;
		input.commissionerNotificationReference = null;
	}
	if (clearFields.includes('commissionerConfirmation')) {
		input.commissionerConfirmationReceivedAt = null;
		input.commissionerConfirmationReference = null;
	}
	if (clearFields.includes('dataSubjectsNotification')) {
		input.dataSubjectsNotifiedAt = null;
		input.dataSubjectsNotificationEvidence = null;
	}

	const query = `
		mutation UpdateMalaysiaPDPABreachIncident($input: UpdateMalaysiaPDPABreachIncidentInput!) {
			updateMalaysiaPDPABreachIncident(input: $input) {
				incident { ${incidentFields} }
			}
		}
	`;
	const responseData = await proboApiRequest.call(this, query, { input });
	return { json: responseData, pairedItem: { item: itemIndex } };
}
