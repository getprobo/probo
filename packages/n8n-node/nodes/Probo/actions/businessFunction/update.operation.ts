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
		displayName: 'Business Function ID',
		name: 'id',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['businessFunction'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the business function to update',
		required: true,
	},
	{
		displayName: 'Reference ID',
		name: 'referenceId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['businessFunction'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The reference ID of the business function',
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['businessFunction'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The name of the business function',
	},
	{
		displayName: 'Classification',
		name: 'classification',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['businessFunction'],
				operation: ['update'],
			},
		},
		options: [
			{ name: '(Unchanged)', value: '' },
			{ name: 'Critical', value: 'CRITICAL' },
			{ name: 'Important', value: 'IMPORTANT' },
			{ name: 'Secondary', value: 'SECONDARY' },
			{ name: 'Standard', value: 'STANDARD' },
		],
		default: '',
		description: 'The classification of the business function',
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
				operation: ['update'],
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
				description: 'Impact tolerance description. Empty string clears the value.',
			},
			{
				displayName: 'MTD Minutes',
				name: 'mtdMinutes',
				type: 'number',
				default: 0,
				description: 'Maximum tolerable downtime in minutes. Include this field to update; 0 is a valid value.',
			},
			{
				displayName: 'Notes',
				name: 'notes',
				type: 'string',
				default: '',
				description: 'Additional notes. Empty string clears the value.',
			},
			{
				displayName: 'Owner ID',
				name: 'ownerId',
				type: 'string',
				default: '',
				description: 'Owner profile ID. Empty string clears the value.',
			},
			{
				displayName: 'RPO Minutes',
				name: 'rpoMinutes',
				type: 'number',
				default: 0,
				description: 'Recovery point objective in minutes. Include this field to update; 0 is a valid value.',
			},
			{
				displayName: 'RTO Minutes',
				name: 'rtoMinutes',
				type: 'number',
				default: 0,
				description: 'Recovery time objective in minutes. Include this field to update; 0 is a valid value.',
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
	const id = this.getNodeParameter('id', itemIndex) as string;
	const referenceId = this.getNodeParameter('referenceId', itemIndex, '') as string;
	const name = this.getNodeParameter('name', itemIndex, '') as string;
	const classification = this.getNodeParameter('classification', itemIndex, '') as string;
	const additionalFields = this.getNodeParameter('additionalFields', itemIndex, {}) as {
		mtdMinutes?: number;
		rtoMinutes?: number;
		rpoMinutes?: number;
		impactTolerance?: string;
		notes?: string;
		ownerId?: string;
		assetIds?: string;
		thirdPartyIds?: string;
	};

	const query = `
		mutation UpdateBusinessFunction($input: UpdateBusinessFunctionInput!) {
			updateBusinessFunction(input: $input) {
				businessFunction {
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
	`;

	const input: Record<string, unknown> = { id };

	if (referenceId) {
		input.referenceId = referenceId;
	}

	if (name) {
		input.name = name;
	}

	if (classification) {
		input.classification = classification;
	}

	if (additionalFields.mtdMinutes !== undefined) {
		input.mtdMinutes = additionalFields.mtdMinutes;
	}

	if (additionalFields.rtoMinutes !== undefined) {
		input.rtoMinutes = additionalFields.rtoMinutes;
	}

	if (additionalFields.rpoMinutes !== undefined) {
		input.rpoMinutes = additionalFields.rpoMinutes;
	}

	if (additionalFields.impactTolerance !== undefined) {
		input.impactTolerance = additionalFields.impactTolerance === '' ? null : additionalFields.impactTolerance;
	}

	if (additionalFields.notes !== undefined) {
		input.notes = additionalFields.notes === '' ? null : additionalFields.notes;
	}

	if (additionalFields.ownerId !== undefined) {
		input.ownerId = additionalFields.ownerId === '' ? null : additionalFields.ownerId;
	}

	if (additionalFields.assetIds !== undefined) {
		input.assetIds = additionalFields.assetIds
			.split(',')
			.map((assetId) => assetId.trim())
			.filter(Boolean);
	}

	if (additionalFields.thirdPartyIds !== undefined) {
		input.thirdPartyIds = additionalFields.thirdPartyIds
			.split(',')
			.map((thirdPartyId) => thirdPartyId.trim())
			.filter(Boolean);
	}

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
