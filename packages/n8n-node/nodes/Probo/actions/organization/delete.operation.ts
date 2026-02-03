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
				operation: ['delete'],
			},
		},
		default: '',
		description: 'The ID of the organization to delete',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;

	const query = `
		mutation DeleteOrganization($input: DeleteOrganizationInput!) {
			deleteOrganization(input: $input) {
				deletedOrganizationId
			}
		}
	`;

	const responseData = await proboConnectApiRequest.call(this, query, { input: { organizationId } });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}

