import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Control ID',
		name: 'controlId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['control'],
				operation: ['get'],
			},
		},
		default: '',
		description: 'The ID of the control',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const controlId = this.getNodeParameter('controlId', itemIndex) as string;

	const query = `
		query GetControl($controlId: ID!) {
			node(id: $controlId) {
				... on Control {
					id
					sectionTitle
					name
					description
					createdAt
					updatedAt
				}
			}
		}
	`;

	const variables = {
		controlId,
	};

	const responseData = await proboApiRequest.call(this, query, variables);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}

