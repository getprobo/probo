import type { INodeProperties, IExecuteFunctions, INodeExecutionData, IDataObject } from 'n8n-workflow';
import { proboApiRequestAllItems } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Organization ID',
		name: 'organizationId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['datum'],
				operation: ['getAll'],
			},
		},
		default: '',
		description: 'The ID of the organization',
		required: true,
	},
	{
		displayName: 'Return All',
		name: 'returnAll',
		type: 'boolean',
		displayOptions: {
			show: {
				resource: ['datum'],
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
				resource: ['datum'],
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
	{
		displayName: 'Options',
		name: 'options',
		type: 'collection',
		placeholder: 'Add Option',
		default: {},
		displayOptions: {
			show: {
				resource: ['datum'],
				operation: ['getAll'],
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
	const returnAll = this.getNodeParameter('returnAll', itemIndex) as boolean;
	const limit = this.getNodeParameter('limit', itemIndex, 50) as number;
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
		query GetData($organizationId: ID!, $first: Int, $after: CursorKey) {
			node(id: $organizationId) {
				... on Organization {
					data(first: $first, after: $after) {
						edges {
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
						pageInfo {
							hasNextPage
							endCursor
						}
					}
				}
			}
		}
	`;

	const data = await proboApiRequestAllItems.call(
		this,
		query,
		{ organizationId },
		(response) => {
			const responseData = response?.data as IDataObject | undefined;
			const node = responseData?.node as IDataObject | undefined;
			return node?.data as IDataObject | undefined;
		},
		returnAll,
		limit,
	);

	return {
		json: { data },
		pairedItem: { item: itemIndex },
	};
}
