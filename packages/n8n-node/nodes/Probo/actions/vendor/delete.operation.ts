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
				operation: ['delete'],
			},
		},
		default: '',
		description: 'The ID of the vendor to delete',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const vendorId = this.getNodeParameter('vendorId', itemIndex) as string;

	const query = `
		mutation DeleteVendor($input: DeleteVendorInput!) {
			deleteVendor(input: $input) {
				deletedVendorId
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, { input: { vendorId } });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}

