import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Framework ID',
		name: 'frameworkId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['framework'],
				operation: ['delete'],
			},
		},
		default: '',
		description: 'The ID of the framework to delete',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const frameworkId = this.getNodeParameter('frameworkId', itemIndex) as string;

	const query = `
		mutation DeleteFramework($input: DeleteFrameworkInput!) {
			deleteFramework(input: $input) {
				deletedFrameworkId
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, { input: { frameworkId } });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}

