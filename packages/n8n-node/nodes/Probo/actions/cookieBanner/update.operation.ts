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
		displayName: 'Consent Mode',
		name: 'consentMode',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['cookieBanner'],
				operation: ['update'],
			},
		},
		options: [
			{
				name: '(Unchanged)',
				value: '',
			},
			{
				name: 'Opt In',
				value: 'OPT_IN',
			},
			{
				name: 'Opt Out',
				value: 'OPT_OUT',
			},
		],
		default: '',
		description: 'The consent mode for the cookie banner',
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
	const consentMode = this.getNodeParameter('consentMode', itemIndex, '') as string;
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
					consentMode
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
	if (consentMode) input.consentMode = consentMode;
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
