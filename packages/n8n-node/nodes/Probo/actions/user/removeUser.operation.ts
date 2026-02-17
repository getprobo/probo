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
				operation: ['removeUser'],
			},
		},
		default: '',
		description: 'The ID of the organization',
		required: true,
	},
	{
		displayName: 'User ID',
		name: 'userId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['user'],
				operation: ['removeUser'],
			},
		},
		default: '',
		description: 'The ID of the user (profile) to remove from the organization',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const userId = this.getNodeParameter('userId', itemIndex) as string;

	const query = `
		mutation RemoveUser($input: RemoveUserInput!) {
			removeUser(input: $input) {
				deletedProfileId
			}
		}
	`;

	const input = { organizationId, profileId: userId };
	const responseData = await proboConnectApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
