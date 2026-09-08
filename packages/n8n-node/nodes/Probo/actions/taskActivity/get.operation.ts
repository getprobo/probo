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
import { NodeOperationError } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Activity ID',
		name: 'activityId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['taskActivity'],
				operation: ['get'],
			},
		},
		default: '',
		description: 'The ID of the task activity',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const activityId = this.getNodeParameter('activityId', itemIndex) as string;

	const query = `
		query GetTaskActivity($activityId: ID!) {
			node(id: $activityId) {
				__typename
				... on TaskActivity {
					id
					activityType
					field
					oldValue
					newValue
					createdAt
					actor {
						id
						fullName
					}
				}
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, { activityId });
	const data = responseData.data as IDataObject | undefined;
	const node = data?.node as IDataObject | undefined;
	if (!node) {
		throw new NodeOperationError(
			this.getNode(),
			`Task activity ${activityId} not found`,
		);
	}

	if (node.__typename !== 'TaskActivity') {
		throw new NodeOperationError(
			this.getNode(),
			`Expected TaskActivity node for ${activityId}, got ${String(node.__typename)}`,
		);
	}

	return {
		json: node,
		pairedItem: { item: itemIndex },
	};
}
