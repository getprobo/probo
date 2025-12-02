import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Control ID',
		name: 'controlId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['control'],
				operation: ['delete'],
			},
		},
		default: '',
		description: 'The ID of the control to delete',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const controlId = this.getNodeParameter('controlId', itemIndex) as string;

	const query = `
		mutation DeleteControl($input: DeleteControlInput!) {
			deleteControl(input: $input) {
				deletedControlId
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, { input: { controlId } });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}

