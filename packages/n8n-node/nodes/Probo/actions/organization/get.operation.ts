import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboConnectApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Organization ID',
		name: 'organizationId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['organization'],
				operation: ['get'],
			},
		},
		default: '',
		description: 'The ID of the organization',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;

	const query = `
		query GetOrganization($organizationId: ID!) {
			node(id: $organizationId) {
				... on Organization {
					id
					name
					description
					websiteUrl
					email
					headquarterAddress
					logoUrl
					horizontalLogoUrl
					createdAt
					updatedAt
				}
			}
		}
	`;

	const variables = {
		organizationId,
	};

	const responseData = await proboConnectApiRequest.call(this, query, variables);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}

