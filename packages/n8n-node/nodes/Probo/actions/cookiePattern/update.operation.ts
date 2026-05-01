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
		displayName: 'Cookie Pattern ID',
		name: 'cookiePatternId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['cookiePattern'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the cookie pattern to update',
		required: true,
	},
	{
		displayName: 'Display Name',
		name: 'displayName',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['cookiePattern'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The display name for the cookie pattern',
	},
	{
		displayName: 'Max Age Seconds',
		name: 'maxAgeSeconds',
		type: 'number',
		displayOptions: {
			show: {
				resource: ['cookiePattern'],
				operation: ['update'],
			},
		},
		default: 0,
		description: 'The maximum age of the cookie in seconds (0 to clear)',
	},
	{
		displayName: 'Excluded',
		name: 'excluded',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['cookiePattern'],
				operation: ['update'],
			},
		},
		options: [
			{
				name: '(Unchanged)',
				value: '',
			},
			{
				name: 'False',
				value: 'false',
			},
			{
				name: 'True',
				value: 'true',
			},
		],
		default: '',
		description: 'Whether the cookie pattern is excluded from the banner',
	},
	{
		displayName: 'Description',
		name: 'patternDescription',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['cookiePattern'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The description of the cookie pattern',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const cookiePatternId = this.getNodeParameter('cookiePatternId', itemIndex) as string;
	const displayName = this.getNodeParameter('displayName', itemIndex, '') as string;
	const maxAgeSeconds = this.getNodeParameter('maxAgeSeconds', itemIndex, 0) as number;
	const excluded = this.getNodeParameter('excluded', itemIndex, '') as string;
	const patternDescription = this.getNodeParameter('patternDescription', itemIndex, '') as string;

	const query = `
		mutation UpdateCookiePattern($input: UpdateCookiePatternInput!) {
			updateCookiePattern(input: $input) {
				cookiePattern {
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
				cookieBanner {
					id
					name
				}
			}
		}
	`;

	const input: Record<string, unknown> = { cookiePatternId };
	if (displayName) input.displayName = displayName;
	if (maxAgeSeconds !== undefined) {
		input.maxAgeSeconds = maxAgeSeconds === 0 ? null : maxAgeSeconds;
	}
	if (excluded) input.excluded = excluded === 'true';
	if (patternDescription) input.description = patternDescription;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
