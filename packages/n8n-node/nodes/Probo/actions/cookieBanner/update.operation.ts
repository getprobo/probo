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
import { cookieBannerLocator, organizationLocator } from '../../locators';

export const description: INodeProperties[] = [
	organizationLocator({
		resource: ['cookieBanner'],
		operation: ['update'],
		required: false,
		description: 'Organization to search in. Leave this empty when entering an ID.',
	}),
		cookieBannerLocator({
		resource: ['cookieBanner'],
		operation: ['update'],
	}),
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
		displayName: 'Publisher Country Code',
		name: 'publisherCountryCode',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['cookieBanner'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'IAB TCF publisher country of establishment (ISO 3166-1 alpha-2, or AA)',
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
			{
				displayName: 'Resource Reporting Enabled',
				name: 'resourceReportingEnabled',
				type: 'boolean',
				default: true,
				description:
					'Whether the SDK reports detected resources. Only add this field when you intend to change the value; omit it to keep the banner\'s current setting.',
			},
			{
				displayName: 'Force Opt-In',
				name: 'forceOptIn',
				type: 'boolean',
				default: false,
				description:
					'Whether every visitor sees the opt-in banner regardless of regulation. Only add this field when you intend to change the value; omit it to keep the banner\'s current setting.',
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const cookieBannerId = this.getNodeParameter('cookieBannerId', itemIndex, '', { extractValue: true }) as string;
	const name = this.getNodeParameter('name', itemIndex, '') as string;
	const cookiePolicyUrl = this.getNodeParameter('cookiePolicyUrl', itemIndex, '') as string;
	const consentExpiryDays = this.getNodeParameter('consentExpiryDays', itemIndex, 0) as number;
	const defaultLanguage = this.getNodeParameter('defaultLanguage', itemIndex, '') as string;
	const publisherCountryCode = this.getNodeParameter('publisherCountryCode', itemIndex, '') as string;
	const additionalFields = this.getNodeParameter('additionalFields', itemIndex, {}) as {
		privacyPolicyUrl?: string;
		resourceReportingEnabled?: boolean;
		forceOptIn?: boolean;
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
					forceOptIn
					capabilities { resourceReporting corsless }
					defaultLanguage
					publisherCountryCode
					tcfCmpId
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
	if (publisherCountryCode) input.publisherCountryCode = publisherCountryCode;
	if (additionalFields.privacyPolicyUrl !== undefined) {
		input.privacyPolicyUrl = additionalFields.privacyPolicyUrl === '' ? null : additionalFields.privacyPolicyUrl;
	}
	if (additionalFields.resourceReportingEnabled !== undefined) {
		input.capabilities = { resourceReporting: additionalFields.resourceReportingEnabled };
	}
	if (additionalFields.forceOptIn !== undefined) {
		input.forceOptIn = additionalFields.forceOptIn;
	}

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
