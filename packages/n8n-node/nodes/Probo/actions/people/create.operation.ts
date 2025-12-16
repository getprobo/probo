import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Organization ID',
		name: 'organizationId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['people'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The ID of the organization',
		required: true,
	},
	{
		displayName: 'Full Name',
		name: 'fullName',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['people'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The full name of the person',
		required: true,
	},
	{
		displayName: 'Primary Email Address',
		name: 'primaryEmailAddress',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['people'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The primary email address of the person',
		required: true,
	},
	{
		displayName: 'Kind',
		name: 'kind',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['people'],
				operation: ['create'],
			},
		},
		options: [
			{
				name: 'Employee',
				value: 'EMPLOYEE',
			},
			{
				name: 'Contractor',
				value: 'CONTRACTOR',
			},
			{
				name: 'Service Account',
				value: 'SERVICE_ACCOUNT',
			},
		],
		default: 'EMPLOYEE',
		description: 'The kind of person',
		required: true,
	},
	{
		displayName: 'Additional Fields',
		name: 'additionalFields',
		type: 'collection',
		placeholder: 'Add Field',
		default: {},
		displayOptions: {
			show: {
				resource: ['people'],
				operation: ['create'],
			},
		},
		options: [
			{
				displayName: 'Additional Email Addresses',
				name: 'additionalEmailAddresses',
				type: 'string',
				default: '',
				description: 'Comma-separated list of additional email addresses',
			},
			{
				displayName: 'Position',
				name: 'position',
				type: 'string',
				default: '',
				description: 'The position of the person',
			},
		{
			displayName: 'Contract Start Date',
			name: 'contractStartDate',
			type: 'dateTime',
			default: '',
		},
		{
			displayName: 'Contract End Date',
			name: 'contractEndDate',
			type: 'dateTime',
			default: '',
		},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const fullName = this.getNodeParameter('fullName', itemIndex) as string;
	const primaryEmailAddress = this.getNodeParameter('primaryEmailAddress', itemIndex) as string;
	const kind = this.getNodeParameter('kind', itemIndex) as string;
	const additionalFields = this.getNodeParameter('additionalFields', itemIndex, {}) as {
		additionalEmailAddresses?: string;
		position?: string;
		contractStartDate?: string;
		contractEndDate?: string;
	};

	const query = `
		mutation CreatePeople($input: CreatePeopleInput!) {
			createPeople(input: $input) {
				peopleEdge {
					node {
						id
						fullName
						primaryEmailAddress
						additionalEmailAddresses
						kind
						position
						contractStartDate
						contractEndDate
						createdAt
						updatedAt
					}
				}
			}
		}
	`;

	const input: Record<string, unknown> = {
		organizationId,
		fullName,
		primaryEmailAddress,
		kind,
	};
	if (additionalFields.additionalEmailAddresses) {
		input.additionalEmailAddresses = additionalFields.additionalEmailAddresses.split(',').map((e) => e.trim()).filter(Boolean);
	}
	if (additionalFields.position) input.position = additionalFields.position;
	if (additionalFields.contractStartDate) input.contractStartDate = new Date(additionalFields.contractStartDate).toISOString();
	if (additionalFields.contractEndDate) input.contractEndDate = new Date(additionalFields.contractEndDate).toISOString();

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
