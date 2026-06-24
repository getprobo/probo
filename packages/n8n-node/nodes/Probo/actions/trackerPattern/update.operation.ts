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
		displayName: 'Tracker Pattern ID',
		name: 'trackerPatternId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['trackerPattern'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the tracker pattern to update',
		required: true,
	},
	{
		displayName: 'Excluded',
		name: 'excluded',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['trackerPattern'],
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
		description: 'Whether the tracker pattern is excluded from the banner',
	},
	{
		displayName: 'Description',
		name: 'patternDescription',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['trackerPattern'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The description of the tracker pattern',
	},
	{
		displayName: 'Additional Fields',
		name: 'additionalFields',
		type: 'collection',
		placeholder: 'Add Field',
		default: {},
		displayOptions: {
			show: {
				resource: ['trackerPattern'],
				operation: ['update'],
			},
		},
		options: [
			{
				displayName: 'Max Age Seconds',
				name: 'maxAgeSeconds',
				type: 'number',
				typeOptions: {
					minValue: 0,
				},
				default: 0,
				description: 'The maximum age of the cookie in seconds. Set to 0 to clear the existing value.',
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const trackerPatternId = this.getNodeParameter('trackerPatternId', itemIndex) as string;
	const excluded = this.getNodeParameter('excluded', itemIndex, '') as string;
	const patternDescription = this.getNodeParameter('patternDescription', itemIndex, '') as string;
	const additionalFields = this.getNodeParameter('additionalFields', itemIndex, {}) as {
		maxAgeSeconds?: number;
	};

	const query = `
		mutation UpdateTrackerPattern($input: UpdateTrackerPatternInput!) {
			updateTrackerPattern(input: $input) {
				trackerPattern {
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

	const input: Record<string, unknown> = { trackerPatternId };
	if (excluded) input.excluded = excluded === 'true';
	if (patternDescription) input.description = patternDescription;
	if (additionalFields.maxAgeSeconds !== undefined) {
		input.maxAgeSeconds = additionalFields.maxAgeSeconds === 0 ? null : additionalFields.maxAgeSeconds;
	}

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
