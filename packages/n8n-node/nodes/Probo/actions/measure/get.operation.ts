import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Measure ID',
		name: 'measureId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['measure'],
				operation: ['get'],
			},
		},
		default: '',
		description: 'The ID of the measure',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const measureId = this.getNodeParameter('measureId', itemIndex) as string;

	const query = `
		query GetMeasure($measureId: ID!) {
			node(id: $measureId) {
				... on Measure {
					id
					name
					description
					category
					state
					createdAt
					updatedAt
				}
			}
		}
	`;

	const variables = {
		measureId,
	};

	const responseData = await proboApiRequest.call(this, query, variables);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}

