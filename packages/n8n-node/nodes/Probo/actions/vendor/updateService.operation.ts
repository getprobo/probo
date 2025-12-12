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
				operation: ['updateService'],
			},
		},
		default: '',
		description: 'The ID of the vendor service to update',
		required: true,
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['vendor'],
				operation: ['updateService'],
			},
		},
		default: '',
		description: 'The name of the vendor service',
	},
	{
		displayName: 'Description',
		name: 'description',
		type: 'string',
		typeOptions: {
			rows: 4,
		},
		displayOptions: {
			show: {
				resource: ['vendor'],
				operation: ['updateService'],
			},
		},
		default: '',
		description: 'The description of the vendor service',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const vendorServiceId = this.getNodeParameter('vendorServiceId', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex, '') as string;
	const description = this.getNodeParameter('description', itemIndex, '') as string;

	const query = `
		mutation UpdateVendorService($input: UpdateVendorServiceInput!) {
			updateVendorService(input: $input) {
				vendorService {
					id
					name
					description
					createdAt
					updatedAt
				}
			}
		}
	`;

	const input: Record<string, unknown> = { id: vendorServiceId };
	if (name) input.name = name;
	if (description !== undefined) input.description = description === '' ? null : description;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
