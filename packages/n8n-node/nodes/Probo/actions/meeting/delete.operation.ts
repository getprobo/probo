import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Meeting ID',
		name: 'meetingId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['meeting'],
				operation: ['delete'],
			},
		},
		default: '',
		description: 'The ID of the meeting to delete',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const meetingId = this.getNodeParameter('meetingId', itemIndex) as string;

	const query = `
		mutation DeleteMeeting($input: DeleteMeetingInput!) {
			deleteMeeting(input: $input) {
				deletedMeetingId
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, { input: { meetingId } });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}

