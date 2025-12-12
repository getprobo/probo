import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Vendor Service ID',
		name: 'vendorServiceId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['vendor'],
				operation: ['getService'],
			},
		},
		default: '',
		description: 'The ID of the vendor service',
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
				operation: ['getService'],
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
	const vendorServiceId = this.getNodeParameter('vendorServiceId', itemIndex) as string;
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
		query GetVendorService($vendorServiceId: ID!) {
			node(id: $vendorServiceId) {
				... on VendorService {
					id
					name
					description
					${vendorFragment}
					createdAt
					updatedAt
				}
			}
		}
	`;

	const variables = {
		vendorServiceId,
	};

	const responseData = await proboApiRequest.call(this, query, variables);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
