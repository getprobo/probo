import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboConnectApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Membership ID',
		name: 'membershipId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['member'],
				operation: ['get'],
			},
		},
		default: '',
		description: 'The ID of the membership',
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
				operation: ['get'],
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
	const membershipId = this.getNodeParameter('membershipId', itemIndex) as string;
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
						emailVerified
					}`
			: '';

	const profileFragment =
		options.includeProfile !== false
			? `profile {
						id
						fullName
						createdAt
						updatedAt
					}`
			: '';

	const organizationFragment =
		options.includeOrganization !== false
			? `organization {
						id
						name
						email
					}`
			: '';

	const query = `
		query GetMember($membershipId: ID!) {
			node(id: $membershipId) {
				... on Membership {
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

	const variables = {
		membershipId,
	};

	const responseData = await proboConnectApiRequest.call(this, query, variables);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
