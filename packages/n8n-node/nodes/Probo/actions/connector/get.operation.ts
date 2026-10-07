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

import type { IDataObject, IExecuteFunctions, INodeExecutionData, INodeProperties } from 'n8n-workflow';
import { proboApiRequest, proboApiRequestAllItems } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Connector ID',
		name: 'connectorId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['connector'],
				operation: ['get'],
			},
		},
		default: '',
		description: 'The ID of the connector',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const connectorId = this.getNodeParameter('connectorId', itemIndex) as string;

	const query = `
		query GetConnector($connectorId: ID!) {
			node(id: $connectorId) {
				... on Connector {
					id
					provider
					protocol
					createdAt
				}
			}
		}
	`;

	const accountsQuery = `
		query GetConnectorAccounts($connectorId: ID!, $first: Int, $after: CursorKey) {
			node(id: $connectorId) {
				... on Connector {
					accounts(first: $first, after: $after) {
						edges {
							node {
								id
								externalAccountId
								name
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

	const responseData = await proboApiRequest.call(this, query, { connectorId });
	const data = responseData.data as IDataObject | undefined;
	const node = data?.node as IDataObject | undefined;
	if (!node?.provider) {
		return {
			json: responseData,
			pairedItem: { item: itemIndex },
		};
	}

	const accounts = await proboApiRequestAllItems.call(
		this,
		accountsQuery,
		{ connectorId },
		(response) => {
			const data = response?.data as IDataObject | undefined;
			const node = data?.node as IDataObject | undefined;
			return node?.accounts as IDataObject | undefined;
		},
	);

	node.accounts = {
		totalCount: accounts.length,
		edges: accounts.map((account) => ({ node: account })),
	};

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
