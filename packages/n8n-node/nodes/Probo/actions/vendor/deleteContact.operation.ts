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
				operation: ['deleteContact'],
			},
		},
		default: '',
		description: 'The ID of the vendor contact to delete',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const vendorContactId = this.getNodeParameter('vendorContactId', itemIndex) as string;

	const query = `
		mutation DeleteVendorContact($input: DeleteVendorContactInput!) {
			deleteVendorContact(input: $input) {
				deletedVendorContactId
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, { input: { vendorContactId } });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
