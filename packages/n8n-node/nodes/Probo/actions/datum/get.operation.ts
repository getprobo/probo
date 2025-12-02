import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Datum ID',
		name: 'datumId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['datum'],
				operation: ['get'],
			},
		},
		default: '',
		description: 'The ID of the datum',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const datumId = this.getNodeParameter('datumId', itemIndex) as string;

	const query = `
		query GetDatum($datumId: ID!) {
			node(id: $datumId) {
				... on Datum {
					id
					name
					dataClassification
					createdAt
					updatedAt
				}
			}
		}
	`;

	const variables = {
		datumId,
	};

	const responseData = await proboApiRequest.call(this, query, variables);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}

