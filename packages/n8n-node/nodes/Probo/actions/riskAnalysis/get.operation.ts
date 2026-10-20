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
				operation: ['get'],
			},
		},
		default: '',
		description: 'The ID of the risk analysis',
		required: true,
	},
	{
		displayName: 'As Of',
		name: 'asOf',
		type: 'dateTime',
		displayOptions: {
			show: {
				resource: ['riskAnalysis'],
				operation: ['get'],
			},
		},
		default: '',
		description:
			'Reconstruct matrix cells as of this instant. Leave empty to use live tables.',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const riskAnalysisId = this.getNodeParameter('riskAnalysisId', itemIndex) as string;
	const asOf = this.getNodeParameter('asOf', itemIndex, '') as string;

	const query = `
		query GetRiskAnalysis($id: ID!, $asOf: Datetime) {
			node(id: $id) {
				... on RiskAnalysis {
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
					matrixCells(asOf: $asOf) {
						type
						likelihood
						impact
						count
					}
					createdAt
					updatedAt
				}
			}
		}
	`;

	const variables: { id: string; asOf?: string } = {
		id: riskAnalysisId,
	};
	if (asOf) {
		variables.asOf = asOf;
	}

	const responseData = await proboApiRequest.call(this, query, variables);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
