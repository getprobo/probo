import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Control ID',
		name: 'id',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['control'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the control to update',
		required: true,
	},
	{
		displayName: 'Section Title',
		name: 'sectionTitle',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['control'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The section title of the control',
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['control'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The name of the control',
	},
	{
		displayName: 'Description',
		name: 'description',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['control'],
				operation: ['update'],
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
	const id = this.getNodeParameter('id', itemIndex) as string;
	const sectionTitle = this.getNodeParameter('sectionTitle', itemIndex, '') as string;
	const name = this.getNodeParameter('name', itemIndex, '') as string;
	const description = this.getNodeParameter('description', itemIndex, '') as string;

	const query = `
		mutation UpdateControl($input: UpdateControlInput!) {
			updateControl(input: $input) {
				control {
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

	const input: Record<string, string> = { id };
	if (sectionTitle) input.sectionTitle = sectionTitle;
	if (name) input.name = name;
	if (description) input.description = description;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
