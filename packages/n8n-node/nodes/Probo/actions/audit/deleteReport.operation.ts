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
				operation: ['deleteReport'],
			},
		},
		default: '',
		description: 'The ID of the audit whose report to delete',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const auditId = this.getNodeParameter('auditId', itemIndex) as string;

	const query = `
		mutation DeleteAuditReport($input: DeleteAuditReportInput!) {
			deleteAuditReport(input: $input) {
				audit {
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

	const responseData = await proboApiRequest.call(this, query, { input: { auditId } });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
