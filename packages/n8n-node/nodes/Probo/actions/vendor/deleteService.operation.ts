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
				operation: ['deleteService'],
			},
		},
		default: '',
		description: 'The ID of the vendor service to delete',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const vendorServiceId = this.getNodeParameter('vendorServiceId', itemIndex) as string;

	const query = `
		mutation DeleteVendorService($input: DeleteVendorServiceInput!) {
			deleteVendorService(input: $input) {
				deletedVendorServiceId
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, { input: { vendorServiceId } });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
