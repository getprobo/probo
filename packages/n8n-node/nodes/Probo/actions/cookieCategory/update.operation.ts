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
		input.gcmConsentTypes = gcmConsentTypes
			.split(',')
			.map((s) => s.trim())
			.filter((s) => s.length > 0);
	}
	if (posthogConsent) input.posthogConsent = posthogConsent === 'true';

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
