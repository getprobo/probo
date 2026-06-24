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
				resource: ['trackerPattern'],
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
				resource: ['trackerPattern'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The tracker name pattern to match',
		required: true,
	},
	{
		displayName: 'Match Type',
		name: 'matchType',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['trackerPattern'],
				operation: ['create'],
			},
		},
		options: [
			{
				name: 'Exact',
				value: 'EXACT',
			},
			{
				name: 'Glob',
				value: 'GLOB',
			},
		],
		default: 'EXACT',
		description: 'How the pattern should be matched against tracker names. GLOB uses * as wildcard.',
		required: true,
	},
	{
		displayName: 'Display Name',
		name: 'displayName',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['trackerPattern'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The display name for the tracker pattern',
		required: true,
	},
	{
		displayName: 'Description',
		name: 'description',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['trackerPattern'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The description of the tracker pattern',
		required: true,
	},
	{
		displayName: 'Max Age Seconds',
		name: 'maxAgeSeconds',
		type: 'number',
		displayOptions: {
			show: {
				resource: ['trackerPattern'],
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
		mutation CreateTrackerPattern($input: CreateTrackerPatternInput!) {
			createTrackerPattern(input: $input) {
				trackerPatternEdge {
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
