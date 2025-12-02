import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Datum ID',
		name: 'datumId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['datum'],
				operation: ['delete'],
			},
		},
		default: '',
		description: 'The ID of the datum to delete',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const datumId = this.getNodeParameter('datumId', itemIndex) as string;

	const query = `
		mutation DeleteDatum($input: DeleteDatumInput!) {
			deleteDatum(input: $input) {
				deletedDatumId
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, { input: { datumId } });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}

