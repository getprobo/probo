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

import type { INodeProperties, IExecuteFunctions, INodeExecutionData, IDataObject } from 'n8n-workflow';
import { plainTextToProseMirrorJSON, proboApiRequest, withPlainTextContent } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Task ID',
		name: 'taskId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['taskComment'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The ID of the task',
		required: true,
	},
	{
		displayName: 'Content',
		name: 'content',
		type: 'string',
		typeOptions: {
			rows: 4,
		},
		displayOptions: {
			show: {
				resource: ['taskComment'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The comment content',
		required: true,
	},
	{
		displayName: 'Owner ID',
		name: 'ownerId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['taskComment'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The owner membership profile ID. Defaults to the authenticated user when omitted.',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const taskId = this.getNodeParameter('taskId', itemIndex) as string;
	const content = this.getNodeParameter('content', itemIndex) as string;
	const ownerId = this.getNodeParameter('ownerId', itemIndex, '') as string;

	const query = `
		mutation CreateTaskComment($input: CreateTaskCommentInput!) {
			createTaskComment(input: $input) {
				taskCommentEdge {
					node {
						id
						content
						createdAt
						updatedAt
						owner {
							id
							fullName
						}
					}
				}
			}
		}
	`;

	const input: Record<string, string> = { taskId, content: plainTextToProseMirrorJSON(content) };
	if (ownerId) {
		input.ownerId = ownerId;
	}

	const responseData = await proboApiRequest.call(this, query, { input });
	const data = responseData.data as IDataObject | undefined;
	const payload = data?.createTaskComment as IDataObject | undefined;
	const edge = payload?.taskCommentEdge as IDataObject | undefined;
	const node = edge?.node as IDataObject | undefined;
	if (edge && node) {
		edge.node = withPlainTextContent(node);
	}

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
