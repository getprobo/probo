// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import type { INodeProperties, IExecuteFunctions, INodeExecutionData, IDataObject } from 'n8n-workflow';
import { proboApiRequestAllItems } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'ThirdParty ID',
		name: 'thirdPartyId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['thirdParty'],
				operation: ['getAllContacts'],
			},
		},
		default: '',
		description: 'The ID of the thirdParty',
		required: true,
	},
	{
		displayName: 'Return All',
		name: 'returnAll',
		type: 'boolean',
		displayOptions: {
			show: {
				resource: ['thirdParty'],
				operation: ['getAllContacts'],
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
				resource: ['thirdParty'],
				operation: ['getAllContacts'],
				returnAll: [false],
			},
		},
		typeOptions: {
			minValue: 1,
		},
		default: 50,
		description: 'Max number of results to return',
	},
	{
		displayName: 'Options',
		name: 'options',
		type: 'collection',
		placeholder: 'Add Option',
		default: {},
		displayOptions: {
			show: {
				resource: ['thirdParty'],
				operation: ['getAllContacts'],
			},
		},
		options: [
			{
				displayName: 'Include ThirdParty',
				name: 'includeThirdParty',
				type: 'boolean',
				default: false,
				description: 'Whether to include thirdParty in the response',
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const thirdPartyId = this.getNodeParameter('thirdPartyId', itemIndex) as string;
	const returnAll = this.getNodeParameter('returnAll', itemIndex) as boolean;
	const limit = this.getNodeParameter('limit', itemIndex, 50) as number;
	const options = this.getNodeParameter('options', itemIndex, {}) as {
		includeThirdParty?: boolean;
	};

	const thirdPartyFragment = options.includeThirdParty
		? `thirdParty {
			id
			name
		}`
		: '';

	const query = `
		query GetThirdPartyContacts($thirdPartyId: ID!, $first: Int, $after: CursorKey) {
			node(id: $thirdPartyId) {
				... on ThirdParty {
					contacts(first: $first, after: $after) {
						edges {
							node {
								id
								fullName
								email
								phone
								role
								${thirdPartyFragment}
								createdAt
								updatedAt
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

	const thirdPartyContacts = await proboApiRequestAllItems.call(
		this,
		query,
		{ thirdPartyId },
		(response) => {
			const data = response?.data as IDataObject | undefined;
			const node = data?.node as IDataObject | undefined;
			return node?.contacts as IDataObject | undefined;
		},
		returnAll,
		limit,
	);

	return {
		json: { thirdPartyContacts },
		pairedItem: { item: itemIndex },
	};
}
