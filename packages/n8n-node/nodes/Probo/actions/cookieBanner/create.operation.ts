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
		typeOptions: {
			minValue: 1,
		},
		default: 365,
		description: 'Number of days before consent expires',
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
	};
	if (privacyPolicyUrl) input.privacyPolicyUrl = privacyPolicyUrl;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
