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
import { proboApiRequest, proboApiRequestAllItems } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Snapshot ID',
		name: 'snapshotId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['assetSnapshot'],
				operation: ['get'],
			},
		},
		default: '',
		description: 'The ID of the asset snapshot',
		required: true,
	},
	{
		displayName: 'Return All',
		name: 'returnAll',
		type: 'boolean',
		displayOptions: {
			show: {
				resource: ['assetSnapshot'],
				operation: ['get'],
			},
		},
		default: false,
		description: 'Whether to return every copied row or only up to a given limit',
	},
	{
		displayName: 'Limit',
		name: 'limit',
		type: 'number',
		displayOptions: {
			show: {
				resource: ['assetSnapshot'],
				operation: ['get'],
				returnAll: [false],
			},
		},
		typeOptions: {
			minValue: 1,
		},
		default: 50,
		description: 'Max number of copied rows to return',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const snapshotId = this.getNodeParameter('snapshotId', itemIndex) as string;
	const returnAll = this.getNodeParameter('returnAll', itemIndex) as boolean;
	const limit = this.getNodeParameter('limit', itemIndex, 50) as number;

	const headerQuery = `
		query GetAssetSnapshot($snapshotId: ID!) {
			node(id: $snapshotId) {
				__typename
				... on AssetSnapshot {
					id
					capturedAt
					name
					note
					actor {
						id
						fullName
					}
				}
			}
		}
	`;

	const headerData = await proboApiRequest.call(this, headerQuery, { snapshotId });
	const header = headerData?.data as IDataObject | undefined;
	const node = header?.node as IDataObject | undefined;

	if (!node) {
		throw new NodeOperationError(this.getNode(), `Asset snapshot ${snapshotId} not found`);
	}

	if (node.__typename !== 'AssetSnapshot') {
		throw new NodeOperationError(
			this.getNode(),
			`Expected AssetSnapshot node for ${snapshotId}, got ${String(node.__typename)}`,
		);
	}

	const assetsQuery = `
		query GetAssetSnapshotAssets($snapshotId: ID!, $first: Int, $after: CursorKey) {
			node(id: $snapshotId) {
				__typename
				... on AssetSnapshot {
					assets(first: $first, after: $after, orderBy: { field: NAME, direction: ASC }) {
						edges {
							node {
								id
								assetId
								name
								amount
								assetType
								dataTypesStored
								kind
								externalId
								region
								configuration
								schemaVersion
								lastSeenAt
								status
								connectorId
								thirdPartyIds
								businessFunctionIds
								tagIds
								owner {
									id
									fullName
								}
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

	const assets = await proboApiRequestAllItems.call(
		this,
		assetsQuery,
		{ snapshotId },
		(response) => {
			const data = response?.data as IDataObject | undefined;
			const snapshot = data?.node as IDataObject | undefined;
			return snapshot?.assets as IDataObject | undefined;
		},
		returnAll,
		limit,
	);

	return {
		json: {
			...node,
			assets,
		},
		pairedItem: { item: itemIndex },
	};
}
