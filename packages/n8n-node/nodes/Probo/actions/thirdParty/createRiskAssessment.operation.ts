// Copyright (c) 2025 Probo Inc <hello@probo.com>.
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
		displayName: 'ThirdParty ID',
		name: 'thirdPartyId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['thirdParty'],
				operation: ['createRiskAssessment'],
			},
		},
		default: '',
		description: 'The ID of the thirdParty',
		required: true,
	},
	{
		displayName: 'Expires At',
		name: 'expiresAt',
		type: 'dateTime',
		displayOptions: {
			show: {
				resource: ['thirdParty'],
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
				resource: ['thirdParty'],
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
				resource: ['thirdParty'],
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
				resource: ['thirdParty'],
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
	const thirdPartyId = this.getNodeParameter('thirdPartyId', itemIndex) as string;
	const expiresAtRaw = this.getNodeParameter('expiresAt', itemIndex) as string;
	const dataSensitivity = this.getNodeParameter('dataSensitivity', itemIndex) as string;
	const businessImpact = this.getNodeParameter('businessImpact', itemIndex) as string;
	const notes = this.getNodeParameter('notes', itemIndex, '') as string;

	// Ensure expiresAt is in RFC3339 format
	const expiresAt = new Date(expiresAtRaw).toISOString();

	const query = `
		mutation CreateThirdPartyRiskAssessment($input: CreateThirdPartyRiskAssessmentInput!) {
			createThirdPartyRiskAssessment(input: $input) {
				thirdPartyRiskAssessmentEdge {
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
		thirdPartyId,
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
