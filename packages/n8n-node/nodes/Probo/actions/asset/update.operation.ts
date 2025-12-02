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
		displayName: 'Vendor IDs',
		name: 'vendorIds',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['asset'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'Comma-separated list of vendor IDs',
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
				displayName: 'Include Vendors',
				name: 'includeVendors',
				type: 'boolean',
				default: false,
				description: 'Whether to include vendors in the response',
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
	const vendorIdsStr = this.getNodeParameter('vendorIds', itemIndex, '') as string;
	const options = this.getNodeParameter('options', itemIndex, {}) as {
		includeOwner?: boolean;
		includeVendors?: boolean;
	};

	const ownerFragment = options.includeOwner
		? `owner {
			id
			fullName
			primaryEmailAddress
		}`
		: '';

	const vendorsFragment = options.includeVendors
		? `vendors(first: 100) {
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
					${vendorsFragment}
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
	if (vendorIdsStr) {
		const vendorIds = vendorIdsStr.split(',').map((vid) => vid.trim()).filter(Boolean);
		if (vendorIds.length > 0) input.vendorIds = vendorIds;
	}

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
