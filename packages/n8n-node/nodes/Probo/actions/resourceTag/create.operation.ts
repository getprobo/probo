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
		displayName: 'Organization ID',
		name: 'organizationId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['resourceTag'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The ID of the organization',
		required: true,
	},
	{
		displayName: 'Key',
		name: 'key',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['resourceTag'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'Unique slug key for the tag within the organization',
		required: true,
	},
	{
		displayName: 'Value',
		name: 'value',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['resourceTag'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'Display value for the tag',
		required: true,
	},
	{
		displayName: 'Color',
		name: 'color',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['resourceTag'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'Optional hex color (#RGB or #RRGGBB)',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const key = this.getNodeParameter('key', itemIndex) as string;
	const value = this.getNodeParameter('value', itemIndex) as string;
	const color = this.getNodeParameter('color', itemIndex, '') as string;

	const query = `
		mutation CreateResourceTag($input: CreateResourceTagInput!) {
			createResourceTag(input: $input) {
				resourceTagEdge {
					node {
						id
						key
						value
						color
						createdAt
						updatedAt
					}
				}
			}
		}
	`;

	const input: Record<string, unknown> = {
		organizationId,
		key,
		value,
	};
	if (color) input.color = color;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
