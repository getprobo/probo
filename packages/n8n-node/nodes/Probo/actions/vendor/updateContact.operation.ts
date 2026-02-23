import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Vendor Contact ID',
		name: 'vendorContactId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['vendor'],
				operation: ['updateContact'],
			},
		},
		default: '',
		description: 'The ID of the vendor contact to update',
		required: true,
	},
	{
		displayName: 'Additional Fields',
		name: 'additionalFields',
		type: 'collection',
		placeholder: 'Add Field',
		default: {},
		displayOptions: {
			show: {
				resource: ['vendor'],
				operation: ['updateContact'],
			},
		},
		options: [
			{
				displayName: 'Email',
				name: 'email',
				type: 'string',
				placeholder: 'name@email.com',
				default: '',
				description: 'The email address of the contact',
			},
			{
				displayName: 'Full Name',
				name: 'fullName',
				type: 'string',
				default: '',
				description: 'The full name of the contact',
			},
			{
				displayName: 'Phone',
				name: 'phone',
				type: 'string',
				default: '',
				description: 'The phone number of the contact',
			},
			{
				displayName: 'Role',
				name: 'role',
				type: 'string',
				default: '',
				description: 'The role of the contact',
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const vendorContactId = this.getNodeParameter('vendorContactId', itemIndex) as string;
	const additionalFields = this.getNodeParameter('additionalFields', itemIndex, {}) as {
		fullName?: string;
		email?: string;
		phone?: string;
		role?: string;
	};

	const query = `
		mutation UpdateVendorContact($input: UpdateVendorContactInput!) {
			updateVendorContact(input: $input) {
				vendorContact {
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
	`;

	const input: Record<string, unknown> = { id: vendorContactId };
	if (additionalFields.fullName !== undefined) input.fullName = additionalFields.fullName === '' ? null : additionalFields.fullName;
	if (additionalFields.email !== undefined) input.email = additionalFields.email === '' ? null : additionalFields.email;
	if (additionalFields.phone !== undefined) input.phone = additionalFields.phone === '' ? null : additionalFields.phone;
	if (additionalFields.role !== undefined) input.role = additionalFields.role === '' ? null : additionalFields.role;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
