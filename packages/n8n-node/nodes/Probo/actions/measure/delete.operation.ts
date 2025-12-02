import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Measure ID',
		name: 'measureId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['measure'],
				operation: ['delete'],
			},
		},
		default: '',
		description: 'The ID of the measure to delete',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const measureId = this.getNodeParameter('measureId', itemIndex) as string;

	const query = `
		mutation DeleteMeasure($input: DeleteMeasureInput!) {
			deleteMeasure(input: $input) {
				deletedMeasureId
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, { input: { measureId } });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
