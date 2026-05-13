// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Asset ID',
		name: 'id',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['asset'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the asset to update',
		required: true,
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['asset'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The name of the asset',
	},
	{
		displayName: 'Amount',
		name: 'amount',
		type: 'number',
		displayOptions: {
			show: {
				resource: ['asset'],
				operation: ['update'],
			},
		},
		default: 0,
		description: 'The amount of the asset',
	},
	{
		displayName: 'Owner ID',
		name: 'ownerId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['asset'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the owner (People)',
	},
	{
		displayName: 'Asset Type',
		name: 'assetType',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['asset'],
				operation: ['update'],
			},
		},
		options: [
			{
				name: 'Physical',
				value: 'PHYSICAL',
			},
			{
				name: 'Virtual',
				value: 'VIRTUAL',
			},
		],
		default: 'PHYSICAL',
		description: 'The type of the asset',
	},
	{
		displayName: 'Data Types Stored',
		name: 'dataTypesStored',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['asset'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The types of data stored in the asset',
	},
	{
		displayName: 'ThirdParty IDs',
		name: 'thirdPartyIds',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['asset'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'Comma-separated list of thirdParty IDs',
	},
	{
		displayName: 'Options',
		name: 'options',
		type: 'collection',
		placeholder: 'Add Option',
		default: {},
		displayOptions: {
			show: {
				resource: ['asset'],
				operation: ['update'],
			},
		},
		options: [
			{
				displayName: 'Include Owner',
				name: 'includeOwner',
				type: 'boolean',
				default: false,
				description: 'Whether to include owner details in the response',
			},
			{
				displayName: 'Include ThirdParties',
				name: 'includeThirdParties',
				type: 'boolean',
				default: false,
				description: 'Whether to include thirdParties in the response',
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const id = this.getNodeParameter('id', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex, '') as string;
	const amount = this.getNodeParameter('amount', itemIndex, 0) as number;
	const ownerId = this.getNodeParameter('ownerId', itemIndex, '') as string;
	const assetType = this.getNodeParameter('assetType', itemIndex, '') as string;
	const dataTypesStored = this.getNodeParameter('dataTypesStored', itemIndex, '') as string;
	const thirdPartyIdsStr = this.getNodeParameter('thirdPartyIds', itemIndex, '') as string;
	const options = this.getNodeParameter('options', itemIndex, {}) as {
		includeOwner?: boolean;
		includeThirdParties?: boolean;
	};

	const ownerFragment = options.includeOwner
		? `owner {
			id
			fullName
			emailAddress
		}`
		: '';

	const thirdPartiesFragment = options.includeThirdParties
		? `thirdParties(first: 100) {
			edges {
				node {
					id
					name
				}
			}
		}`
		: '';

	const query = `
		mutation UpdateAsset($input: UpdateAssetInput!) {
			updateAsset(input: $input) {
				asset {
					id
					name
					amount
					assetType
					dataTypesStored
					${ownerFragment}
					${thirdPartiesFragment}
					createdAt
					updatedAt
				}
			}
		}
	`;

	const input: Record<string, string | number | string[]> = { id };
	if (name) input.name = name;
	if (amount) input.amount = amount;
	if (ownerId) input.ownerId = ownerId;
	if (assetType) input.assetType = assetType;
	if (dataTypesStored) input.dataTypesStored = dataTypesStored;
	if (thirdPartyIdsStr) {
		const thirdPartyIds = thirdPartyIdsStr.split(',').map((vid) => vid.trim()).filter(Boolean);
		if (thirdPartyIds.length > 0) input.thirdPartyIds = thirdPartyIds;
	}

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
