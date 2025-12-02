import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Framework ID',
		name: 'id',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['framework'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the framework to update',
		required: true,
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['framework'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The name of the framework',
	},
	{
		displayName: 'Description',
		name: 'description',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['framework'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The description of the framework',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const id = this.getNodeParameter('id', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex, '') as string;
	const description = this.getNodeParameter('description', itemIndex, '') as string;

	const query = `
		mutation UpdateFramework($input: UpdateFrameworkInput!) {
			updateFramework(input: $input) {
				framework {
					id
					name
					description
					createdAt
					updatedAt
				}
			}
		}
	`;

	const input: Record<string, string> = { id };
	if (name) input.name = name;
	if (description) input.description = description;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}

