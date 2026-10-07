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

import type { IExecuteFunctions, INodeExecutionData, INodeProperties } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

const sourceDescription =
	'Connect the provider first. Create the connector with Create API Key, Create Client Credentials, or Create AWS, GCP, or Azure Workload Identity Federation, then pass that connector id here. This operation does not collect credentials. A CSV source is the path with no provider.';

export const description: INodeProperties[] = [
	{
		displayName: 'Organization ID',
		name: 'organizationId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['accessReviewSource'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The ID of the organization',
		required: true,
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['accessReviewSource'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The name of the access source',
		required: true,
	},
	{
		displayName: 'Source Type',
		name: 'sourceType',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['accessReviewSource'],
				operation: ['create'],
			},
		},
		options: [
			{
				name: 'Connector',
				value: 'connector',
				description: sourceDescription,
			},
			{
				name: 'CSV',
				value: 'csv',
				description: 'A CSV source has no provider',
			},
		],
		default: 'connector',
		description: sourceDescription,
	},
	{
		displayName: 'Connector ID',
		name: 'connectorId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['accessReviewSource'],
				operation: ['create'],
				sourceType: ['connector'],
			},
		},
		default: '',
		description: 'ID of a connector that already exists',
		required: true,
	},
	{
		displayName: 'Connector Account ID',
		name: 'connectorAccountId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['accessReviewSource'],
				operation: ['create'],
				sourceType: ['connector'],
			},
		},
		default: '',
		description:
			'Account on the connector. Leave empty to use the account this credential is, or its only account.',
	},
	{
		displayName: 'CSV Data',
		name: 'csvData',
		type: 'string',
		typeOptions: {
			rows: 4,
		},
		displayOptions: {
			show: {
				resource: ['accessReviewSource'],
				operation: ['create'],
				sourceType: ['csv'],
			},
		},
		default: '',
		description: 'CSV access data. This path has no provider.',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex) as string;
	const sourceType = this.getNodeParameter('sourceType', itemIndex) as string;

	const source: Record<string, string> = { name };

	if (sourceType === 'csv') {
		source.csvData = this.getNodeParameter('csvData', itemIndex) as string;
	} else {
		source.connectorId = this.getNodeParameter('connectorId', itemIndex) as string;
		const connectorAccountId = this.getNodeParameter('connectorAccountId', itemIndex, '') as string;
		if (connectorAccountId !== '') {
			source.connectorAccountId = connectorAccountId;
		}
	}

	const query = `
		mutation CreateAccessReviewSources($input: CreateAccessReviewSourcesInput!) {
			createAccessReviewSources(input: $input) {
				results {
					created
					accessReviewSourceEdge {
						node {
							id
							name
							connectorId
							createdAt
						}
					}
				}
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, {
		input: {
			organizationId,
			sources: [source],
		},
	});

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
