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
				operation: ['publish'],
			},
		},
		default: '',
		description: 'The ID of the risk analysis to publish',
		required: true,
	},
	{
		displayName: 'Approver IDs',
		name: 'approverIds',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['riskAnalysis'],
				operation: ['publish'],
			},
		},
		default: '',
		description: 'Comma-separated list of approver profile IDs',
	},
	{
		displayName: 'Minor',
		name: 'minor',
		type: 'boolean',
		displayOptions: {
			show: {
				resource: ['riskAnalysis'],
				operation: ['publish'],
			},
		},
		default: false,
		description: 'Whether to publish as a minor version. Approvers are ignored when set. A first release publishes as 0.1; later minors increment the current minor.',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const riskAnalysisId = this.getNodeParameter('riskAnalysisId', itemIndex) as string;
	const approverIds = this.getNodeParameter('approverIds', itemIndex, '') as string;
	const minor = this.getNodeParameter('minor', itemIndex, false) as boolean;

	const query = `
		mutation PublishRiskAnalysis($input: PublishRiskAnalysisInput!) {
			publishRiskAnalysis(input: $input) {
				documentEdge {
					node {
						id
						status
						currentPublishedMajor
						currentPublishedMinor
						createdAt
						updatedAt
					}
				}
				documentVersionEdge {
					node {
						id
						title
						major
						minor
						status
						classification
						documentType
						publishedAt
						createdAt
						updatedAt
					}
				}
			}
		}
	`;

	const input: Record<string, unknown> = { riskAnalysisId, minor };

	if (approverIds) {
		input.approverIds = approverIds
			.split(',')
			.map(id => id.trim())
			.filter(Boolean);
	}

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
