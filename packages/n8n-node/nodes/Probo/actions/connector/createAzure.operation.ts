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

const operation = 'createAzureWorkloadIdentity';

export const description: INodeProperties[] = [
	organizationIdField(operation),
	connectorNameField(operation),
	{
		displayName: 'Azure Tenant ID',
		name: 'azureTenantId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['connector'],
				operation: [operation],
			},
		},
		default: '',
		description: 'Entra directory (tenant) ID',
		required: true,
	},
	{
		displayName: 'Azure Client ID',
		name: 'azureClientId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['connector'],
				operation: [operation],
			},
		},
		default: '',
		description: 'Entra application (client) ID',
		required: true,
	},
	{
		displayName: 'Azure Subscription ID',
		name: 'azureSubscriptionId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['connector'],
				operation: [operation],
			},
		},
		default: '',
		description: 'Subscription this connector audits',
		required: true,
	},
	{
		displayName: 'Azure Environment',
		name: 'azureEnvironment',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['connector'],
				operation: [operation],
			},
		},
		options: [
			{ name: 'Public (GCC Uses This)', value: 'AZURE_PUBLIC' },
			{ name: 'Government (GCC High)', value: 'AZURE_GOVERNMENT' },
			{ name: 'Government DoD', value: 'AZURE_GOVERNMENT_DOD' },
			{ name: 'China', value: 'AZURE_CHINA' },
		],
		default: 'AZURE_PUBLIC',
		description: 'Azure cloud environment. GCC uses Public. Only GCC High and DoD use Government.',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex) as string;
	const azureTenantId = this.getNodeParameter('azureTenantId', itemIndex) as string;
	const azureClientId = this.getNodeParameter('azureClientId', itemIndex) as string;
	const azureSubscriptionId = this.getNodeParameter('azureSubscriptionId', itemIndex) as string;
	const azureEnvironment = this.getNodeParameter('azureEnvironment', itemIndex) as string;

	const query = `
		mutation CreateWorkloadIdentityConnector($input: CreateWorkloadIdentityConnectorInput!) {
			createWorkloadIdentityConnector(input: $input) {
				connector {
					${connectorResultFields}
				}
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, {
		input: {
			organizationId,
			name,
			provider: 'AZURE',
			azureTenantId,
			azureClientId,
			azureSubscriptionId,
			azureEnvironment,
		},
	});

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
