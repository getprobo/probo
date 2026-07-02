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
				resource: ['cookieCategory'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The ID of the cookie banner',
		required: true,
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['cookieCategory'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The name of the cookie category',
		required: true,
	},
	{
		displayName: 'Slug',
		name: 'slug',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['cookieCategory'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The slug of the cookie category',
		required: true,
	},
	{
		displayName: 'Description',
		name: 'description',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['cookieCategory'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The description of the cookie category',
		required: true,
	},
	{
		displayName: 'Rank',
		name: 'rank',
		type: 'number',
		displayOptions: {
			show: {
				resource: ['cookieCategory'],
				operation: ['create'],
			},
		},
		default: 0,
		description: 'The display order rank of the cookie category',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const cookieBannerId = this.getNodeParameter('cookieBannerId', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex) as string;
	const slug = this.getNodeParameter('slug', itemIndex) as string;
	const description = this.getNodeParameter('description', itemIndex) as string;
	const rank = this.getNodeParameter('rank', itemIndex) as number;

	const query = `
		mutation CreateCookieCategory($input: CreateCookieCategoryInput!) {
			createCookieCategory(input: $input) {
				cookieCategoryEdge {
					node {
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
				}
				cookieBanner {
					id
					name
				}
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, {
		input: { cookieBannerId, name, slug, description, rank },
	});

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
