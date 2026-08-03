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

export const description: INodeProperties[] = [
	{
		displayName: 'Organization ID',
		name: 'organizationId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['malaysiaPDPA'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the organization',
		required: true,
	},
	{
		displayName: 'Total Data Subjects',
		name: 'totalDataSubjects',
		type: 'number',
		typeOptions: {
			minValue: 0,
			numberPrecision: 0,
		},
		displayOptions: {
			show: {
				resource: ['malaysiaPDPA'],
				operation: ['update'],
			},
		},
		default: 0,
		description: 'Estimated total number of data subjects',
		required: true,
	},
	{
		displayName: 'Sensitive or Financial Data Subjects',
		name: 'sensitiveDataSubjects',
		type: 'number',
		typeOptions: {
			minValue: 0,
			numberPrecision: 0,
		},
		displayOptions: {
			show: {
				resource: ['malaysiaPDPA'],
				operation: ['update'],
			},
		},
		default: 0,
		description: 'Estimated number of data subjects whose sensitive or financial data is processed',
		required: true,
	},
	{
		displayName: 'Regular and Systematic Monitoring',
		name: 'regularSystematicMonitoring',
		type: 'boolean',
		displayOptions: {
			show: {
				resource: ['malaysiaPDPA'],
				operation: ['update'],
			},
		},
		default: false,
		description: 'Whether processing includes regular and systematic monitoring',
		required: true,
	},
	{
		displayName: 'DPO and Notification',
		name: 'additionalFields',
		type: 'collection',
		placeholder: 'Add Field',
		default: {},
		displayOptions: {
			show: {
				resource: ['malaysiaPDPA'],
				operation: ['update'],
			},
		},
		options: [
			{
				displayName: 'Clear Commissioner Notification',
				name: 'clearCommissionerNotification',
				type: 'boolean',
				default: false,
				description: 'Whether to clear the Commissioner notification while preserving the DPO appointment',
			},
			{
				displayName: 'Clear DPO Appointment',
				name: 'clearDPO',
				type: 'boolean',
				default: false,
				description: 'Whether to clear the DPO appointment and Commissioner notification',
			},
			{
				displayName: 'Commissioner Notification Reference',
				name: 'commissionerNotificationReference',
				type: 'string',
				default: '',
				description: 'Evidence or reference for the Commissioner notification',
			},
			{
				displayName: 'Commissioner Notified At',
				name: 'commissionerNotifiedAt',
				type: 'dateTime',
				default: '',
				description: 'Date and time the DPO appointment was notified to the Commissioner',
			},
			{
				displayName: 'DPO Appointed At',
				name: 'dpoAppointedAt',
				type: 'dateTime',
				default: '',
				description: 'Date and time the DPO was appointed',
			},
			{
				displayName: 'DPO Profile ID',
				name: 'dpoProfileId',
				type: 'string',
				default: '',
				description: 'Appointed DPO membership profile ID',
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const totalDataSubjects = this.getNodeParameter('totalDataSubjects', itemIndex) as number;
	const sensitiveDataSubjects = this.getNodeParameter('sensitiveDataSubjects', itemIndex) as number;
	const regularSystematicMonitoring = this.getNodeParameter(
		'regularSystematicMonitoring',
		itemIndex,
	) as boolean;
	const additionalFields = this.getNodeParameter('additionalFields', itemIndex, {}) as IDataObject;

	const currentQuery = `
		query GetCurrentMalaysiaPDPAProfile($organizationId: ID!) {
			node(id: $organizationId) {
				... on Organization {
					malaysiaPDPAProfile {
						dpoProfileId
						dpoAppointedAt
						commissionerNotifiedAt
						commissionerNotificationReference
					}
				}
			}
		}
	`;

	const currentResponse = await proboApiRequest.call(this, currentQuery, { organizationId });
	const currentData = currentResponse.data as IDataObject | undefined;
	const currentNode = currentData?.node as IDataObject | undefined;
	const currentProfile = currentNode?.malaysiaPDPAProfile as IDataObject | undefined;

	const input: IDataObject = {
		organizationId,
		totalDataSubjects,
		sensitiveDataSubjects,
		regularSystematicMonitoring,
	};

	const clearDPO = additionalFields.clearDPO === true;
	const clearCommissionerNotification = additionalFields.clearCommissionerNotification === true;

	if (!clearDPO && currentProfile) {
		if (currentProfile.dpoProfileId) input.dpoProfileId = currentProfile.dpoProfileId;
		if (currentProfile.dpoAppointedAt) input.dpoAppointedAt = currentProfile.dpoAppointedAt;
		if (!clearCommissionerNotification && currentProfile.commissionerNotifiedAt) {
			input.commissionerNotifiedAt = currentProfile.commissionerNotifiedAt;
		}
		if (!clearCommissionerNotification && currentProfile.commissionerNotificationReference) {
			input.commissionerNotificationReference =
				currentProfile.commissionerNotificationReference;
		}
	}

	if (!clearDPO && additionalFields.dpoProfileId) {
		input.dpoProfileId = additionalFields.dpoProfileId;
	}
	if (!clearDPO && additionalFields.dpoAppointedAt) {
		input.dpoAppointedAt = additionalFields.dpoAppointedAt;
	}
	if (!clearDPO && !clearCommissionerNotification && additionalFields.commissionerNotifiedAt) {
		input.commissionerNotifiedAt = additionalFields.commissionerNotifiedAt;
	}
	if (
		!clearDPO &&
		!clearCommissionerNotification &&
		additionalFields.commissionerNotificationReference !== undefined
	) {
		input.commissionerNotificationReference =
			additionalFields.commissionerNotificationReference;
	}

	const mutation = `
		mutation UpdateMalaysiaPDPAProfile($input: UpdateMalaysiaPDPAProfileInput!) {
			updateMalaysiaPDPAProfile(input: $input) {
				malaysiaPDPAProfile {
					organizationId
					totalDataSubjects
					sensitiveDataSubjects
					regularSystematicMonitoring
					dpoRequired
					dpoRequirementReasons
					assessedByProfileId
					assessedAt
					dpoProfileId
					dpoAppointedAt
					commissionerNotificationDueAt
					commissionerNotifiedAt
					commissionerNotificationReference
					createdAt
					updatedAt
				}
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, mutation, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
