import type {
	INodeProperties,
	IExecuteFunctions,
	INodeExecutionData,
	IDataObject,
} from 'n8n-workflow';
import { proboConnectApiRequest } from '../../GenericFunctions';

const kindOptions = [
	{ name: 'Employee', value: 'EMPLOYEE' },
	{ name: 'Contractor', value: 'CONTRACTOR' },
	{ name: 'Service Account', value: 'SERVICE_ACCOUNT' },
];

export const description: INodeProperties[] = [
	{
		displayName: 'User ID',
		name: 'userId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['user'],
				operation: ['updateUser'],
			},
		},
		default: '',
		description: 'The ID of the user (profile) to update',
		required: true,
	},
	{
		displayName: 'Full Name',
		name: 'fullName',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['user'],
				operation: ['updateUser'],
			},
		},
		default: '',
		description: 'Full name of the user',
		required: true,
	},
	{
		displayName: 'Kind',
		name: 'kind',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['user'],
				operation: ['updateUser'],
			},
		},
		options: kindOptions,
		default: 'EMPLOYEE',
		description: 'User kind',
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
				resource: ['user'],
				operation: ['updateUser'],
			},
		},
		options: [
			{
				displayName: 'Additional Email Addresses',
				name: 'additionalEmailAddresses',
				type: 'string',
				default: '',
				description: 'Comma-separated additional email addresses',
			},
			{
				displayName: 'Position',
				name: 'position',
				type: 'string',
				default: '',
				description: 'Job or role position',
			},
			{
				displayName: 'Contract Start Date',
				name: 'contractStartDate',
				type: 'string',
				default: '',
				description: 'Contract start date (ISO 8601)',
			},
			{
				displayName: 'Contract End Date',
				name: 'contractEndDate',
				type: 'string',
				default: '',
				description: 'Contract end date (ISO 8601)',
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const userId = this.getNodeParameter('userId', itemIndex) as string;
	const fullName = this.getNodeParameter('fullName', itemIndex) as string;
	const kind = this.getNodeParameter('kind', itemIndex) as string;
	const additionalFields = this.getNodeParameter('additionalFields', itemIndex, {}) as {
		additionalEmailAddresses?: string;
		position?: string;
		contractStartDate?: string;
		contractEndDate?: string;
	};

	const input: IDataObject = {
		id: userId,
		fullName,
		kind,
	};
	if (additionalFields.additionalEmailAddresses !== undefined) {
		input.additionalEmailAddresses = additionalFields.additionalEmailAddresses
			? (additionalFields.additionalEmailAddresses as string)
					.split(',')
					.map((e: string) => e.trim())
					.filter(Boolean)
			: [];
	}
	if (additionalFields.position) {
		input.position = additionalFields.position;
	}
	if (additionalFields.contractStartDate) {
		input.contractStartDate = additionalFields.contractStartDate;
	}
	if (additionalFields.contractEndDate) {
		input.contractEndDate = additionalFields.contractEndDate;
	}

	const query = `
		mutation UpdateUser($input: UpdateUserInput!) {
			updateUser(input: $input) {
				profile {
					id
					fullName
					emailAddress
					kind
					position
					contractStartDate
					contractEndDate
					createdAt
					updatedAt
					organization { id name }
					membership { id role createdAt }
				}
			}
		}
	`;

	const responseData = await proboConnectApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
