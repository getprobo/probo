import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'People ID',
		name: 'peopleId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['people'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the person to update',
		required: true,
	},
	{
		displayName: 'Full Name',
		name: 'fullName',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['people'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The full name of the person',
	},
	{
		displayName: 'Primary Email Address',
		name: 'primaryEmailAddress',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['people'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The primary email address of the person',
	},
	{
		displayName: 'Kind',
		name: 'kind',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['people'],
				operation: ['update'],
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
				operation: ['update'],
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
	const peopleId = this.getNodeParameter('peopleId', itemIndex) as string;
	const fullName = this.getNodeParameter('fullName', itemIndex, '') as string;
	const primaryEmailAddress = this.getNodeParameter('primaryEmailAddress', itemIndex, '') as string;
	const kind = this.getNodeParameter('kind', itemIndex, '') as string;
	const additionalFields = this.getNodeParameter('additionalFields', itemIndex, {}) as {
		additionalEmailAddresses?: string;
		position?: string;
		contractStartDate?: string;
		contractEndDate?: string;
	};

	const query = `
		mutation UpdatePeople($input: UpdatePeopleInput!) {
			updatePeople(input: $input) {
				people {
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
	`;

	const input: Record<string, unknown> = { id: peopleId };
	if (fullName) input.fullName = fullName;
	if (primaryEmailAddress) input.primaryEmailAddress = primaryEmailAddress;
	if (kind) input.kind = kind;
	if (additionalFields.additionalEmailAddresses !== undefined) {
		if (additionalFields.additionalEmailAddresses === '') {
			input.additionalEmailAddresses = [];
		} else {
			input.additionalEmailAddresses = additionalFields.additionalEmailAddresses.split(',').map((e) => e.trim()).filter(Boolean);
		}
	}
	if (additionalFields.position !== undefined) input.position = additionalFields.position === '' ? null : additionalFields.position;
	if (additionalFields.contractStartDate !== undefined) input.contractStartDate = additionalFields.contractStartDate === '' ? null : new Date(additionalFields.contractStartDate).toISOString();
	if (additionalFields.contractEndDate !== undefined) input.contractEndDate = additionalFields.contractEndDate === '' ? null : new Date(additionalFields.contractEndDate).toISOString();

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
