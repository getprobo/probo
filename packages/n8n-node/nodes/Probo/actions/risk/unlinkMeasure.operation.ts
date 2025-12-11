import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Risk ID',
		name: 'riskId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['risk'],
				operation: ['unlinkMeasure'],
			},
		},
		default: '',
		description: 'The ID of the risk',
		required: true,
	},
	{
		displayName: 'Measure ID',
		name: 'measureId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['risk'],
				operation: ['unlinkMeasure'],
			},
		},
		default: '',
		description: 'The ID of the measure to unlink',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const riskId = this.getNodeParameter('riskId', itemIndex) as string;
	const measureId = this.getNodeParameter('measureId', itemIndex) as string;

	const query = `
		mutation DeleteRiskMeasureMapping($input: DeleteRiskMeasureMappingInput!) {
			deleteRiskMeasureMapping(input: $input) {
				deletedRiskId
				deletedMeasureId
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, { input: { riskId, measureId } });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
