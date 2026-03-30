// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Organization ID',
		name: 'organizationId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['meeting'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The ID of the organization',
		required: true,
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['meeting'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The name of the meeting',
		required: true,
	},
	{
		displayName: 'Date',
		name: 'date',
		type: 'dateTime',
		displayOptions: {
			show: {
				resource: ['meeting'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The date and time of the meeting',
		required: true,
	},
	{
		displayName: 'Attendee IDs',
		name: 'attendeeIds',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['meeting'],
				operation: ['create'],
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
				operation: ['create'],
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
				operation: ['create'],
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
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex) as string;
	const date = this.getNodeParameter('date', itemIndex) as string;
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
		mutation CreateMeeting($input: CreateMeetingInput!) {
			createMeeting(input: $input) {
				meetingEdge {
					node {
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
		}
	`;

	const attendeeIds = attendeeIdsStr ? attendeeIdsStr.split(',').map((id) => id.trim()).filter(Boolean) : undefined;

	const input: Record<string, unknown> = {
		organizationId,
		name,
		date: new Date(date).toISOString(),
	};
	if (attendeeIds && attendeeIds.length > 0) {
		input.attendeeIds = attendeeIds;
	}
	if (minutes) {
		input.minutes = minutes;
	}

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
