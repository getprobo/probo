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
		displayName: 'Node ID',
		name: 'nodeId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['riskAssessment'],
				operation: ['updateNode'],
			},
		},
		default: '',
		description: 'The ID of the node to update',
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
				resource: ['riskAssessment'],
				operation: ['updateNode'],
			},
		},
		options: [
			{
				displayName: 'Name',
				name: 'name',
				type: 'string',
				default: '',
				description: 'The name of the node',
			},
			{
				displayName: 'Node Type',
				name: 'nodeType',
				type: 'options',
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
			},
			{
				displayName: 'Boundary ID',
				name: 'boundaryId',
				type: 'string',
				default: '',
				description: 'The ID of the boundary that contains this node. Leave empty to move it to the top level.',
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const nodeId = this.getNodeParameter('nodeId', itemIndex) as string;
	const additionalFields = this.getNodeParameter('additionalFields', itemIndex, {}) as {
		name?: string;
		nodeType?: string;
		boundaryId?: string;
	};

	const query = `
		mutation UpdateRiskAssessmentNode($input: UpdateRiskAssessmentNodeInput!) {
			updateRiskAssessmentNode(input: $input) {
				riskAssessmentNode {
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
	`;

	const input: Record<string, unknown> = { id: nodeId };
	if (additionalFields.name) input.name = additionalFields.name;
	if (additionalFields.nodeType) input.nodeType = additionalFields.nodeType;
	if (additionalFields.boundaryId !== undefined) {
		input.boundaryId = additionalFields.boundaryId || null;
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
