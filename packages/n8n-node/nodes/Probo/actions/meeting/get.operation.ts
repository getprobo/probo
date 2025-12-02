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
				operation: ['get'],
			},
		},
		default: '',
		description: 'The ID of the meeting',
		required: true,
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
				operation: ['get'],
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
		query GetMeeting($meetingId: ID!) {
			node(id: $meetingId) {
				... on Meeting {
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

	const variables = {
		meetingId,
	};

	const responseData = await proboApiRequest.call(this, query, variables);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}

