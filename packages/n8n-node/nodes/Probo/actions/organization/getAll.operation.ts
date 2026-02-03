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
		query GetOrganizations($first: Int, $after: CursorKey) {
			viewer {
				memberships(first: $first, after: $after) {
					edges {
						node {
							organization {
								id
								name
								description
								websiteUrl
								email
								headquarterAddress
								logoUrl
								horizontalLogoUrl
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
		{},
		(response: IDataObject) => {
			const data = response?.data as IDataObject | undefined;
			const viewer = data?.viewer as IDataObject | undefined;
			return viewer?.memberships as IDataObject | undefined;
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

