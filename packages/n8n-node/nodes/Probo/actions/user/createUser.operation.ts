import type {
	INodeProperties,
	IExecuteFunctions,
	INodeExecutionData,
	IDataObject,
} from 'n8n-workflow';
import { proboConnectApiRequest } from '../../GenericFunctions';

const roleOptions = [
	{ name: 'Owner', value: 'OWNER' },
	{ name: 'Admin', value: 'ADMIN' },
	{ name: 'Employee', value: 'EMPLOYEE' },
	{ name: 'Viewer', value: 'VIEWER' },
	{ name: 'Auditor', value: 'AUDITOR' },
];

const kindOptions = [
	{ name: 'Employee', value: 'EMPLOYEE' },
	{ name: 'Contractor', value: 'CONTRACTOR' },
	{ name: 'Service Account', value: 'SERVICE_ACCOUNT' },
];

export const description: INodeProperties[] = [
	{
		displayName: 'Organization ID',
		name: 'organizationId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['user'],
				operation: ['createUser'],
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
				resource: ['user'],
				operation: ['createUser'],
			},
		},
		default: '',
		description: 'Full name of the user',
		required: true,
	},
	{
		displayName: 'Email Address',
		name: 'emailAddress',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['user'],
				operation: ['createUser'],
			},
		},
		default: '',
		placeholder: 'name@example.com',
		description: 'Email address of the user',
		required: true,
	},
	{
		displayName: 'Role',
		name: 'role',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['user'],
				operation: ['createUser'],
			},
		},
		options: roleOptions,
		default: 'EMPLOYEE',
		description: 'Membership role to assign',
		required: true,
	},
	{
		displayName: 'Kind',
		name: 'kind',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['user'],
				operation: ['createUser'],
			},
		},
		options: kindOptions,
		default: 'EMPLOYEE',
		description: 'User kind (employee, contractor, or service account)',
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
				operation: ['createUser'],
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
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const fullName = this.getNodeParameter('fullName', itemIndex) as string;
	const emailAddress = this.getNodeParameter('emailAddress', itemIndex) as string;
	const role = this.getNodeParameter('role', itemIndex) as string;
	const kind = this.getNodeParameter('kind', itemIndex) as string;
	const additionalFields = this.getNodeParameter('additionalFields', itemIndex, {}) as {
		additionalEmailAddresses?: string;
		position?: string;
		contractStartDate?: string;
		contractEndDate?: string;
	};

	const input: IDataObject = {
		organizationId,
		fullName,
		emailAddress,
		role,
		kind,
	};
	if (additionalFields.additionalEmailAddresses) {
		input.additionalEmailAddresses = additionalFields.additionalEmailAddresses
			.split(',')
			.map((e: string) => e.trim())
			.filter(Boolean);
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
		mutation CreateUser($input: CreateUserInput!) {
			createUser(input: $input) {
				profileEdge {
					node {
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
		}
	`;

	const responseData = await proboConnectApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
