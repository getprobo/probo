import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Report ID',
		name: 'reportId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['report'],
				operation: ['deleteReport'],
			},
		},
		default: '',
		description: 'The ID of the audit report file to delete',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const reportId = this.getNodeParameter('reportId', itemIndex) as string;

	const query = `
		mutation DeleteAuditReport($input: DeleteAuditReportInput!) {
			deleteAuditReport(input: $input) {
				report {
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

	const responseData = await proboApiRequest.call(this, query, { input: { reportId } });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
