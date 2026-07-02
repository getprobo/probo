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

import { createHmac, timingSafeEqual } from 'crypto';
import {
	NodeApiError,
	NodeConnectionTypes,
	type IDataObject,
	type IHookFunctions,
	type INodeType,
	type INodeTypeDescription,
	type IWebhookFunctions,
	type IWebhookResponseData,
	type JsonObject,
} from 'n8n-workflow';
import { proboApiRequest } from '../Probo/GenericFunctions';
import { WEBHOOK_EVENT_OPTIONS } from '../Probo/actions/webhook/events';

function extractNode(response: IDataObject): IDataObject | undefined {
	const data = response?.data as IDataObject | undefined;
	return data?.node as IDataObject | undefined;
}

function sameEventSet(a: string[], b: string[]): boolean {
	if (a.length !== b.length) {
		return false;
	}

	const set = new Set(a);
	return b.every((value) => set.has(value));
}

async function deleteSubscription(this: IHookFunctions, subscriptionId: string): Promise<boolean> {
	const query = `
		mutation DeleteWebhookSubscription($input: DeleteWebhookSubscriptionInput!) {
			deleteWebhookSubscription(input: $input) {
				deletedWebhookSubscriptionId
			}
		}
	`;

	try {
		await proboApiRequest.call(this, query, {
			input: { webhookSubscriptionId: subscriptionId },
		});
	} catch {
		return false;
	}

	return true;
}

export class ProboTrigger implements INodeType {
	description: INodeTypeDescription = {
		displayName: 'Probo Trigger',
		name: 'proboTrigger',
		icon: { light: 'file:../../icons/probo-light.svg', dark: 'file:../../icons/probo.svg' },
		group: ['trigger'],
		version: 1,
		subtitle: '={{$parameter["events"].join(", ")}}',
		description: 'Starts a workflow when Probo events occur',
		usableAsTool: true,
		defaults: {
			name: 'Probo Trigger',
		},
		inputs: [],
		outputs: [NodeConnectionTypes.Main],
		credentials: [
			{
				name: 'proboApi',
				required: true,
			},
		],
		webhooks: [
			{
				name: 'default',
				httpMethod: 'POST',
				responseMode: 'onReceived',
				path: 'webhook',
			},
		],
		properties: [
			{
				displayName: 'Organization ID',
				name: 'organizationId',
				type: 'string',
				default: '',
				required: true,
				description: 'The ID of the organization to receive events from',
			},
			{
				displayName: 'Events',
				name: 'events',
				type: 'multiOptions',
				options: WEBHOOK_EVENT_OPTIONS,
				default: [],
				required: true,
				description: 'The event types that trigger this workflow',
			},
			{
				displayName: 'Verify Signature',
				name: 'verifySignature',
				type: 'boolean',
				default: true,
				description:
					'Whether to reject deliveries whose HMAC-SHA256 signature does not match the subscription signing secret',
			},
			{
				displayName: 'Timestamp Tolerance (Seconds)',
				name: 'toleranceSeconds',
				type: 'number',
				default: 300,
				typeOptions: {
					minValue: 0,
				},
				displayOptions: {
					show: {
						verifySignature: [true],
					},
				},
				description:
					'Reject deliveries whose signed timestamp is older or further in the future than this many seconds, preventing replay of captured deliveries. Set to 0 to disable the freshness check.',
			},
		],
	};

