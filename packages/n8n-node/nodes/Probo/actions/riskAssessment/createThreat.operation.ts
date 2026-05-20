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
				operation: ['createThreat'],
			},
		},
		default: '',
		description: 'The ID of the scope',
		required: true,
	},
	{
		displayName: 'Process ID',
		name: 'processId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['riskAssessment'],
				operation: ['createThreat'],
			},
		},
		default: '',
		description: 'The ID of the process',
		required: true,
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['riskAssessment'],
				operation: ['createThreat'],
			},
		},
		default: '',
		description: 'The name of the threat',
		required: true,
	},
	{
		displayName: 'Category',
		name: 'category',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['riskAssessment'],
				operation: ['createThreat'],
			},
		},
		default: '',
		description: 'The category of the threat',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const riskAssessmentScopeId = this.getNodeParameter('riskAssessmentScopeId', itemIndex) as string;
	const processId = this.getNodeParameter('processId', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex) as string;
	const category = this.getNodeParameter('category', itemIndex) as string;

	const query = `
		mutation CreateRiskAssessmentThreat($input: CreateRiskAssessmentThreatInput!) {
			createRiskAssessmentThreat(input: $input) {
				riskAssessmentThreatEdge {
					node {
						id
						riskAssessmentScopeId
						processId
						name
						category
						createdAt
						updatedAt
					}
				}
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, {
		input: { riskAssessmentScopeId, processId, name, category },
	});

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
