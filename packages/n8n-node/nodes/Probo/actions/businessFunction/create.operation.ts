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

import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Organization ID',
		name: 'organizationId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['businessFunction'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The ID of the organization',
		required: true,
	},
	{
		displayName: 'Reference ID',
		name: 'referenceId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['businessFunction'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The reference ID of the business function',
		required: true,
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['businessFunction'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The name of the business function',
		required: true,
	},
	{
		displayName: 'Classification',
		name: 'classification',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['businessFunction'],
				operation: ['create'],
			},
		},
		options: [
			{ name: 'Critical', value: 'CRITICAL' },
			{ name: 'Important', value: 'IMPORTANT' },
			{ name: 'Secondary', value: 'SECONDARY' },
			{ name: 'Standard', value: 'STANDARD' },
		],
		default: 'STANDARD',
		description: 'The classification of the business function',
		required: true,
	},
	{
		displayName: 'MTD Minutes',
		name: 'mtdMinutes',
		type: 'number',
		displayOptions: {
			show: {
				resource: ['businessFunction'],
				operation: ['create'],
			},
		},
		default: 0,
		description: 'Maximum tolerable downtime in minutes',
		required: true,
	},
	{
		displayName: 'RTO Minutes',
		name: 'rtoMinutes',
		type: 'number',
		displayOptions: {
			show: {
				resource: ['businessFunction'],
				operation: ['create'],
			},
		},
		default: 0,
		description: 'Recovery time objective in minutes',
		required: true,
	},
	{
		displayName: 'RPO Minutes',
		name: 'rpoMinutes',
		type: 'number',
		displayOptions: {
			show: {
				resource: ['businessFunction'],
				operation: ['create'],
			},
		},
		default: 0,
		description: 'Recovery point objective in minutes',
		required: true,
	},
	{
		displayName: 'Additional Fields',
		name: 'additionalFields',
		type: 'collection',
		placeholder: 'Add Field',
		default: {},
		displayOptions: {
			show: {
				resource: ['businessFunction'],
				operation: ['create'],
			},
		},
		options: [
			{
				displayName: 'Asset IDs',
				name: 'assetIds',
				type: 'string',
				default: '',
				description: 'Comma-separated asset IDs',
			},
			{
				displayName: 'Impact Tolerance',
				name: 'impactTolerance',
				type: 'string',
				default: '',
				description: 'Impact tolerance description',
			},
			{
				displayName: 'Notes',
				name: 'notes',
				type: 'string',
				default: '',
				description: 'Additional notes',
			},
			{
				displayName: 'Owner ID',
				name: 'ownerId',
				type: 'string',
				default: '',
				description: 'Owner profile ID',
			},
			{
				displayName: 'Third Party IDs',
				name: 'thirdPartyIds',
				type: 'string',
				default: '',
				description: 'Comma-separated third party IDs',
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const referenceId = this.getNodeParameter('referenceId', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex) as string;
	const classification = this.getNodeParameter('classification', itemIndex) as string;
	const mtdMinutes = this.getNodeParameter('mtdMinutes', itemIndex) as number;
	const rtoMinutes = this.getNodeParameter('rtoMinutes', itemIndex) as number;
	const rpoMinutes = this.getNodeParameter('rpoMinutes', itemIndex) as number;
	const additionalFields = this.getNodeParameter('additionalFields', itemIndex, {}) as {
		impactTolerance?: string;
		notes?: string;
		ownerId?: string;
		assetIds?: string;
		thirdPartyIds?: string;
	};

	const query = `
		mutation CreateBusinessFunction($input: CreateBusinessFunctionInput!) {
			createBusinessFunction(input: $input) {
				businessFunctionEdge {
					node {
						id
						referenceId
						name
						classification
						mtdMinutes
						rtoMinutes
						rpoMinutes
						impactTolerance
						notes
						createdAt
						updatedAt
					}
				}
			}
		}
	`;

	const input: Record<string, unknown> = {
		organizationId,
		referenceId,
		name,
		classification,
		mtdMinutes,
		rtoMinutes,
		rpoMinutes,
	};

	if (additionalFields.impactTolerance) {
		input.impactTolerance = additionalFields.impactTolerance;
	}

	if (additionalFields.notes) {
		input.notes = additionalFields.notes;
	}

	if (additionalFields.ownerId) {
		input.ownerId = additionalFields.ownerId;
	}

	if (additionalFields.assetIds) {
		input.assetIds = additionalFields.assetIds
			.split(',')
			.map((id) => id.trim())
			.filter(Boolean);
	}

	if (additionalFields.thirdPartyIds) {
		input.thirdPartyIds = additionalFields.thirdPartyIds
			.split(',')
			.map((id) => id.trim())
			.filter(Boolean);
	}

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
