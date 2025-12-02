import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['organization'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The name of the organization',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const name = this.getNodeParameter('name', itemIndex) as string;

	const query = `
		mutation CreateOrganization($input: CreateOrganizationInput!) {
			createOrganization(input: $input) {
				organizationEdge {
					node {
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
		}
	`;

	const variables = {
		input: {
			name,
		},
	};

	const responseData = await proboApiRequest.call(this, query, variables);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}

