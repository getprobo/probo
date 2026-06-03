// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
				resource: ['riskAssessment'],
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
				resource: ['riskAssessment'],
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
				resource: ['riskAssessment'],
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
	const riskAssessmentScopeId = this.getNodeParameter('riskAssessmentScopeId', itemIndex) as string;
	const sourceNodeId = this.getNodeParameter('sourceNodeId', itemIndex) as string;
	const targetNodeId = this.getNodeParameter('targetNodeId', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex) as string;

	const query = `
		mutation CreateRiskAssessmentProcess($input: CreateRiskAssessmentProcessInput!) {
			createRiskAssessmentProcess(input: $input) {
				riskAssessmentProcessEdge {
					node {
						id
						riskAssessmentScopeId
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
		input: { riskAssessmentScopeId, sourceNodeId, targetNodeId, name },
	});

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
