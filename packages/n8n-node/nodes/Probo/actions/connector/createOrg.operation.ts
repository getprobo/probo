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
import { NodeOperationError } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Organization ID',
		name: 'organizationId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['connector'],
				operation: ['createOrg'],
			},
		},
		default: '',
		description: 'The ID of the organization',
		required: true,
	},
	{
		displayName: 'Provider',
		name: 'provider',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['connector'],
				operation: ['createOrg'],
			},
		},
		options: [
			{
				name: 'AWS',
				value: 'AWS',
			},
			{
				name: 'Azure',
				value: 'AZURE',
			},
			{
				name: 'GCP',
				value: 'GCP',
			},
		],
		default: 'AWS',
		description: 'Cloud provider for the organization connector',
		required: true,
	},
	{
		displayName: 'AWS Role ARN',
		name: 'awsRoleArn',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['connector'],
				operation: ['createOrg'],
				provider: ['AWS'],
			},
		},
		default: '',
		description: 'IAM role ARN, including partition, account, and role name',
		required: true,
	},
	{
		displayName: 'AWS Member Role Name',
		name: 'awsMemberRoleName',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['connector'],
				operation: ['createOrg'],
				provider: ['AWS'],
			},
		},
		default: '',
		description: 'IAM role assumed in member accounts. Empty uses ProboAudit.',
	},
	{
		displayName: 'GCP Workload Identity Provider',
		name: 'gcpWorkloadIdentityProvider',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['connector'],
				operation: ['createOrg'],
				provider: ['GCP'],
			},
		},
		default: '',
		description: 'Workload identity provider resource',
		required: true,
	},
	{
		displayName: 'GCP Service Account Email',
		name: 'gcpServiceAccountEmail',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['connector'],
				operation: ['createOrg'],
				provider: ['GCP'],
			},
		},
		default: '',
		description: 'Service account email to impersonate',
		required: true,
	},
	{
		displayName: 'GCP Parent',
		name: 'gcpParent',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['connector'],
				operation: ['createOrg'],
				provider: ['GCP'],
			},
		},
		default: '',
		description: 'Cloud Asset parent, for example organizations/123 or folders/456',
	},
	{
		displayName: 'Azure Tenant ID',
		name: 'azureTenantId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['connector'],
				operation: ['createOrg'],
				provider: ['AZURE'],
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
				operation: ['createOrg'],
				provider: ['AZURE'],
			},
		},
		default: '',
		description: 'Entra application (client) ID',
		required: true,
	},
	{
		displayName: 'Azure Environment',
		name: 'azureEnvironment',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['connector'],
				operation: ['createOrg'],
				provider: ['AZURE'],
			},
		},
		options: [
			{
				name: 'Public (GCC Uses This)',
				value: 'AZURE_PUBLIC',
			},
			{
				name: 'Government (GCC High)',
				value: 'AZURE_GOVERNMENT',
			},
			{
				name: 'Government DoD',
				value: 'AZURE_GOVERNMENT_DOD',
			},
			{
				name: 'China',
				value: 'AZURE_CHINA',
			},
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
	const provider = this.getNodeParameter('provider', itemIndex) as string;

	const input: Record<string, string> = {
		organizationId,
		provider,
	};

	switch (provider) {
		case 'AWS':
			input.awsRoleArn = this.getNodeParameter('awsRoleArn', itemIndex) as string;
			{
				const memberRoleName = this.getNodeParameter(
					'awsMemberRoleName',
					itemIndex,
					'',
				) as string;
				if (memberRoleName !== '') {
					input.awsMemberRoleName = memberRoleName;
				}
			}
			break;
		case 'GCP':
			input.gcpWorkloadIdentityProvider = this.getNodeParameter(
				'gcpWorkloadIdentityProvider',
				itemIndex,
			) as string;
			input.gcpServiceAccountEmail = this.getNodeParameter(
				'gcpServiceAccountEmail',
				itemIndex,
			) as string;
			{
				const parent = this.getNodeParameter('gcpParent', itemIndex, '') as string;
				if (parent !== '') {
					input.gcpParent = parent;
				}
			}
			break;
		case 'AZURE':
			input.azureTenantId = this.getNodeParameter('azureTenantId', itemIndex) as string;
			input.azureClientId = this.getNodeParameter('azureClientId', itemIndex) as string;
			input.azureEnvironment = this.getNodeParameter(
				'azureEnvironment',
				itemIndex,
			) as string;
			break;
		default:
			throw new NodeOperationError(this.getNode(), `Unsupported provider ${provider}`, {
				itemIndex,
			});
	}

	const query = `
		mutation CreateOrganizationConnector($input: CreateOrganizationConnectorInput!) {
			createOrganizationConnector(input: $input) {
				connector {
					id
					provider
					protocol
					createdAt
				}
				discoveredAccounts {
					externalAccountId
					name
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
