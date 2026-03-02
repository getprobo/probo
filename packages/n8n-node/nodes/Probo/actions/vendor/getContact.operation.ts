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
				operation: ['getContact'],
			},
		},
		default: '',
		description: 'The ID of the vendor contact',
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
				resource: ['vendor'],
				operation: ['getContact'],
			},
		},
		options: [
			{
				displayName: 'Include Vendor',
				name: 'includeVendor',
				type: 'boolean',
				default: false,
				description: 'Whether to include vendor in the response',
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const vendorContactId = this.getNodeParameter('vendorContactId', itemIndex) as string;
	const options = this.getNodeParameter('options', itemIndex, {}) as {
		includeVendor?: boolean;
	};

	const vendorFragment = options.includeVendor
		? `vendor {
			id
			name
		}`
		: '';

	const query = `
		query GetVendorContact($vendorContactId: ID!) {
			node(id: $vendorContactId) {
				... on VendorContact {
					id
					fullName
					email
					phone
					role
					${vendorFragment}
					createdAt
					updatedAt
				}
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, { vendorContactId });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
