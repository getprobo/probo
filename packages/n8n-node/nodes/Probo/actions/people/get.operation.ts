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
				operation: ['get'],
			},
		},
		default: '',
		description: 'The ID of the person',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const peopleId = this.getNodeParameter('peopleId', itemIndex) as string;

	const query = `
		query GetPeople($peopleId: ID!) {
			node(id: $peopleId) {
				... on People {
					id
					fullName
					primaryEmailAddress
					additionalEmailAddresses
					kind
					position
					contractStartDate
					contractEndDate
					createdAt
					updatedAt
				}
			}
		}
	`;

	const variables = {
		peopleId,
	};

	const responseData = await proboApiRequest.call(this, query, variables);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
