import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Datum ID',
		name: 'id',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['datum'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the datum to update',
		required: true,
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['datum'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The name of the datum',
	},
	{
		displayName: 'Data Classification',
		name: 'dataClassification',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['datum'],
				operation: ['update'],
			},
		},
		options: [
			{
				name: 'Public',
				value: 'PUBLIC',
			},
			{
				name: 'Internal',
				value: 'INTERNAL',
			},
			{
				name: 'Confidential',
				value: 'CONFIDENTIAL',
			},
			{
				name: 'Secret',
				value: 'SECRET',
			},
		],
		default: 'PUBLIC',
		description: 'The classification of the data',
	},
	{
		displayName: 'Owner ID',
		name: 'ownerId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['datum'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the owner (People)',
	},
	{
		displayName: 'Vendor IDs',
		name: 'vendorIds',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['datum'],
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
				resource: ['datum'],
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
	const dataClassification = this.getNodeParameter('dataClassification', itemIndex, '') as string;
	const ownerId = this.getNodeParameter('ownerId', itemIndex, '') as string;
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
		mutation UpdateDatum($input: UpdateDatumInput!) {
			updateDatum(input: $input) {
				datum {
					id
					name
					dataClassification
					${ownerFragment}
					${vendorsFragment}
					createdAt
					updatedAt
				}
			}
		}
	`;

	const input: Record<string, string | string[]> = { id };
	if (name) input.name = name;
	if (dataClassification) input.dataClassification = dataClassification;
	if (ownerId) input.ownerId = ownerId;
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
