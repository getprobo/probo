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

import type { INodeProperties, IExecuteFunctions, INodeExecutionData, IDataObject } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Organization ID',
		name: 'organizationId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['connector'],
				operation: ['enableAccounts'],
			},
		},
		default: '',
		description: 'The ID of the organization',
		required: true,
	},
	{
		displayName: 'Connector ID',
		name: 'connectorId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['connector'],
				operation: ['enableAccounts'],
			},
		},
		default: '',
		description: 'The ID of the connector',
		required: true,
	},
	{
		displayName: 'Accounts',
		name: 'accounts',
		type: 'fixedCollection',
		typeOptions: {
			multipleValues: true,
		},
		displayOptions: {
			show: {
				resource: ['connector'],
				operation: ['enableAccounts'],
			},
		},
		default: {},
		description: 'The vendor accounts to record under this connector',
		options: [
			{
				name: 'account',
				displayName: 'Account',
				values: [
					{
						displayName: 'External Account ID',
						name: 'externalAccountId',
						type: 'string',
						default: '',
						description: 'The identifier the vendor uses for this account',
						required: true,
					},
					{
						displayName: 'Name',
						name: 'name',
						type: 'string',
						default: '',
						description: 'Display name, defaulting to the identifier',
					},
				],
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const connectorId = this.getNodeParameter('connectorId', itemIndex) as string;
	const accountsInput = this.getNodeParameter('accounts', itemIndex, {}) as IDataObject;
	const accountEntries = (accountsInput.account as IDataObject[] | undefined) ?? [];

	const accounts = accountEntries.map((entry) => {
		const account: IDataObject = { externalAccountId: entry.externalAccountId };

		if (entry.name) {
			account.name = entry.name;
		}

		return account;
	});

	const query = `
		mutation EnableConnectorAccounts($input: EnableConnectorAccountsInput!) {
			enableConnectorAccounts(input: $input) {
				connectorAccountEdges {
					node {
						id
						externalAccountId
						name
					}
				}
			}
		}
	`;

	const response = await proboApiRequest.call(this, query, {
		input: { organizationId, connectorId, accounts },
	});
	const data = response?.data as IDataObject | undefined;
	const payload = data?.enableConnectorAccounts as IDataObject | undefined;

	return {
		json: { connectorAccountEdges: payload?.connectorAccountEdges ?? [] },
		pairedItem: { item: itemIndex },
	};
}
