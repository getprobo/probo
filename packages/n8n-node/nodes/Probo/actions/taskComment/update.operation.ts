// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
import { organizationLocator, personLocator } from '../../locators';

export const description: INodeProperties[] = [
	organizationLocator({
		resource: ['taskComment'],
		operation: ['update'],
		required: false,
		description: 'Organization to search in. Leave this empty when entering an ID.',
	}),
	{
		displayName: 'Comment ID',
		name: 'commentId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['taskComment'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the task comment to update',
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
				resource: ['taskComment'],
				operation: ['update'],
			},
		},
		options: [
			{
				displayName: 'Content',
				name: 'content',
				type: 'string',
				typeOptions: {
					rows: 6,
				},
				default: '',
				description: 'The comment content as a ProseMirror document JSON string. An image node sets fileId from an attachment upload. A file node sets fileId, fileName, mimeType, and size',
			},
			personLocator({
				displayName: 'Owner',
				required: false,
			}),
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const commentId = this.getNodeParameter('commentId', itemIndex) as string;
	const additionalFields = this.getNodeParameter('additionalFields', itemIndex, {}) as {
		content?: string;
		ownerId?: string;
	};

	const query = `
		mutation UpdateTaskComment($input: UpdateTaskCommentInput!) {
			updateTaskComment(input: $input) {
				taskComment {
					id
					content
					createdAt
					updatedAt
					owner {
						id
						fullName
						emailAddress
					}
				}
			}
		}
	`;

	const input: Record<string, string | null> = { taskCommentId: commentId };
	if (additionalFields.content !== undefined) {
		input.content = additionalFields.content === ''
			? null
			: additionalFields.content;
	}
	if (this.getNodeParameter('additionalFields.ownerId', itemIndex, '', { extractValue: true }) as string) {
		input.ownerId = this.getNodeParameter('additionalFields.ownerId', itemIndex, '', { extractValue: true }) as string;
	}

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
