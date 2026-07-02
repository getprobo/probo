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
		displayName: 'Cookie Banner ID',
		name: 'cookieBannerId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['cookieBanner'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the cookie banner to update',
		required: true,
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['cookieBanner'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The name of the cookie banner',
	},
	{
		displayName: 'Cookie Policy URL',
		name: 'cookiePolicyUrl',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['cookieBanner'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The URL to the cookie policy',
	},
	{
		displayName: 'Consent Expiry Days',
		name: 'consentExpiryDays',
		type: 'number',
		displayOptions: {
			show: {
				resource: ['cookieBanner'],
				operation: ['update'],
			},
		},
		default: 0,
		description: 'Number of days before consent expires (0 to leave unchanged)',
	},
	{
		displayName: 'Default Language',
		name: 'defaultLanguage',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['cookieBanner'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The default language for the cookie banner',
	},
	{
		displayName: 'Additional Fields',
		name: 'additionalFields',
		type: 'collection',
		placeholder: 'Add Field',
		default: {},
		displayOptions: {
			show: {
				resource: ['cookieBanner'],
				operation: ['update'],
			},
		},
		options: [
			{
				displayName: 'Privacy Policy URL',
				name: 'privacyPolicyUrl',
				type: 'string',
				default: '',
				description: 'The URL to the privacy policy. Leave empty to clear the existing value.',
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const cookieBannerId = this.getNodeParameter('cookieBannerId', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex, '') as string;
	const cookiePolicyUrl = this.getNodeParameter('cookiePolicyUrl', itemIndex, '') as string;
	const consentExpiryDays = this.getNodeParameter('consentExpiryDays', itemIndex, 0) as number;
	const defaultLanguage = this.getNodeParameter('defaultLanguage', itemIndex, '') as string;
	const additionalFields = this.getNodeParameter('additionalFields', itemIndex, {}) as {
		privacyPolicyUrl?: string;
	};

	const query = `
		mutation UpdateCookieBanner($input: UpdateCookieBannerInput!) {
			updateCookieBanner(input: $input) {
				cookieBanner {
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
	`;

	const input: Record<string, unknown> = { cookieBannerId };
	if (name) input.name = name;
	if (cookiePolicyUrl) input.cookiePolicyUrl = cookiePolicyUrl;
	if (consentExpiryDays) input.consentExpiryDays = consentExpiryDays;
	if (defaultLanguage) input.defaultLanguage = defaultLanguage;
	if (additionalFields.privacyPolicyUrl !== undefined) {
		input.privacyPolicyUrl = additionalFields.privacyPolicyUrl === '' ? null : additionalFields.privacyPolicyUrl;
	}

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
