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
