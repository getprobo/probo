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
import { proboConnectApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Service Account ID',
		name: 'serviceAccountId',
		type: 'string',
		displayOptions: {
			show: { resource: ['serviceAccount'], operation: ['createCredential'] },
		},
		default: '',
		description: 'The ID of the service account',
		required: true,
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: { resource: ['serviceAccount'], operation: ['createCredential'] },
		},
		default: '',
		description: 'The credential name',
		required: true,
	},
	{
		displayName: 'Scopes',
		name: 'scopes',
		type: 'string',
		displayOptions: {
			show: { resource: ['serviceAccount'], operation: ['createCredential'] },
		},
		default: '',
		description: 'Comma-separated OAuth2 scopes',
		required: true,
	},
	{
		displayName: 'Expires At',
		name: 'expiresAt',
		type: 'dateTime',
		displayOptions: {
			show: { resource: ['serviceAccount'], operation: ['createCredential'] },
		},
		default: '',
		description: 'The credential expiration timestamp',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const serviceAccountId = this.getNodeParameter('serviceAccountId', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex) as string;
	const scopes = (this.getNodeParameter('scopes', itemIndex) as string)
		.split(',')
		.map((scope) => scope.trim())
		.filter(Boolean);
	const expiresAt = this.getNodeParameter('expiresAt', itemIndex) as string;
	const query = `
		mutation CreateServiceAccountCredential($input: CreateServiceAccountCredentialInput!) {
			createServiceAccountCredential(input: $input) {
				serviceAccountCredentialEdge {
					node {
						id
						serviceAccountId
						name
						scopes
						expiresAt
						lastUsedAt
						revokedAt
						createdAt
					}
				}
				token
			}
		}
	`;

	const responseData = await proboConnectApiRequest.call(this, query, {
		input: { serviceAccountId, name, scopes, expiresAt },
	});

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
