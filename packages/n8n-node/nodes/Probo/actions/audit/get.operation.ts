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
				operation: ['get'],
			},
		},
		default: '',
		description: 'The ID of the report',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const reportId = this.getNodeParameter('reportId', itemIndex) as string;

	const query = `
		query GetReport($reportId: ID!) {
			node(id: $reportId) {
				... on Report {
					id
					name
					frameworkType
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
		reportId,
	};

	const responseData = await proboApiRequest.call(this, query, variables);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
