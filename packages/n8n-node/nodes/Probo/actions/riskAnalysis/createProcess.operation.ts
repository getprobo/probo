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
		displayName: 'Scope ID',
		name: 'riskAnalysisScopeId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['riskAnalysis'],
				operation: ['createProcess'],
			},
		},
		default: '',
		description: 'The ID of the scope',
		required: true,
	},
	{
		displayName: 'Source Node ID',
		name: 'sourceNodeId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['riskAnalysis'],
				operation: ['createProcess'],
			},
		},
		default: '',
		description: 'The ID of the source node',
		required: true,
	},
	{
		displayName: 'Target Node ID',
		name: 'targetNodeId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['riskAnalysis'],
				operation: ['createProcess'],
			},
		},
		default: '',
		description: 'The ID of the target node',
		required: true,
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['riskAnalysis'],
				operation: ['createProcess'],
			},
		},
		default: '',
		description: 'The name of the process',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const riskAnalysisScopeId = this.getNodeParameter('riskAnalysisScopeId', itemIndex) as string;
	const sourceNodeId = this.getNodeParameter('sourceNodeId', itemIndex) as string;
	const targetNodeId = this.getNodeParameter('targetNodeId', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex) as string;

	const query = `
		mutation CreateRiskAnalysisProcess($input: CreateRiskAnalysisProcessInput!) {
			createRiskAnalysisProcess(input: $input) {
				riskAnalysisProcessEdge {
					node {
						id
						riskAnalysisScopeId
						sourceNodeId
						targetNodeId
						name
						createdAt
						updatedAt
					}
				}
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, {
		input: { riskAnalysisScopeId, sourceNodeId, targetNodeId, name },
	});

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
