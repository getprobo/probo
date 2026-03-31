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

import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Vendor ID',
		name: 'vendorId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['vendor'],
				operation: ['createRiskAssessment'],
			},
		},
		default: '',
		description: 'The ID of the vendor',
		required: true,
	},
	{
		displayName: 'Expires At',
		name: 'expiresAt',
		type: 'dateTime',
		displayOptions: {
			show: {
				resource: ['vendor'],
				operation: ['createRiskAssessment'],
			},
		},
		default: '',
		description: 'The expiration date of the risk assessment',
		required: true,
	},
	{
		displayName: 'Data Sensitivity',
		name: 'dataSensitivity',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['vendor'],
				operation: ['createRiskAssessment'],
			},
		},
		options: [
			{ name: 'Critical', value: 'CRITICAL' },
			{ name: 'High', value: 'HIGH' },
			{ name: 'Low', value: 'LOW' },
			{ name: 'Medium', value: 'MEDIUM' },
			{ name: 'None', value: 'NONE' },
		],
		default: 'LOW',
		description: 'The data sensitivity level',
		required: true,
	},
	{
		displayName: 'Business Impact',
		name: 'businessImpact',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['vendor'],
				operation: ['createRiskAssessment'],
			},
		},
		options: [
			{ name: 'Critical', value: 'CRITICAL' },
			{ name: 'High', value: 'HIGH' },
			{ name: 'Low', value: 'LOW' },
			{ name: 'Medium', value: 'MEDIUM' },
		],
		default: 'LOW',
		description: 'The business impact level',
		required: true,
	},
	{
		displayName: 'Notes',
		name: 'notes',
		type: 'string',
		typeOptions: {
			rows: 4,
		},
		displayOptions: {
			show: {
				resource: ['vendor'],
				operation: ['createRiskAssessment'],
			},
		},
		default: '',
		description: 'Additional notes for the risk assessment',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const vendorId = this.getNodeParameter('vendorId', itemIndex) as string;
	const expiresAtRaw = this.getNodeParameter('expiresAt', itemIndex) as string;
	const dataSensitivity = this.getNodeParameter('dataSensitivity', itemIndex) as string;
	const businessImpact = this.getNodeParameter('businessImpact', itemIndex) as string;
	const notes = this.getNodeParameter('notes', itemIndex, '') as string;

	// Ensure expiresAt is in RFC3339 format
	const expiresAt = new Date(expiresAtRaw).toISOString();

	const query = `
		mutation CreateVendorRiskAssessment($input: CreateVendorRiskAssessmentInput!) {
			createVendorRiskAssessment(input: $input) {
				vendorRiskAssessmentEdge {
					node {
						id
						expiresAt
						dataSensitivity
						businessImpact
						notes
						createdAt
						updatedAt
					}
				}
			}
		}
	`;

	const input: Record<string, unknown> = {
		vendorId,
		expiresAt,
		dataSensitivity,
		businessImpact,
	};
	if (notes) input.notes = notes;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
