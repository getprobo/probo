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
				resource: ['cookieBanner'],
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
				resource: ['cookieBanner'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The name of the cookie banner',
		required: true,
	},
	{
		displayName: 'Origin',
		name: 'origin',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['cookieBanner'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The origin URL for the cookie banner',
		required: true,
	},
	{
		displayName: 'Cookie Policy URL',
		name: 'cookiePolicyUrl',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['cookieBanner'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The URL to the cookie policy',
		required: true,
	},
	{
		displayName: 'Consent Expiry Days',
		name: 'consentExpiryDays',
		type: 'number',
		displayOptions: {
			show: {
				resource: ['cookieBanner'],
				operation: ['create'],
			},
		},
		default: 365,
		description: 'Number of days before consent expires',
		required: true,
	},
	{
		displayName: 'Consent Mode',
		name: 'consentMode',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['cookieBanner'],
				operation: ['create'],
			},
		},
		options: [
			{
				name: 'Opt In',
				value: 'OPT_IN',
			},
			{
				name: 'Opt Out',
				value: 'OPT_OUT',
			},
		],
		default: 'OPT_IN',
		description: 'The consent mode for the cookie banner',
		required: true,
	},
	{
		displayName: 'Privacy Policy URL',
		name: 'privacyPolicyUrl',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['cookieBanner'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The URL to the privacy policy',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex) as string;
	const origin = this.getNodeParameter('origin', itemIndex) as string;
	const cookiePolicyUrl = this.getNodeParameter('cookiePolicyUrl', itemIndex) as string;
	const consentExpiryDays = this.getNodeParameter('consentExpiryDays', itemIndex) as number;
	const consentMode = this.getNodeParameter('consentMode', itemIndex) as string;
	const privacyPolicyUrl = this.getNodeParameter('privacyPolicyUrl', itemIndex, '') as string;

	const query = `
		mutation CreateCookieBanner($input: CreateCookieBannerInput!) {
			createCookieBanner(input: $input) {
				cookieBannerEdge {
					node {
						id
						name
						origin
						state
						privacyPolicyUrl
						cookiePolicyUrl
						consentExpiryDays
						consentMode
						showBranding
						defaultLanguage
						createdAt
						updatedAt
					}
				}
			}
		}
	`;

	const input: Record<string, unknown> = {
		organizationId,
		name,
		origin,
		cookiePolicyUrl,
		consentExpiryDays,
		consentMode,
	};
	if (privacyPolicyUrl) input.privacyPolicyUrl = privacyPolicyUrl;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
