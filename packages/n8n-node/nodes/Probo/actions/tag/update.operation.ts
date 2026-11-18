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
import { NodeOperationError } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Tag ID',
		name: 'id',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['tag'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the tag to update',
		required: true,
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['tag'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'Tag name; unique within the organization',
	},
	{
		displayName: 'Color',
		name: 'color',
		type: 'color',
		displayOptions: {
			show: {
				resource: ['tag'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'Optional hex color (#RGB or #RRGGBB)',
	},
	{
		displayName: 'Clear Color',
		name: 'clearColor',
		type: 'boolean',
		displayOptions: {
			show: {
				resource: ['tag'],
				operation: ['update'],
			},
		},
		default: false,
		description: 'Whether to remove the color from the tag',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const id = this.getNodeParameter('id', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex, '') as string;
	const color = this.getNodeParameter('color', itemIndex, '') as string;
	const clearColor = this.getNodeParameter('clearColor', itemIndex, false) as boolean;

	if (!name && !color && !clearColor) {
		throw new NodeOperationError(
			this.getNode(),
			'Provide a name, a color, or Clear Color to update the tag',
			{ itemIndex },
		);
	}

	const query = `
		mutation UpdateTag($input: UpdateTagInput!) {
			updateTag(input: $input) {
				tag {
					id
					name
					color
					createdAt
					updatedAt
				}
			}
		}
	`;

	const input: Record<string, unknown> = { id };
	if (name) input.name = name;
	if (clearColor) input.color = null;
	else if (color) input.color = color;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
