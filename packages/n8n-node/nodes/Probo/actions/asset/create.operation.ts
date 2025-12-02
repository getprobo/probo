import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Organization ID',
		name: 'organizationId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['asset'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The ID of the organization',
		required: true,
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['asset'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The name of the asset',
		required: true,
	},
	{
		displayName: 'Amount',
		name: 'amount',
		type: 'number',
		displayOptions: {
			show: {
				resource: ['asset'],
				operation: ['create'],
			},
		},
		default: 1,
		description: 'The amount of the asset',
		required: true,
	},
	{
		displayName: 'Owner ID',
		name: 'ownerId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['asset'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The ID of the owner (People)',
		required: true,
	},
	{
		displayName: 'Asset Type',
		name: 'assetType',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['asset'],
				operation: ['create'],
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
		required: true,
	},
	{
		displayName: 'Data Types Stored',
		name: 'dataTypesStored',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['asset'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The types of data stored in the asset',
		required: true,
	},
	{
		displayName: 'Vendor IDs',
		name: 'vendorIds',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['asset'],
				operation: ['create'],
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
				operation: ['create'],
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
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex) as string;
	const amount = this.getNodeParameter('amount', itemIndex) as number;
	const ownerId = this.getNodeParameter('ownerId', itemIndex) as string;
	const assetType = this.getNodeParameter('assetType', itemIndex) as string;
	const dataTypesStored = this.getNodeParameter('dataTypesStored', itemIndex) as string;
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
		mutation CreateAsset($input: CreateAssetInput!) {
			createAsset(input: $input) {
				assetEdge {
					node {
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
		}
	`;

	const vendorIds = vendorIdsStr ? vendorIdsStr.split(',').map((id) => id.trim()).filter(Boolean) : undefined;

	const variables = {
		input: {
			organizationId,
			name,
			amount,
			ownerId,
			assetType,
			dataTypesStored,
			...(vendorIds && vendorIds.length > 0 && { vendorIds }),
		},
	};

	const responseData = await proboApiRequest.call(this, query, variables);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
