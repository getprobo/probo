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
		displayName: 'Cookie Category ID',
		name: 'cookieCategoryId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['cookieCategory'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the cookie category to update',
		required: true,
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['cookieCategory'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The name of the cookie category',
	},
	{
		displayName: 'Slug',
		name: 'slug',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['cookieCategory'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The slug of the cookie category',
	},
	{
		displayName: 'Description',
		name: 'categoryDescription',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['cookieCategory'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The description of the cookie category',
	},
	{
		displayName: 'GCM Consent Types',
		name: 'gcmConsentTypes',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['cookieCategory'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'Comma-separated list of GCM consent types',
	},
	{
		displayName: 'PostHog Consent',
		name: 'posthogConsent',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['cookieCategory'],
				operation: ['update'],
			},
		},
		options: [
			{
				name: '(Unchanged)',
				value: '',
			},
			{
				name: 'True',
				value: 'true',
			},
			{
				name: 'False',
				value: 'false',
			},
		],
		default: '',
		description: 'Whether this category maps to PostHog consent',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const cookieCategoryId = this.getNodeParameter('cookieCategoryId', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex, '') as string;
	const slug = this.getNodeParameter('slug', itemIndex, '') as string;
	const categoryDescription = this.getNodeParameter('categoryDescription', itemIndex, '') as string;
	const gcmConsentTypes = this.getNodeParameter('gcmConsentTypes', itemIndex, '') as string;
	const posthogConsent = this.getNodeParameter('posthogConsent', itemIndex, '') as string;

	const query = `
		mutation UpdateCookieCategory($input: UpdateCookieCategoryInput!) {
			updateCookieCategory(input: $input) {
				cookieCategory {
					id
					name
					slug
					description
					kind
					rank
					gcmConsentTypes
					posthogConsent
					createdAt
					updatedAt
				}
				cookieBanner {
					id
					name
				}
			}
		}
	`;

	const input: Record<string, unknown> = { cookieCategoryId };
	if (name) input.name = name;
	if (slug) input.slug = slug;
	if (categoryDescription) input.description = categoryDescription;
	if (gcmConsentTypes) {
		input.gcmConsentTypes = gcmConsentTypes.split(',').map((s) => s.trim());
	}
	if (posthogConsent) input.posthogConsent = posthogConsent === 'true';

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
