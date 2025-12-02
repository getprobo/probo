import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Framework ID',
		name: 'frameworkId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['framework'],
				operation: ['get'],
			},
		},
		default: '',
		description: 'The ID of the framework',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const frameworkId = this.getNodeParameter('frameworkId', itemIndex) as string;

	const query = `
		query GetFramework($frameworkId: ID!) {
			node(id: $frameworkId) {
				... on Framework {
					id
					name
					description
					createdAt
					updatedAt
				}
			}
		}
	`;

	const variables = {
		frameworkId,
	};

	const responseData = await proboApiRequest.call(this, query, variables);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}

