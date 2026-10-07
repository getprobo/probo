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

const operation = 'createGcpWorkloadIdentity';

export const description: INodeProperties[] = [
	organizationIdField(operation),
	connectorNameField(operation),
	{
		displayName: 'GCP Workload Identity Provider',
		name: 'gcpWorkloadIdentityProvider',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['connector'],
				operation: [operation],
			},
		},
		default: '',
		description: 'Workload identity provider resource, including the S3NS IAM host when applicable',
		required: true,
	},
	{
		displayName: 'GCP Service Account Email',
		name: 'gcpServiceAccountEmail',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['connector'],
				operation: [operation],
			},
		},
		default: '',
		description: 'Service account email to impersonate, including the universe-specific suffix',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex) as string;
	const gcpWorkloadIdentityProvider = this.getNodeParameter(
		'gcpWorkloadIdentityProvider',
		itemIndex,
	) as string;
	const gcpServiceAccountEmail = this.getNodeParameter(
		'gcpServiceAccountEmail',
		itemIndex,
	) as string;

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
			provider: 'GCP',
			gcpWorkloadIdentityProvider,
			gcpServiceAccountEmail,
		},
	});

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
