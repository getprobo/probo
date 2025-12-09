import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'People ID',
		name: 'peopleId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['people'],
				operation: ['delete'],
			},
		},
		default: '',
		description: 'The ID of the person to delete',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const peopleId = this.getNodeParameter('peopleId', itemIndex) as string;

	const query = `
		mutation DeletePeople($input: DeletePeopleInput!) {
			deletePeople(input: $input) {
				deletedPeopleId
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, { input: { peopleId } });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
