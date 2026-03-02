import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Framework ID',
		name: 'frameworkId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['control'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The ID of the framework',
		required: true,
	},
	{
		displayName: 'Section Title',
		name: 'sectionTitle',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['control'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The section title of the control',
		required: true,
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['control'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The name of the control',
		required: true,
	},
	{
		displayName: 'Description',
		name: 'description',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['control'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The description of the control',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const frameworkId = this.getNodeParameter('frameworkId', itemIndex) as string;
	const sectionTitle = this.getNodeParameter('sectionTitle', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex) as string;
	const description = this.getNodeParameter('description', itemIndex, '') as string;

	const query = `
		mutation CreateControl($input: CreateControlInput!) {
			createControl(input: $input) {
				controlEdge {
					node {
						id
						sectionTitle
						name
						description
						createdAt
						updatedAt
					}
				}
			}
		}
	`;

	const variables = {
		input: {
			frameworkId,
			sectionTitle,
			name,
			bestPractice: true,
			...(description && { description }),
		},
	};

	const responseData = await proboApiRequest.call(this, query, variables);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
