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
		name: 'riskAssessmentScopeId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['riskAssessment'],
				operation: ['createNode'],
			},
		},
		default: '',
		description: 'The ID of the scope',
		required: true,
	},
	{
		displayName: 'Node Type',
		name: 'nodeType',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['riskAssessment'],
				operation: ['createNode'],
			},
		},
		options: [
			{
				name: 'Entity',
				value: 'ENTITY',
			},
			{
				name: 'Asset',
				value: 'ASSET',
			},
			{
				name: 'Data',
				value: 'DATA',
			},
		],
		default: 'ENTITY',
		description: 'The type of the node',
		required: true,
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['riskAssessment'],
				operation: ['createNode'],
			},
		},
		default: '',
		description: 'The name of the node',
		required: true,
	},
	{
		displayName: 'Boundary ID',
		name: 'boundaryId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['riskAssessment'],
				operation: ['createNode'],
			},
		},
		default: '',
		description: 'The ID of the boundary that contains this node (optional)',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const riskAssessmentScopeId = this.getNodeParameter('riskAssessmentScopeId', itemIndex) as string;
	const nodeType = this.getNodeParameter('nodeType', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex) as string;
	const boundaryId = this.getNodeParameter('boundaryId', itemIndex, '') as string;

	const query = `
		mutation CreateRiskAssessmentNode($input: CreateRiskAssessmentNodeInput!) {
			createRiskAssessmentNode(input: $input) {
				riskAssessmentNodeEdge {
					node {
						id
						riskAssessmentScopeId
						boundaryId
						nodeType
						name
						createdAt
						updatedAt
					}
				}
			}
		}
	`;

	const input: Record<string, unknown> = { riskAssessmentScopeId, nodeType, name };
	if (boundaryId) input.boundaryId = boundaryId;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
