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
				operation: ['delete'],
			},
		},
		default: '',
		description: 'The ID of the asset to delete',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const assetId = this.getNodeParameter('assetId', itemIndex) as string;

	const query = `
		mutation DeleteAsset($input: DeleteAssetInput!) {
			deleteAsset(input: $input) {
				deletedAssetId
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, { input: { assetId } });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}

