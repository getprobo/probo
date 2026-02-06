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
				operation: ['update'],
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
				resource: ['member'],
				operation: ['update'],
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
				resource: ['member'],
				operation: ['update'],
			},
		},
		options: roleOptions,
		default: 'EMPLOYEE',
		description: 'New role for the member',
		required: true,
	},
	{
		displayName: 'Options',
		name: 'options',
		type: 'collection',
		placeholder: 'Add Option',
		default: {},
		displayOptions: {
			show: {
				resource: ['member'],
				operation: ['update'],
			},
		},
		options: [
			{
				displayName: 'Include Identity',
				name: 'includeIdentity',
				type: 'boolean',
				default: true,
				description: 'Whether to include identity (email, fullName) in the response',
			},
			{
				displayName: 'Include Profile',
				name: 'includeProfile',
				type: 'boolean',
				default: true,
				description: 'Whether to include profile in the response',
			},
			{
				displayName: 'Include Organization',
				name: 'includeOrganization',
				type: 'boolean',
				default: true,
				description: 'Whether to include organization in the response',
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const membershipId = this.getNodeParameter('membershipId', itemIndex) as string;
	const role = this.getNodeParameter('role', itemIndex) as string;
	const options = this.getNodeParameter('options', itemIndex, {}) as {
		includeIdentity?: boolean;
		includeProfile?: boolean;
		includeOrganization?: boolean;
	};

	const identityFragment =
		options.includeIdentity !== false
			? `identity {
						id
						email
						fullName
					}`
			: '';

	const profileFragment =
		options.includeProfile !== false
			? `profile {
						id
						fullName
					}`
			: '';

	const organizationFragment =
		options.includeOrganization !== false
			? `organization {
						id
						name
					}`
			: '';

	const query = `
		mutation UpdateMembership($input: UpdateMembershipInput!) {
			updateMembership(input: $input) {
				membership {
					id
					createdAt
					role
					source
					state
					${identityFragment}
					${profileFragment}
					${organizationFragment}
				}
			}
		}
	`;

	const input = {
		organizationId,
		membershipId,
		role,
	};

	const responseData = await proboConnectApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
