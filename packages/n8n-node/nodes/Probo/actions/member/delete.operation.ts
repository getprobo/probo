import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboConnectApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Organization ID',
		name: 'organizationId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['member'],
				operation: ['delete'],
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
				operation: ['delete'],
			},
		},
		default: '',
		description: 'The ID of the membership to remove',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const membershipId = this.getNodeParameter('membershipId', itemIndex) as string;

	const query = `
		mutation RemoveMember($input: RemoveMemberInput!) {
			removeMember(input: $input) {
				deletedMembershipId
			}
		}
	`;

	const responseData = await proboConnectApiRequest.call(this, query, {
		input: { organizationId, membershipId },
	});

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
