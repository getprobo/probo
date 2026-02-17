import type {
	INodeProperties,
	IExecuteFunctions,
	INodeExecutionData,
	IDataObject,
} from 'n8n-workflow';
import { proboConnectApiRequestAllItems } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Organization ID',
		name: 'organizationId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['user'],
				operation: ['listUsers'],
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
				resource: ['user'],
				operation: ['listUsers'],
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
				resource: ['user'],
				operation: ['listUsers'],
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
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const returnAll = this.getNodeParameter('returnAll', itemIndex) as boolean;
	const limit = this.getNodeParameter('limit', itemIndex, 50) as number;

	const query = `
		query ListUsers($organizationId: ID!, $first: Int, $after: CursorKey, $orderBy: ProfileOrder) {
			node(id: $organizationId) {
				... on Organization {
					profiles(first: $first, after: $after, orderBy: $orderBy) {
						edges {
							node {
								id
								fullName
								emailAddress
								source
								state
								additionalEmailAddresses
								kind
								position
								contractStartDate
								contractEndDate
								createdAt
								updatedAt
								organization { id name }
								membership { id role createdAt }
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

	const users = await proboConnectApiRequestAllItems.call(
		this,
		query,
		{ organizationId },
		(response: IDataObject) => {
			const data = response?.data as IDataObject | undefined;
			const node = data?.node as IDataObject | undefined;
			const org = node as IDataObject | undefined;
			return org?.profiles as IDataObject | undefined;
		},
		returnAll,
		limit,
	);

	return {
		json: { users },
		pairedItem: { item: itemIndex },
	};
}
