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
				operation: ['unlinkObligation'],
			},
		},
		default: '',
		description: 'The ID of the risk',
		required: true,
	},
	{
		displayName: 'Obligation ID',
		name: 'obligationId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['risk'],
				operation: ['unlinkObligation'],
			},
		},
		default: '',
		description: 'The ID of the obligation to unlink',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const riskId = this.getNodeParameter('riskId', itemIndex) as string;
	const obligationId = this.getNodeParameter('obligationId', itemIndex) as string;

	const query = `
		mutation DeleteRiskObligationMapping($input: DeleteRiskObligationMappingInput!) {
			deleteRiskObligationMapping(input: $input) {
				deletedRiskId
				deletedObligationId
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, { input: { riskId, obligationId } });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
