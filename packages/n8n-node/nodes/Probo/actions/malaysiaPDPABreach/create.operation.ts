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
import { decisionOptions, incidentFields } from './fields';

const show = { resource: ['malaysiaPDPABreach'], operation: ['create'] };

export const description: INodeProperties[] = [
	{
		displayName: 'Organization ID',
		name: 'organizationId',
		type: 'string',
		displayOptions: { show },
		default: '',
		description: 'The ID of the organization',
		required: true,
	},
	{
		displayName: 'Title',
		name: 'title',
		type: 'string',
		displayOptions: { show },
		default: '',
		description: 'A concise incident title',
		required: true,
	},
	{
		displayName: 'Discovered At',
		name: 'discoveredAt',
		type: 'dateTime',
		displayOptions: { show },
		default: '',
		description: 'When the incident was discovered',
		required: true,
	},
	{
		displayName: 'Awareness At',
		name: 'awarenessAt',
		type: 'dateTime',
		displayOptions: { show },
		default: '',
		description: 'When the organization became aware of the breach',
		required: true,
	},
	{
		displayName: 'Affected Data Subjects',
		name: 'affectedDataSubjects',
		type: 'number',
		typeOptions: { minValue: 0, numberPrecision: 0 },
		displayOptions: { show },
		default: 0,
		description: 'Estimated number of affected data subjects',
		required: true,
	},
	{
		displayName: 'Affected Data Records',
		name: 'affectedDataRecords',
		type: 'number',
		typeOptions: { minValue: 0, numberPrecision: 0 },
		displayOptions: { show },
		default: 0,
		description: 'Estimated number of affected personal data records',
		required: true,
	},
	{
		displayName: 'Personal Data Types',
		name: 'personalDataTypes',
		type: 'string',
		typeOptions: { rows: 3 },
		displayOptions: { show },
		default: '',
		description: 'Categories of affected personal data',
		required: true,
	},
	{
		displayName: 'Potential Physical Harm',
		name: 'potentialPhysicalHarm',
		type: 'boolean',
		displayOptions: { show },
		default: false,
		description: 'Whether the breach could cause physical harm',
	},
	{
		displayName: 'Potential Financial Loss',
		name: 'potentialFinancialLoss',
		type: 'boolean',
		displayOptions: { show },
		default: false,
		description: 'Whether the breach could cause financial loss',
	},
	{
		displayName: 'Potential Credit or Property Damage',
		name: 'potentialCreditOrPropertyDamage',
		type: 'boolean',
		displayOptions: { show },
		default: false,
		description: 'Whether the breach could damage credit records or property',
	},
	{
		displayName: 'Potential Illegal Use',
		name: 'potentialIllegalUse',
		type: 'boolean',
		displayOptions: { show },
		default: false,
		description: 'Whether the data could be used for an illegal purpose',
	},
	{
		displayName: 'Sensitive Personal Data',
		name: 'sensitivePersonalData',
		type: 'boolean',
		displayOptions: { show },
		default: false,
		description: 'Whether sensitive personal data is involved',
	},
	{
		displayName: 'Potential Identity Fraud',
		name: 'potentialIdentityFraud',
		type: 'boolean',
		displayOptions: { show },
		default: false,
		description: 'Whether the breach could enable identity fraud',
	},
	{
		displayName: 'Human Notification Decision',
		name: 'notificationDecision',
		type: 'options',
		displayOptions: { show },
		options: decisionOptions,
		default: 'PENDING',
		description: 'Human decision recorded separately from the server recommendation',
		required: true,
	},
	{
		displayName: 'Additional Fields',
		name: 'additionalFields',
		type: 'collection',
		placeholder: 'Add Field',
		default: {},
		displayOptions: { show },
		options: [
			{
				displayName: 'Affected System',
				name: 'affectedSystem',
				type: 'string',
				default: '',
				description: 'System or service affected by the breach',
			},
			{
				displayName: 'Containment Actions',
				name: 'containmentActions',
				type: 'string',
				typeOptions: { rows: 3 },
				default: '',
				description: 'Actions taken to contain the incident',
			},
			{
				displayName: 'Decision Evidence',
				name: 'decisionEvidence',
				type: 'string',
				default: '',
				description: 'Evidence supporting the human decision',
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
				displayName: 'Description',
				name: 'description',
				type: 'string',
				typeOptions: { rows: 3 },
				default: '',
				description: 'Incident description',
			},
			{
				displayName: 'Likely Consequences',
				name: 'likelyConsequences',
				type: 'string',
				typeOptions: { rows: 3 },
				default: '',
				description: 'Likely consequences for affected data subjects',
			},
			{
				displayName: 'Occurred At',
				name: 'occurredAt',
				type: 'dateTime',
				default: '',
				description: 'When the incident occurred, if known',
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const input: IDataObject = {
		organizationId: this.getNodeParameter('organizationId', itemIndex) as string,
		title: this.getNodeParameter('title', itemIndex) as string,
		discoveredAt: this.getNodeParameter('discoveredAt', itemIndex) as string,
		awarenessAt: this.getNodeParameter('awarenessAt', itemIndex) as string,
		affectedDataSubjects: this.getNodeParameter('affectedDataSubjects', itemIndex) as number,
		affectedDataRecords: this.getNodeParameter('affectedDataRecords', itemIndex) as number,
		personalDataTypes: this.getNodeParameter('personalDataTypes', itemIndex) as string,
		potentialPhysicalHarm: this.getNodeParameter('potentialPhysicalHarm', itemIndex) as boolean,
		potentialFinancialLoss: this.getNodeParameter('potentialFinancialLoss', itemIndex) as boolean,
		potentialCreditOrPropertyDamage: this.getNodeParameter(
			'potentialCreditOrPropertyDamage',
			itemIndex,
		) as boolean,
		potentialIllegalUse: this.getNodeParameter('potentialIllegalUse', itemIndex) as boolean,
		sensitivePersonalData: this.getNodeParameter('sensitivePersonalData', itemIndex) as boolean,
		potentialIdentityFraud: this.getNodeParameter('potentialIdentityFraud', itemIndex) as boolean,
		notificationDecision: this.getNodeParameter('notificationDecision', itemIndex) as string,
	};
	const additionalFields = this.getNodeParameter('additionalFields', itemIndex, {}) as IDataObject;
	for (const [name, value] of Object.entries(additionalFields)) {
		if (value !== '' && value !== undefined) input[name] = value;
	}

	const query = `
		mutation CreateMalaysiaPDPABreachIncident($input: CreateMalaysiaPDPABreachIncidentInput!) {
			createMalaysiaPDPABreachIncident(input: $input) {
				incidentEdge { node { ${incidentFields} } }
			}
		}
	`;
	const responseData = await proboApiRequest.call(this, query, { input });
	return { json: responseData, pairedItem: { item: itemIndex } };
}
