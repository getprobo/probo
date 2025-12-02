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
		displayName: 'Status',
		name: 'status',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['control'],
				operation: ['create'],
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
		default: 'INCLUDED',
		description: 'The status of the control',
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
	{
		displayName: 'Exclusion Justification',
		name: 'exclusionJustification',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['control'],
				operation: ['create'],
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
	const frameworkId = this.getNodeParameter('frameworkId', itemIndex) as string;
	const sectionTitle = this.getNodeParameter('sectionTitle', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex) as string;
	const status = this.getNodeParameter('status', itemIndex) as string;
	const description = this.getNodeParameter('description', itemIndex, '') as string;
	const exclusionJustification = this.getNodeParameter('exclusionJustification', itemIndex, '') as string;

	const query = `
		mutation CreateControl($input: CreateControlInput!) {
			createControl(input: $input) {
				controlEdge {
					node {
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
		}
	`;

	const variables = {
		input: {
			frameworkId,
			sectionTitle,
			name,
			status,
			...(description && { description }),
			...(exclusionJustification && { exclusionJustification }),
		},
	};

	const responseData = await proboApiRequest.call(this, query, variables);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}

