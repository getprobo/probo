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
		displayName: 'Risk Analysis ID',
		name: 'riskAnalysisId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['riskAnalysis'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the risk analysis to update',
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
				resource: ['riskAnalysis'],
				operation: ['update'],
			},
		},
		options: [
			{
				displayName: 'Description',
				name: 'description',
				type: 'string',
				default: '',
				description: 'The description of the risk analysis',
			},
			{
				displayName: 'Name',
				name: 'name',
				type: 'string',
				default: '',
				description: 'The name of the risk analysis',
			},
			{
				displayName: 'Period End',
				name: 'periodEnd',
				type: 'dateTime',
				default: '',
				description: 'End of the analysis period',
			},
			{
				displayName: 'Period Start',
				name: 'periodStart',
				type: 'dateTime',
				default: '',
				description: 'Start of the analysis period',
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const riskAnalysisId = this.getNodeParameter('riskAnalysisId', itemIndex) as string;
	const additionalFields = this.getNodeParameter('additionalFields', itemIndex, {}) as {
		name?: string;
		description?: string;
		periodStart?: string;
		periodEnd?: string;
	};

	const query = `
		mutation UpdateRiskAnalysis($input: UpdateRiskAnalysisInput!) {
			updateRiskAnalysis(input: $input) {
				riskAnalysis {
					id
					name
					description
					period {
						start
						end
					}
					matrixSize {
						rows
						cols
					}
					createdAt
					updatedAt
				}
			}
		}
	`;

	const input: Record<string, unknown> = { id: riskAnalysisId };
	if (additionalFields.name) input.name = additionalFields.name;
	if (additionalFields.description !== undefined) input.description = additionalFields.description === '' ? null : additionalFields.description;
	if (additionalFields.periodStart || additionalFields.periodEnd) {
		input.period = {
			...(additionalFields.periodStart ? { start: additionalFields.periodStart } : {}),
			...(additionalFields.periodEnd ? { end: additionalFields.periodEnd } : {}),
		};
	}

	if (Object.keys(input).length === 1) {
		throw new Error('At least one field must be provided to update');
	}

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
