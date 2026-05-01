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
				resource: ['cookiePattern'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The ID of the cookie category',
		required: true,
	},
	{
		displayName: 'Pattern',
		name: 'pattern',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['cookiePattern'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The cookie name pattern to match',
		required: true,
	},
	{
		displayName: 'Match Type',
		name: 'matchType',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['cookiePattern'],
				operation: ['create'],
			},
		},
		options: [
			{
				name: 'Exact',
				value: 'EXACT',
			},
			{
				name: 'Prefix',
				value: 'PREFIX',
			},
		],
		default: 'EXACT',
		description: 'How the pattern should be matched against cookie names',
		required: true,
	},
	{
		displayName: 'Display Name',
		name: 'displayName',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['cookiePattern'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The display name for the cookie pattern',
		required: true,
	},
	{
		displayName: 'Description',
		name: 'description',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['cookiePattern'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The description of the cookie pattern',
		required: true,
	},
	{
		displayName: 'Max Age Seconds',
		name: 'maxAgeSeconds',
		type: 'number',
		displayOptions: {
			show: {
				resource: ['cookiePattern'],
				operation: ['create'],
			},
		},
		default: 0,
		description: 'The maximum age of the cookie in seconds (0 to omit)',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const cookieCategoryId = this.getNodeParameter('cookieCategoryId', itemIndex) as string;
	const pattern = this.getNodeParameter('pattern', itemIndex) as string;
	const matchType = this.getNodeParameter('matchType', itemIndex) as string;
	const displayName = this.getNodeParameter('displayName', itemIndex) as string;
	const description = this.getNodeParameter('description', itemIndex) as string;
	const maxAgeSeconds = this.getNodeParameter('maxAgeSeconds', itemIndex, 0) as number;

	const query = `
		mutation CreateCookiePattern($input: CreateCookiePatternInput!) {
			createCookiePattern(input: $input) {
				cookiePatternEdge {
					node {
						id
						pattern
						matchType
						displayName
						maxAgeSeconds
						description
						source
						excluded
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

	const input: Record<string, unknown> = {
		cookieCategoryId,
		pattern,
		matchType,
		displayName,
		description,
	};
	if (maxAgeSeconds) input.maxAgeSeconds = maxAgeSeconds;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
