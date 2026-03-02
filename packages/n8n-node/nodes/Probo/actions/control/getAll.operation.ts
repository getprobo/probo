import type { INodeProperties, IExecuteFunctions, INodeExecutionData, IDataObject } from 'n8n-workflow';
import { proboApiRequestAllItems } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Framework ID',
		name: 'frameworkId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['control'],
				operation: ['getAll'],
			},
		},
		default: '',
		description: 'The ID of the framework',
		required: true,
	},
	{
		displayName: 'Return All',
		name: 'returnAll',
		type: 'boolean',
		displayOptions: {
			show: {
				resource: ['control'],
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
				resource: ['control'],
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
	const frameworkId = this.getNodeParameter('frameworkId', itemIndex) as string;
	const returnAll = this.getNodeParameter('returnAll', itemIndex) as boolean;
	const limit = this.getNodeParameter('limit', itemIndex, 50) as number;

	const query = `
		query GetControls($frameworkId: ID!, $first: Int, $after: CursorKey) {
			node(id: $frameworkId) {
				... on Framework {
					controls(first: $first, after: $after) {
						edges {
							node {
								id
								sectionTitle
								name
								description
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

	const controls = await proboApiRequestAllItems.call(
		this,
		query,
		{ frameworkId },
		(response) => {
			const data = response?.data as IDataObject | undefined;
			const node = data?.node as IDataObject | undefined;
			return node?.controls as IDataObject | undefined;
		},
		returnAll,
		limit,
	);

	return {
		json: { controls },
		pairedItem: { item: itemIndex },
	};
}

