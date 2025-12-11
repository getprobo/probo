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
				operation: ['get'],
			},
		},
		default: '',
		description: 'The ID of the risk',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const riskId = this.getNodeParameter('riskId', itemIndex) as string;

	const query = `
		query GetRisk($riskId: ID!) {
			node(id: $riskId) {
				... on Risk {
					id
					name
					description
					category
					treatment
					inherentLikelihood
					inherentImpact
					inherentRiskScore
					residualLikelihood
					residualImpact
					residualRiskScore
					note
					createdAt
					updatedAt
				}
			}
		}
	`;

	const variables = {
		riskId,
	};

	const responseData = await proboApiRequest.call(this, query, variables);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
