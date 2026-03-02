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
				operation: ['delete'],
			},
		},
		default: '',
		description: 'The ID of the audit to delete',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const auditId = this.getNodeParameter('auditId', itemIndex) as string;

	const query = `
		mutation DeleteAudit($input: DeleteAuditInput!) {
			deleteAudit(input: $input) {
				deletedAuditId
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, { input: { auditId } });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
