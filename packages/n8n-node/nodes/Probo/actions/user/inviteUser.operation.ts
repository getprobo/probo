import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboConnectApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Organization ID',
		name: 'organizationId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['user'],
				operation: ['inviteUser'],
			},
		},
		default: '',
		description: 'The ID of the organization',
		required: true,
	},
	{
		displayName: 'Profile ID',
		name: 'profileId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['user'],
				operation: ['inviteUser'],
			},
		},
		default: '',
		description: 'The ID of the user (profile) to invite to the organization',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const profileId = this.getNodeParameter('profileId', itemIndex) as string;

	const query = `
		mutation InviteUser($input: InviteUserInput!) {
			inviteUser(input: $input) {
				invitationEdge {
					node {
						id
						expiresAt
						status
						createdAt
						user { id fullName emailAddress }
						organization { id name }
					}
				}
			}
		}
	`;

	const input = { organizationId, profileId };
	const responseData = await proboConnectApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