	webhookMethods = {
		default: {
			async checkExists(this: IHookFunctions): Promise<boolean> {
				const webhookData = this.getWorkflowStaticData('node');
				const subscriptionId = webhookData.subscriptionId as string | undefined;

				if (subscriptionId === undefined) {
					return false;
				}

				const query = `
					query CheckWebhookSubscription($id: ID!) {
						node(id: $id) {
							... on WebhookSubscription {
								id
								endpointUrl
								selectedEvents
								organization {
									id
								}
							}
						}
					}
				`;

				const response = await proboApiRequest.call(this, query, { id: subscriptionId });
				const node = extractNode(response);

				if (node?.id === undefined) {
					delete webhookData.subscriptionId;
					delete webhookData.signingSecret;
					return false;
				}

				const webhookUrl = this.getNodeWebhookUrl('default');
				const organizationId = this.getNodeParameter('organizationId') as string;
				const events = this.getNodeParameter('events') as string[];
				const organization = node.organization as IDataObject | undefined;
				const remoteEvents = (node.selectedEvents as string[] | undefined) ?? [];

				const urlMatches = webhookUrl === undefined || node.endpointUrl === webhookUrl;
				const organizationMatches = organization?.id === organizationId;
				const eventsMatch = sameEventSet(remoteEvents, events);

				if (urlMatches && organizationMatches && eventsMatch) {
					return true;
				}

				// A node parameter changed since the subscription was created. The
				// backend cannot move a subscription between organizations, and a stale
				// subscription keeps delivering the old event set to the same n8n URL,
				// so drop it here and let n8n re-create one matching the current config.
				if (!(await deleteSubscription.call(this, subscriptionId))) {
					// Keep the existing subscription rather than creating a second one
					// that would deliver duplicates to the same URL; a later activation
					// retries the cleanup.
					return true;
				}

				delete webhookData.subscriptionId;
				delete webhookData.signingSecret;

				return false;
			},

			async create(this: IHookFunctions): Promise<boolean> {
				const webhookUrl = this.getNodeWebhookUrl('default');
				const organizationId = this.getNodeParameter('organizationId') as string;
				const events = this.getNodeParameter('events') as string[];

				if (webhookUrl === undefined) {
					throw new NodeApiError(this.getNode(), {
						message: 'Cannot create Probo webhook subscription: no webhook URL available',
					} as JsonObject);
				}

				const query = `
					mutation CreateWebhookSubscription($input: CreateWebhookSubscriptionInput!) {
						createWebhookSubscription(input: $input) {
							webhookSubscriptionEdge {
								node {
									id
									signingSecret
								}
							}
						}
					}
				`;

				const response = await proboApiRequest.call(this, query, {
					input: {
						organizationId,
						endpointUrl: webhookUrl,
						selectedEvents: events,
					},
				});

				const data = response?.data as IDataObject | undefined;
				const payload = data?.createWebhookSubscription as IDataObject | undefined;
				const edge = payload?.webhookSubscriptionEdge as IDataObject | undefined;
				const node = edge?.node as IDataObject | undefined;

				if (node?.id === undefined) {
					throw new NodeApiError(this.getNode(), {
						message: 'Cannot create Probo webhook subscription: unexpected API response',
					} as JsonObject);
				}

				const webhookData = this.getWorkflowStaticData('node');
				webhookData.subscriptionId = node.id as string;
				webhookData.signingSecret = node.signingSecret as string;

				return true;
			},

			async delete(this: IHookFunctions): Promise<boolean> {
				const webhookData = this.getWorkflowStaticData('node');
				const subscriptionId = webhookData.subscriptionId as string | undefined;

				if (subscriptionId === undefined) {
					return true;
				}

				if (!(await deleteSubscription.call(this, subscriptionId))) {
					return false;
				}

				delete webhookData.subscriptionId;
				delete webhookData.signingSecret;

				return true;
			},
		},
	};

	async webhook(this: IWebhookFunctions): Promise<IWebhookResponseData> {
		const verifySignature = this.getNodeParameter('verifySignature', true) as boolean;
		const headers = this.getHeaderData() as IDataObject;
		const body = this.getBodyData();

		if (verifySignature) {
			const webhookData = this.getWorkflowStaticData('node');
			const signingSecret = webhookData.signingSecret as string | undefined;
			const signature = headers['x-probo-webhook-signature'] as string | undefined;
			const timestamp = headers['x-probo-webhook-timestamp'] as string | undefined;

			const rejected = this.getResponseObject();

			if (signingSecret === undefined || signature === undefined || timestamp === undefined) {
				rejected.status(403).json({ message: 'Missing Probo webhook signature' });
				return { noWebhookResponse: true };
			}

			// n8n's webhook pipeline calls parseBody() -> req.readRawBody() before
			// invoking this handler for application/json payloads (which Probo always
			// sends), so req.rawBody holds the exact bytes the signature was computed
			// over. Re-serializing the parsed body would not match Go's json.Marshal
			// output, so fail closed if the raw bytes are somehow unavailable.
			const request = this.getRequestObject() as unknown as { rawBody?: Buffer };
			const rawBody = request.rawBody;

			if (rawBody === undefined) {
				rejected
					.status(403)
					.json({ message: 'Probo webhook raw body unavailable for signature verification' });
				return { noWebhookResponse: true };
			}

			const hmac = createHmac('sha256', signingSecret);
			hmac.update(`${timestamp}:`);
			hmac.update(rawBody);
			const expected = hmac.digest('hex');

			const expectedBuffer = Buffer.from(expected, 'hex');
			const receivedBuffer = Buffer.from(signature, 'hex');

			if (
				expectedBuffer.length !== receivedBuffer.length ||
				!timingSafeEqual(expectedBuffer, receivedBuffer)
			) {
				rejected.status(403).json({ message: 'Invalid Probo webhook signature' });
				return { noWebhookResponse: true };
			}

			// The signature authenticates the timestamp, so a valid request could
			// still be replayed verbatim. Reject deliveries outside the tolerance
			// window to bound that replay surface.
			const toleranceSeconds = this.getNodeParameter('toleranceSeconds', 300) as number;
			if (toleranceSeconds > 0) {
				const timestampSeconds = Number(timestamp);
				const nowSeconds = Date.now() / 1000;

				if (
					!Number.isFinite(timestampSeconds) ||
					Math.abs(nowSeconds - timestampSeconds) > toleranceSeconds
				) {
					rejected.status(403).json({ message: 'Stale Probo webhook timestamp' });
					return { noWebhookResponse: true };
				}
			}
		}

		return {
			workflowData: [this.helpers.returnJsonArray(body as IDataObject)],
		};
	}
}
