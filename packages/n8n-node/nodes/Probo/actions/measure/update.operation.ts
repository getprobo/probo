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
		displayName: 'Measure ID',
		name: 'id',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['measure'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the measure to update',
		required: true,
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['measure'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The name of the measure',
	},
	{
		displayName: 'Description',
		name: 'description',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['measure'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The description of the measure',
	},
	{
		displayName: 'Category',
		name: 'category',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['measure'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The category of the measure',
	},
	{
		displayName: 'State',
		name: 'state',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['measure'],
				operation: ['update'],
			},
		},
		options: [
			{
				name: 'Not Started',
				value: 'NOT_STARTED',
			},
			{
				name: 'In Progress',
				value: 'IN_PROGRESS',
			},
			{
				name: 'Not Applicable',
				value: 'NOT_APPLICABLE',
			},
			{
				name: 'Implemented',
				value: 'IMPLEMENTED',
			},
		],
		default: 'NOT_STARTED',
		description: 'The state of the measure',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const id = this.getNodeParameter('id', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex, '') as string;
	const description = this.getNodeParameter('description', itemIndex, '') as string;
	const category = this.getNodeParameter('category', itemIndex, '') as string;
	const state = this.getNodeParameter('state', itemIndex, '') as string;

	const query = `
		mutation UpdateMeasure($input: UpdateMeasureInput!) {
			updateMeasure(input: $input) {
				measure {
					id
					name
					description
					category
					state
					createdAt
					updatedAt
				}
			}
		}
	`;

	const input: Record<string, string> = { id };
	if (name) input.name = name;
	if (description) input.description = description;
	if (category) input.category = category;
	if (state) input.state = state;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
