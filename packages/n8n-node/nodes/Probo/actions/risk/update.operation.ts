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
		displayName: 'Risk ID',
		name: 'riskId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['risk'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the risk to update',
		required: true,
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['risk'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The name of the risk',
	},
	{
		displayName: 'Category',
		name: 'category',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['risk'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The category of the risk',
	},
	{
		displayName: 'Additional Fields',
		name: 'additionalFields',
		type: 'collection',
		placeholder: 'Add Field',
		default: {},
		displayOptions: {
			show: {
				resource: ['risk'],
				operation: ['update'],
			},
		},
		options: [
			{
				displayName: 'Description',
				name: 'description',
				type: 'string',
				default: '',
				description: 'The description of the risk',
			},
			{
				displayName: 'Note',
				name: 'note',
				type: 'string',
				default: '',
				description: 'Additional notes about the risk',
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const riskId = this.getNodeParameter('riskId', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex, '') as string;
	const category = this.getNodeParameter('category', itemIndex, '') as string;
	const additionalFields = this.getNodeParameter('additionalFields', itemIndex, {}) as {
		description?: string;
		note?: string;
	};

	const query = `
		mutation UpdateRisk($input: UpdateRiskInput!) {
			updateRisk(input: $input) {
				risk {
					id
					name
					description
					category
					note
					createdAt
					updatedAt
				}
			}
		}
	`;

	const input: Record<string, unknown> = { id: riskId };
	if (name) input.name = name;
	if (category) input.category = category;
	if (additionalFields.description !== undefined) input.description = additionalFields.description === '' ? null : additionalFields.description;
	if (additionalFields.note !== undefined) input.note = additionalFields.note === '' ? null : additionalFields.note;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
