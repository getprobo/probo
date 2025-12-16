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
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the meeting to update',
		required: true,
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['meeting'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The name of the meeting',
	},
	{
		displayName: 'Date',
		name: 'date',
		type: 'dateTime',
		displayOptions: {
			show: {
				resource: ['meeting'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The date and time of the meeting',
	},
	{
		displayName: 'Attendee IDs',
		name: 'attendeeIds',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['meeting'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'Comma-separated list of attendee IDs (People IDs)',
	},
	{
		displayName: 'Minutes',
		name: 'minutes',
		type: 'string',
		typeOptions: {
			rows: 4,
		},
		displayOptions: {
			show: {
				resource: ['meeting'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The minutes of the meeting',
	},
	{
		displayName: 'Options',
		name: 'options',
		type: 'collection',
		placeholder: 'Add Option',
		default: {},
		displayOptions: {
			show: {
				resource: ['meeting'],
				operation: ['update'],
			},
		},
		options: [
			{
				displayName: 'Include Attendees',
				name: 'includeAttendees',
				type: 'boolean',
				default: false,
				description: 'Whether to include attendees in the response',
			},
			{
				displayName: 'Include Organization',
				name: 'includeOrganization',
				type: 'boolean',
				default: false,
				description: 'Whether to include organization in the response',
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const meetingId = this.getNodeParameter('meetingId', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex, '') as string;
	const date = this.getNodeParameter('date', itemIndex, '') as string;
	const attendeeIdsStr = this.getNodeParameter('attendeeIds', itemIndex, '') as string;
	const minutes = this.getNodeParameter('minutes', itemIndex, '') as string;
	const options = this.getNodeParameter('options', itemIndex, {}) as {
		includeAttendees?: boolean;
		includeOrganization?: boolean;
	};

	const attendeesFragment = options.includeAttendees
		? `attendees {
			id
			fullName
		}`
		: '';

	const organizationFragment = options.includeOrganization
		? `organization {
			id
			name
		}`
		: '';

	const query = `
		mutation UpdateMeeting($input: UpdateMeetingInput!) {
			updateMeeting(input: $input) {
				meeting {
					id
					name
					date
					minutes
					${attendeesFragment}
					${organizationFragment}
					createdAt
					updatedAt
				}
			}
		}
	`;

	const attendeeIds = attendeeIdsStr ? attendeeIdsStr.split(',').map((id) => id.trim()).filter(Boolean) : undefined;

	const input: Record<string, unknown> = { meetingId };
	if (name) input.name = name;
	if (date) input.date = new Date(date).toISOString();
	if (attendeeIds && attendeeIds.length > 0) {
		input.attendeeIds = attendeeIds;
	}
	if (minutes !== undefined && minutes !== null) {
		input.minutes = minutes;
	}

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
