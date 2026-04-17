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
				resource: ['webhook'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The ID of the organization',
		required: true,
	},
	{
		displayName: 'Endpoint URL',
		name: 'endpointUrl',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['webhook'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The HTTPS endpoint URL that receives webhook events',
		required: true,
	},
	{
		displayName: 'Selected Events',
		name: 'selectedEvents',
		type: 'multiOptions',
		displayOptions: {
			show: {
				resource: ['webhook'],
				operation: ['create'],
			},
		},
		options: [
			{ name: 'Meeting Created', value: 'MEETING_CREATED' },
			{ name: 'Meeting Deleted', value: 'MEETING_DELETED' },
			{ name: 'Meeting Updated', value: 'MEETING_UPDATED' },
			{ name: 'Obligation Created', value: 'OBLIGATION_CREATED' },
			{ name: 'Obligation Deleted', value: 'OBLIGATION_DELETED' },
			{ name: 'Obligation Updated', value: 'OBLIGATION_UPDATED' },
			{ name: 'User Created', value: 'USER_CREATED' },
			{ name: 'User Deleted', value: 'USER_DELETED' },
			{ name: 'User Updated', value: 'USER_UPDATED' },
			{ name: 'Vendor Created', value: 'VENDOR_CREATED' },
			{ name: 'Vendor Deleted', value: 'VENDOR_DELETED' },
			{ name: 'Vendor Updated', value: 'VENDOR_UPDATED' },
		],
		default: [],
		description: 'The event types to subscribe to',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const endpointUrl = this.getNodeParameter('endpointUrl', itemIndex) as string;
	const selectedEvents = this.getNodeParameter('selectedEvents', itemIndex) as string[];

	const query = `
		mutation CreateWebhookSubscription($input: CreateWebhookSubscriptionInput!) {
			createWebhookSubscription(input: $input) {
				webhookSubscriptionEdge {
					node {
						id
						endpointUrl
						signingSecret
						selectedEvents
						createdAt
						updatedAt
					}
				}
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, {
		input: {
			organizationId,
			endpointUrl,
			selectedEvents,
		},
	});

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
