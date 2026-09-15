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

import type {
	IDataObject,
	IExecuteFunctions,
	INodeExecutionData,
	INodeProperties,
} from 'n8n-workflow';
import { proboConnectApiRequestAllItems } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Service Account ID',
		name: 'serviceAccountId',
		type: 'string',
		displayOptions: {
			show: { resource: ['serviceAccount'], operation: ['listCredentials'] },
		},
		default: '',
		description: 'The ID of the service account',
		required: true,
	},
	{
		displayName: 'Return All',
		name: 'returnAll',
		type: 'boolean',
		displayOptions: {
			show: { resource: ['serviceAccount'], operation: ['listCredentials'] },
		},
		default: false,
		description: 'Whether to return all results or only up to a given limit',
	},
	{
		displayName: 'Limit',
		name: 'limit',
		type: 'number',
		displayOptions: {
			show: {
				resource: ['serviceAccount'],
				operation: ['listCredentials'],
				returnAll: [false],
			},
		},
		typeOptions: { minValue: 1 },
		default: 50,
		description: 'Max number of results to return',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const serviceAccountId = this.getNodeParameter('serviceAccountId', itemIndex) as string;
	const returnAll = this.getNodeParameter('returnAll', itemIndex) as boolean;
	const limit = this.getNodeParameter('limit', itemIndex, 50) as number;
	const query = `
		query ListServiceAccountCredentials(
			$serviceAccountId: ID!
			$first: Int
			$after: CursorKey
		) {
			node(id: $serviceAccountId) {
				... on ServiceAccount {
					credentials(first: $first, after: $after) {
						edges {
							node {
								id
								serviceAccountId
								name
								scopes
								expiresAt
								lastUsedAt
								revokedAt
								createdAt
							}
						}
						pageInfo {
							hasNextPage
							endCursor
						}
					}
				}
			}
		}
	`;

	const serviceAccountCredentials = await proboConnectApiRequestAllItems.call(
		this,
		query,
		{ serviceAccountId },
		(response) => {
			const data = response.data as IDataObject | undefined;
			const node = data?.node as IDataObject | undefined;
			return node?.credentials as IDataObject | undefined;
		},
		returnAll,
		limit,
	);

	return {
		json: { serviceAccountCredentials },
		pairedItem: { item: itemIndex },
	};
}
