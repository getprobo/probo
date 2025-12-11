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
				operation: ['linkMeasure'],
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
				operation: ['linkMeasure'],
			},
		},
		default: '',
		description: 'The ID of the measure to link',
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
		mutation CreateRiskMeasureMapping($input: CreateRiskMeasureMappingInput!) {
			createRiskMeasureMapping(input: $input) {
				riskEdge {
					node {
						id
						name
					}
				}
				measureEdge {
					node {
						id
						name
					}
				}
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, { input: { riskId, measureId } });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
