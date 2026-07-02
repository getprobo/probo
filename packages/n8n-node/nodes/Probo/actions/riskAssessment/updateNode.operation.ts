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
