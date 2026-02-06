import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboConnectApiRequest } from '../../GenericFunctions';

const roleOptions = [
	{ name: 'Owner', value: 'OWNER' },
	{ name: 'Admin', value: 'ADMIN' },
	{ name: 'Employee', value: 'EMPLOYEE' },
	{ name: 'Viewer', value: 'VIEWER' },
	{ name: 'Auditor', value: 'AUDITOR' },
];

export const description: INodeProperties[] = [
	{
		displayName: 'Organization ID',
		name: 'organizationId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['member'],
				operation: ['invite'],
			},
		},
		default: '',
		description: 'The ID of the organization',
		required: true,
	},
	{
		displayName: 'Email',
		name: 'email',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['member'],
				operation: ['invite'],
			},
		},
		default: '',
		placeholder: 'name@email.com',
		description: 'Email address of the person to invite',
		required: true,
	},
	{
		displayName: 'Full Name',
		name: 'fullName',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['member'],
				operation: ['invite'],
			},
		},
		default: '',
		description: 'Full name of the person to invite',
		required: true,
	},
	{
		displayName: 'Role',
		name: 'role',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['member'],
				operation: ['invite'],
			},
		},
		options: roleOptions,
		default: 'EMPLOYEE',
		description: 'Role to assign to the member',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const email = this.getNodeParameter('email', itemIndex) as string;
	const fullName = this.getNodeParameter('fullName', itemIndex) as string;
	const role = this.getNodeParameter('role', itemIndex) as string;

	const query = `
		mutation InviteMember($input: InviteMemberInput!) {
			inviteMember(input: $input) {
				invitationEdge {
					node {
						id
						email
						fullName
						role
						expiresAt
						status
						createdAt
					}
				}
			}
		}
	`;

	const input = {
		organizationId,
		email,
		fullName,
		role,
	};

	const responseData = await proboConnectApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
