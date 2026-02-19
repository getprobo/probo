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
				operation: ['delete'],
			},
		},
		default: '',
		description: 'The ID of the report to delete',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const reportId = this.getNodeParameter('reportId', itemIndex) as string;

	const query = `
		mutation DeleteReport($input: DeleteReportInput!) {
			deleteReport(input: $input) {
				deletedReportId
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, { input: { reportId } });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
