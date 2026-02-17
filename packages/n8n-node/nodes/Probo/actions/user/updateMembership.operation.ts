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
				resource: ['user'],
				operation: ['updateMembership'],
			},
		},
		default: '',
		description: 'The ID of the organization',
		required: true,
	},
	{
		displayName: 'Membership ID',
		name: 'membershipId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['user'],
				operation: ['updateMembership'],
			},
		},
		default: '',
		description: 'The ID of the membership to update',
		required: true,
	},
	{
		displayName: 'Role',
		name: 'role',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['user'],
				operation: ['updateMembership'],
			},
		},
		options: roleOptions,
		default: 'EMPLOYEE',
		description: 'New role for the membership',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const membershipId = this.getNodeParameter('membershipId', itemIndex) as string;
	const role = this.getNodeParameter('role', itemIndex) as string;

	const query = `
		mutation UpdateMembership($input: UpdateMembershipInput!) {
			updateMembership(input: $input) {
				membership {
					id
					role
					createdAt
				}
			}
		}
	`;

	const input = { organizationId, membershipId, role };
	const responseData = await proboConnectApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
