// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
