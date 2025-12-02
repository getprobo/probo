import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Asset ID',
		name: 'assetId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['asset'],
				operation: ['get'],
			},
		},
		default: '',
		description: 'The ID of the asset',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const assetId = this.getNodeParameter('assetId', itemIndex) as string;

	const query = `
		query GetAsset($assetId: ID!) {
			node(id: $assetId) {
				... on Asset {
					id
					name
					amount
					assetType
					dataTypesStored
					createdAt
					updatedAt
				}
			}
		}
	`;

	const variables = {
		assetId,
	};

	const responseData = await proboApiRequest.call(this, query, variables);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}

