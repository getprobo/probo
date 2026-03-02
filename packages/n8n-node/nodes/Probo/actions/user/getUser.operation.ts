import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboConnectApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'User ID',
		name: 'userId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['user'],
				operation: ['getUser'],
			},
		},
		default: '',
		description: 'The ID of the user (profile)',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const userId = this.getNodeParameter('userId', itemIndex) as string;

	const query = `
		query GetUser($userId: ID!) {
			node(id: $userId) {
				... on Profile {
					id
					fullName
					emailAddress
					source
					state
					additionalEmailAddresses
					kind
					position
					contractStartDate
					contractEndDate
					createdAt
					updatedAt
					identity { id email fullName emailVerified }
					organization { id name email }
					membership { id role createdAt }
				}
			}
		}
	`;

	const responseData = await proboConnectApiRequest.call(this, query, { userId });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
