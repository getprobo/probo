// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

const scanSelection = `
	id
	state
	failureReason
	truncated
	pagesScanned
	cookiesFound
	storageFound
	resourcesFound
	startedAt
	completedAt
	createdAt
	updatedAt
	checkpoint {
		visitedTemplates
		queue { url }
		coveredHosts
		pagesScanned
		elapsedMs
		truncated
	}
`;

export const description: INodeProperties[] = [
	{
		displayName: 'Scan ID',
		name: 'scanId',
		type: 'string',
		required: true,
		displayOptions: {
			show: {
				resource: ['cookieBanner'],
				operation: ['getScan'],
			},
		},
		default: '',
		description: 'Cookie banner scan ID',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const scanId = this.getNodeParameter('scanId', itemIndex) as string;

	const query = `
		query GetCookieBannerScan($scanId: ID!) {
			node(id: $scanId) {
				... on CookieBannerScan {
					${scanSelection}
				}
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, { scanId });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
