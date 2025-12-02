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
	{
		displayName: 'Status',
		name: 'status',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['control'],
				operation: ['update'],
			},
		},
		options: [
			{
				name: 'Included',
				value: 'INCLUDED',
			},
			{
				name: 'Excluded',
				value: 'EXCLUDED',
			},
		],
		default: '',
		description: 'The status of the control',
	},
	{
		displayName: 'Exclusion Justification',
		name: 'exclusionJustification',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['control'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The justification for excluding the control',
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
	const status = this.getNodeParameter('status', itemIndex, '') as string;
	const exclusionJustification = this.getNodeParameter('exclusionJustification', itemIndex, '') as string;

	const query = `
		mutation UpdateControl($input: UpdateControlInput!) {
			updateControl(input: $input) {
				control {
					id
					sectionTitle
					name
					description
					status
					exclusionJustification
					createdAt
					updatedAt
				}
			}
		}
	`;

	const input: Record<string, any> = { id };
	if (sectionTitle) input.sectionTitle = sectionTitle;
	if (name) input.name = name;
	if (description) input.description = description;
	if (status) input.status = status;
	if (exclusionJustification) input.exclusionJustification = exclusionJustification;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}

