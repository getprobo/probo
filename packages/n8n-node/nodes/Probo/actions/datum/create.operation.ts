import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Organization ID',
		name: 'organizationId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['datum'],
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
				resource: ['datum'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The name of the datum',
		required: true,
	},
	{
		displayName: 'Data Classification',
		name: 'dataClassification',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['datum'],
				operation: ['create'],
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
		default: 'INTERNAL',
		description: 'The classification of the data',
		required: true,
	},
	{
		displayName: 'Owner ID',
		name: 'ownerId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['datum'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The ID of the owner (People)',
		required: true,
	},
	{
		displayName: 'Vendor IDs',
		name: 'vendorIds',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['datum'],
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
				resource: ['datum'],
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
	const dataClassification = this.getNodeParameter('dataClassification', itemIndex) as string;
	const ownerId = this.getNodeParameter('ownerId', itemIndex) as string;
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
		mutation CreateDatum($input: CreateDatumInput!) {
			createDatum(input: $input) {
				datumEdge {
					node {
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
		}
	`;

	const vendorIds = vendorIdsStr ? vendorIdsStr.split(',').map((id) => id.trim()).filter(Boolean) : undefined;

	const variables = {
		input: {
			organizationId,
			name,
			dataClassification,
			ownerId,
			...(vendorIds && vendorIds.length > 0 && { vendorIds }),
		},
	};

	const responseData = await proboApiRequest.call(this, query, variables);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
