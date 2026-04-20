// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
		displayName: 'Framework ID',
		name: 'frameworkId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['control'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The ID of the framework',
		required: true,
	},
	{
		displayName: 'Section Title',
		name: 'sectionTitle',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['control'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The section title of the control',
		required: true,
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['control'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The name of the control',
		required: true,
	},
	{
		displayName: 'Description',
		name: 'description',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['control'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The description of the control',
	},
	{
		displayName: 'Maturity Level',
		name: 'maturityLevel',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['control'],
				operation: ['create'],
			},
		},
		options: [
			{ name: '0 - None', value: 'NONE' },
			{ name: '1 - Initial', value: 'INITIAL' },
			{ name: '2 - Managed', value: 'MANAGED' },
			{ name: '3 - Defined', value: 'DEFINED' },
			{ name: '4 - Quantitatively Managed', value: 'QUANTITATIVELY_MANAGED' },
			{ name: '5 - Optimizing', value: 'OPTIMIZING' },
			{ name: 'Not Set', value: '' },
		],
		default: '',
		description: 'CMMI 0-5 maturity level (optional)',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const frameworkId = this.getNodeParameter('frameworkId', itemIndex) as string;
	const sectionTitle = this.getNodeParameter('sectionTitle', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex) as string;
	const description = this.getNodeParameter('description', itemIndex, '') as string;
	const maturityLevel = this.getNodeParameter('maturityLevel', itemIndex, '') as string;

	const query = `
		mutation CreateControl($input: CreateControlInput!) {
			createControl(input: $input) {
				controlEdge {
					node {
						id
						sectionTitle
						name
						description
						maturityLevel
						createdAt
						updatedAt
					}
				}
			}
		}
	`;

	const variables = {
		input: {
			frameworkId,
			sectionTitle,
			name,
			bestPractice: true,
			...(description && { description }),
			...(maturityLevel && { maturityLevel }),
		},
	};

	const responseData = await proboApiRequest.call(this, query, variables);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
