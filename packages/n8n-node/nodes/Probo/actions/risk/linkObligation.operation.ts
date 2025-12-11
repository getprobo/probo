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
				operation: ['linkObligation'],
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
				operation: ['linkObligation'],
			},
		},
		default: '',
		description: 'The ID of the obligation to link',
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
		mutation CreateRiskObligationMapping($input: CreateRiskObligationMappingInput!) {
			createRiskObligationMapping(input: $input) {
				riskEdge {
					node {
						id
						name
					}
				}
				obligationEdge {
					node {
						id
						name
					}
				}
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, { input: { riskId, obligationId } });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
