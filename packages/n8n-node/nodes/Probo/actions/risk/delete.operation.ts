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
				operation: ['delete'],
			},
		},
		default: '',
		description: 'The ID of the risk to delete',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const riskId = this.getNodeParameter('riskId', itemIndex) as string;

	const query = `
		mutation DeleteRisk($input: DeleteRiskInput!) {
			deleteRisk(input: $input) {
				deletedRiskId
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, { input: { riskId } });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
