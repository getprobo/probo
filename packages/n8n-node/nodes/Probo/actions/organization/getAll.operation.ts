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

import type { INodeProperties, IExecuteFunctions, INodeExecutionData, IDataObject } from 'n8n-workflow';
import { proboConnectApiRequestAllItems } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Return All',
		name: 'returnAll',
		type: 'boolean',
		displayOptions: {
			show: {
				resource: ['organization'],
				operation: ['getAll'],
			},
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
				resource: ['organization'],
				operation: ['getAll'],
				returnAll: [false],
			},
		},
		typeOptions: {
			minValue: 1,
		},
		default: 50,
		description: 'Max number of results to return',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const returnAll = this.getNodeParameter('returnAll', itemIndex) as boolean;
	const limit = this.getNodeParameter('limit', itemIndex, 50) as number;

	const query = `
		query GetOrganizations($first: Int, $after: CursorKey, $filter: ProfileFilter) {
			viewer {
				profiles(first: $first, after: $after, filter: $filter) {
					edges {
						node {
							organization {
								id
								name
								logo {
									id
									fileName
									downloadUrl
								}
								horizontalLogo {
									id
									fileName
									downloadUrl
								}
								createdAt
								updatedAt
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
	`;

	const memberships = await proboConnectApiRequestAllItems.call(
		this,
		query,
		{ filter: { state: 'ACTIVE' } },
		(response: IDataObject) => {
			const data = response?.data as IDataObject | undefined;
			const viewer = data?.viewer as IDataObject | undefined;
			return viewer?.profiles as IDataObject | undefined;
		},
		returnAll,
		limit,
	);

	const organizationMap = new Map<string, IDataObject>();
	for (const membership of memberships) {
		const org = membership.organization as IDataObject | undefined;
		if (org && org.id) {
			organizationMap.set(org.id as string, org);
		}
	}
	const organizations = Array.from(organizationMap.values());

	return {
		json: { organizations },
		pairedItem: { item: itemIndex },
	};
}
