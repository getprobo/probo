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
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the organization to update',
		required: true,
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['organization'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The name of the organization',
	},
	{
		displayName: 'Description',
		name: 'description',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['organization'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The description of the organization',
	},
	{
		displayName: 'Website URL',
		name: 'websiteUrl',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['organization'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The website URL of the organization',
	},
	{
		displayName: 'Email',
		name: 'email',
		type: 'string',
		placeholder: 'name@example.com',
		displayOptions: {
			show: {
				resource: ['organization'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The email address of the organization',
	},
	{
		displayName: 'Headquarter Address',
		name: 'headquarterAddress',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['organization'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The headquarter address of the organization',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex, '') as string;
	const description = this.getNodeParameter('description', itemIndex, '') as string;
	const websiteUrl = this.getNodeParameter('websiteUrl', itemIndex, '') as string;
	const email = this.getNodeParameter('email', itemIndex, '') as string;
	const headquarterAddress = this.getNodeParameter('headquarterAddress', itemIndex, '') as string;

	const query = `
		mutation UpdateOrganization($input: UpdateOrganizationInput!) {
			updateOrganization(input: $input) {
				organization {
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

	const input: Record<string, string> = { organizationId };
	if (name) input.name = name;
	if (description) input.description = description;
	if (websiteUrl) input.websiteUrl = websiteUrl;
	if (email) input.email = email;
	if (headquarterAddress) input.headquarterAddress = headquarterAddress;

	const responseData = await proboConnectApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}

