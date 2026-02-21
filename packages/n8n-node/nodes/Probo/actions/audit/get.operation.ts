import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Audit ID',
		name: 'auditId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['audit'],
				operation: ['get'],
			},
		},
		default: '',
		description: 'The ID of the audit',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const auditId = this.getNodeParameter('auditId', itemIndex) as string;

	const query = `
		query GetAudit($auditId: ID!) {
			node(id: $auditId) {
				... on Audit {
					id
					name
					state
					validFrom
					validUntil
					reportUrl
					trustCenterVisibility
					createdAt
					updatedAt
				}
			}
		}
	`;

	const variables = {
		auditId,
	};

	const responseData = await proboApiRequest.call(this, query, variables);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
