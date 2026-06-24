// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';
import { WEBHOOK_EVENT_OPTIONS } from './events';

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
		options: WEBHOOK_EVENT_OPTIONS,
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
