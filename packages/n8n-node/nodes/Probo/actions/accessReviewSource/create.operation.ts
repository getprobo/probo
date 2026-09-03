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
import { NodeOperationError } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

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
		displayName: 'Provider',
		name: 'provider',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['accessReviewSource'],
				operation: ['create'],
			},
		},
		options: [
			{
				name: 'AWS',
				value: 'AWS',
			},
			{
				name: 'GCP',
				value: 'GCP',
			},
		],
		default: 'AWS',
		description: 'Cloud provider for the workload-identity connector',
		required: true,
	},
	{
		displayName: 'AWS Role ARN',
		name: 'awsRoleArn',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['accessReviewSource'],
				operation: ['create'],
				provider: ['AWS'],
			},
		},
		default: '',
		description: 'IAM role ARN, including partition, account, and role name',
		required: true,
	},
	{
		displayName: 'GCP Workload Identity Provider',
		name: 'gcpWorkloadIdentityProvider',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['accessReviewSource'],
				operation: ['create'],
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
				resource: ['accessReviewSource'],
				operation: ['create'],
				provider: ['GCP'],
			},
		},
		default: '',
		description: 'Service account email to impersonate',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex) as string;
	const provider = this.getNodeParameter('provider', itemIndex) as string;

	const createConnectorQuery = `
		mutation CreateWorkloadIdentityConnector($input: CreateWorkloadIdentityConnectorInput!) {
			createWorkloadIdentityConnector(input: $input) {
				connector {
					id
					provider
					protocol
					connectionStatus
					createdAt
				}
			}
		}
	`;

	const connectorInput: Record<string, string> = {
		organizationId,
		provider,
	};

	if (provider === 'GCP') {
		connectorInput.gcpWorkloadIdentityProvider = this.getNodeParameter(
			'gcpWorkloadIdentityProvider',
			itemIndex,
		) as string;
		connectorInput.gcpServiceAccountEmail = this.getNodeParameter(
			'gcpServiceAccountEmail',
			itemIndex,
		) as string;
	} else {
		connectorInput.awsRoleArn = this.getNodeParameter('awsRoleArn', itemIndex) as string;
	}

	const connectorResponse = await proboApiRequest.call(this, createConnectorQuery, {
		input: connectorInput,
	});

	const connectorPayload = (connectorResponse.data as IDataObject | undefined)
		?.createWorkloadIdentityConnector as IDataObject | undefined;
	const connector = connectorPayload?.connector as IDataObject | undefined;
	const connectorId = connector?.id as string | undefined;
	if (!connectorId) {
		throw new NodeOperationError(this.getNode(), 'Workload identity connector was not created');
	}

	const connectionStatus = connector?.connectionStatus as string | undefined;

	const createSourceQuery = `
		mutation CreateAccessReviewSource($input: CreateAccessReviewSourceInput!) {
			createAccessReviewSource(input: $input) {
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
	`;

	const deleteConnectorQuery = `
		mutation DeleteConnector($input: DeleteConnectorInput!) {
			deleteConnector(input: $input) {
				deletedConnectorId
			}
		}
	`;

	try {
		if (connectionStatus !== 'CONNECTED') {
			throw new NodeOperationError(
				this.getNode(),
				`Connector is ${connectionStatus ?? 'in an unknown state'}`,
				{ itemIndex },
			);
		}

		const sourceResponse = await proboApiRequest.call(this, createSourceQuery, {
			input: {
				organizationId,
				name,
				connectorId,
			},
		});

		return {
			json: sourceResponse,
			pairedItem: { item: itemIndex },
		};
	} catch (error) {
		try {
			await proboApiRequest.call(this, deleteConnectorQuery, {
				input: { connectorId },
			});
		} catch (cleanupError) {
			const cleanupMessage =
				cleanupError instanceof Error ? cleanupError.message : String(cleanupError);

			throw new NodeOperationError(
				this.getNode(),
				error instanceof Error ? error : new Error(String(error)),
				{
					itemIndex,
					description: `Cannot delete leftover connector ${connectorId}: ${cleanupMessage}`,
				},
			);
		}

		throw new NodeOperationError(this.getNode(), error as Error, { itemIndex });
	}
}
