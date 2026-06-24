// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
