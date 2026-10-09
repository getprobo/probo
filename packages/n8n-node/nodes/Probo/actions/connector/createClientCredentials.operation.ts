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
import { connectorNameField, connectorResultFields, organizationIdField } from './fields';

const operation = 'createClientCredentials';

export const description: INodeProperties[] = [
	organizationIdField(operation),
	connectorNameField(operation),
	{
		displayName: 'Provider Name or ID',
		name: 'provider',
		type: 'options',
		typeOptions: {
			loadOptionsMethod: 'getClientCredentialsConnectorProviders',
		},
		displayOptions: {
			show: {
				resource: ['connector'],
				operation: [operation],
			},
		},
		default: '',
		description:
			'Choose from the list, or specify an ID using an <a href="https://docs.n8n.io/code/expressions/">expression</a>',
		required: true,
	},
	{
		displayName: 'Client ID',
		name: 'clientId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['connector'],
				operation: [operation],
			},
		},
		default: '',
		required: true,
	},
	{
		displayName: 'Client Secret',
		name: 'clientSecret',
		type: 'string',
		typeOptions: {
			password: true,
		},
		displayOptions: {
			show: {
				resource: ['connector'],
				operation: [operation],
			},
		},
		default: '',
		required: true,
	},
	{
		displayName: 'Token Endpoint',
		name: 'tokenUrl',
		type: 'string',
		typeOptions: {
			password: true,
		},
		displayOptions: {
			show: {
				resource: ['connector'],
				operation: [operation],
			},
		},
		default: '',
		description: 'Leave empty when the provider pins a token URL. The server ignores a value in that case.',
	},
	{
		displayName: 'Scope',
		name: 'scope',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['connector'],
				operation: [operation],
			},
		},
		default: '',
		description: 'OAuth scope. The provider\'s own scopes win when it declares them.',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex) as string;
	const provider = this.getNodeParameter('provider', itemIndex) as string;
	const clientId = this.getNodeParameter('clientId', itemIndex) as string;
	const clientSecret = this.getNodeParameter('clientSecret', itemIndex) as string;
	const tokenUrl = this.getNodeParameter('tokenUrl', itemIndex, '') as string;
	const scope = this.getNodeParameter('scope', itemIndex, '') as string;

	const input: Record<string, string> = {
		organizationId,
		name,
		provider,
		clientId,
		clientSecret,
	};

	if (tokenUrl !== '') {
		input.tokenUrl = tokenUrl;
	}
	if (scope !== '') {
		input.scope = scope;
	}

	const query = `
		mutation CreateClientCredentialsConnector($input: CreateClientCredentialsConnectorInput!) {
			createClientCredentialsConnector(input: $input) {
				connector {
					${connectorResultFields}
				}
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
