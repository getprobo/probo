import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Vendor ID',
		name: 'vendorId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['vendor'],
				operation: ['createContact'],
			},
		},
		default: '',
		description: 'The ID of the vendor',
		required: true,
	},
	{
		displayName: 'Full Name',
		name: 'fullName',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['vendor'],
				operation: ['createContact'],
			},
		},
		default: '',
		description: 'The full name of the contact',
	},
	{
		displayName: 'Email',
		name: 'email',
		type: 'string',
		placeholder: 'name@email.com',
		displayOptions: {
			show: {
				resource: ['vendor'],
				operation: ['createContact'],
			},
		},
		default: '',
		description: 'The email address of the contact',
	},
	{
		displayName: 'Phone',
		name: 'phone',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['vendor'],
				operation: ['createContact'],
			},
		},
		default: '',
		description: 'The phone number of the contact',
	},
	{
		displayName: 'Role',
		name: 'role',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['vendor'],
				operation: ['createContact'],
			},
		},
		default: '',
		description: 'The role of the contact',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const vendorId = this.getNodeParameter('vendorId', itemIndex) as string;
	const fullName = this.getNodeParameter('fullName', itemIndex, '') as string;
	const email = this.getNodeParameter('email', itemIndex, '') as string;
	const phone = this.getNodeParameter('phone', itemIndex, '') as string;
	const role = this.getNodeParameter('role', itemIndex, '') as string;

	const query = `
		mutation CreateVendorContact($input: CreateVendorContactInput!) {
			createVendorContact(input: $input) {
				vendorContactEdge {
					node {
						id
						fullName
						email
						phone
						role
						createdAt
						updatedAt
					}
				}
			}
		}
	`;

	const input: Record<string, unknown> = { vendorId };
	if (fullName) input.fullName = fullName;
	if (email) input.email = email;
	if (phone) input.phone = phone;
	if (role) input.role = role;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
