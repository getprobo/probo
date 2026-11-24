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
import { organizationLocator } from '../../locators';
import { proboApiRequestAllItems, executionItems } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	organizationLocator({
		resource: ['accessInventory'],
		operation: ['listChangeEvents'],
	}),
	{
		displayName: 'Access Review Source ID',
		name: 'accessReviewSourceId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['accessInventory'],
				operation: ['listChangeEvents'],
			},
		},
		default: '',
		description: 'Filter change events by access review source ID',
	},
	{
		displayName: 'Change Types',
		name: 'changeTypes',
		type: 'multiOptions',
		displayOptions: {
			show: {
				resource: ['accessInventory'],
				operation: ['listChangeEvents'],
			},
		},
		options: [
			{ name: 'Appeared', value: 'APPEARED' },
			{ name: 'Disappeared', value: 'DISAPPEARED' },
			{ name: 'Attribute Changed', value: 'ATTRIBUTE_CHANGED' },
		],
		default: [],
		description: 'Filter by change event type',
	},
	{
		displayName: 'Observed At On Or After',
		name: 'observedAtOnOrAfter',
		type: 'dateTime',
		displayOptions: {
			show: {
				resource: ['accessInventory'],
				operation: ['listChangeEvents'],
			},
		},
		default: '',
		description: 'Include events observed at or after this time',
	},
	{
		displayName: 'Observed At On Or Before',
		name: 'observedAtOnOrBefore',
		type: 'dateTime',
		displayOptions: {
			show: {
				resource: ['accessInventory'],
				operation: ['listChangeEvents'],
			},
		},
		default: '',
		description: 'Include events observed at or before this time',
	},
	{
		displayName: 'Return All',
		name: 'returnAll',
		type: 'boolean',
		displayOptions: {
			show: {
				resource: ['accessInventory'],
				operation: ['listChangeEvents'],
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
				resource: ['accessInventory'],
				operation: ['listChangeEvents'],
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
): Promise<INodeExecutionData[]> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex, '', { extractValue: true }) as string;
	const accessReviewSourceId = this.getNodeParameter('accessReviewSourceId', itemIndex, '') as string;
	const changeTypes = this.getNodeParameter('changeTypes', itemIndex, []) as string[];
	const observedAtOnOrAfter = this.getNodeParameter('observedAtOnOrAfter', itemIndex, '') as string;
	const observedAtOnOrBefore = this.getNodeParameter('observedAtOnOrBefore', itemIndex, '') as string;
	const returnAll = this.getNodeParameter('returnAll', itemIndex) as boolean;
	const limit = this.getNodeParameter('limit', itemIndex, 50) as number;

	const filter: IDataObject = {};
	if (accessReviewSourceId) {
		filter.accessReviewSourceId = accessReviewSourceId;
	}
	if (changeTypes.length > 0) {
		filter.changeTypes = changeTypes;
	}
	if (observedAtOnOrAfter) {
		filter.observedAtOnOrAfter = observedAtOnOrAfter;
	}
	if (observedAtOnOrBefore) {
		filter.observedAtOnOrBefore = observedAtOnOrBefore;
	}

	const query = `
		query(
			$organizationId: ID!,
			$first: Int,
			$after: CursorKey,
			$filter: AccessInventoryChangeEventFilter
		) {
			node(id: $organizationId) {
				... on Organization {
					accessInventoryChangeEvents(first: $first, after: $after, filter: $filter) {
						edges {
							node {
								id
								changeType
								principalKey
								resourceKey
								entitlementKey
								observedAt
								providerOccurredAt
								effectiveAt
								createdAt
								before {
									entitlementDisplayName
									presence
								}
								after {
									entitlementDisplayName
									presence
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

	const events = await proboApiRequestAllItems.call(
		this,
		query,
		{
			organizationId,
			filter: Object.keys(filter).length > 0 ? filter : undefined,
		},
		(response) => {
			const data = response?.data as IDataObject | undefined;
			const node = data?.node as IDataObject | undefined;
			return node?.accessInventoryChangeEvents as IDataObject | undefined;
		},
		returnAll,
		limit,
	);

	return executionItems(events, itemIndex);
}
