// Copyright (c) 2025 Probo Inc <hello@probo.com>.
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
		displayName: 'ThirdParty Service ID',
		name: 'thirdPartyServiceId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['thirdParty'],
				operation: ['getService'],
			},
		},
		default: '',
		description: 'The ID of the thirdParty service',
		required: true,
	},
	{
		displayName: 'Options',
		name: 'options',
		type: 'collection',
		placeholder: 'Add Option',
		default: {},
		displayOptions: {
			show: {
				resource: ['thirdParty'],
				operation: ['getService'],
			},
		},
		options: [
			{
				displayName: 'Include ThirdParty',
				name: 'includeThirdParty',
				type: 'boolean',
				default: false,
				description: 'Whether to include thirdParty in the response',
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const thirdPartyServiceId = this.getNodeParameter('thirdPartyServiceId', itemIndex) as string;
	const options = this.getNodeParameter('options', itemIndex, {}) as {
		includeThirdParty?: boolean;
	};

	const thirdPartyFragment = options.includeThirdParty
		? `thirdParty {
			id
			name
		}`
		: '';

	const query = `
		query GetThirdPartyService($thirdPartyServiceId: ID!) {
			node(id: $thirdPartyServiceId) {
				... on ThirdPartyService {
					id
					name
					description
					${thirdPartyFragment}
					createdAt
					updatedAt
				}
			}
		}
	`;

	const variables = {
		thirdPartyServiceId,
	};

	const responseData = await proboApiRequest.call(this, query, variables);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
