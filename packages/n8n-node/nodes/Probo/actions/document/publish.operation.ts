// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
		displayName: 'Document ID',
		name: 'documentId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['document'],
				operation: ['publish'],
			},
		},
		default: '',
		description: 'The ID of the document',
		required: true,
	},
	{
		displayName: 'Minor',
		name: 'minor',
		type: 'boolean',
		displayOptions: {
			show: {
				resource: ['document'],
				operation: ['publish'],
			},
		},
		default: false,
		description: 'Whether to publish as a minor version. Approvers are ignored when set.',
	},
	{
		displayName: 'Approver IDs',
		name: 'approverIds',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['document'],
				operation: ['publish'],
				minor: [false],
			},
		},
		default: '',
		description: 'Comma-separated list of approver profile IDs. Provide IDs to request approval; leave empty to publish the major version directly without approval.',
	},
	{
		displayName: 'Changelog',
		name: 'changelog',
		type: 'string',
		typeOptions: {
			rows: 4,
		},
		displayOptions: {
			show: {
				resource: ['document'],
				operation: ['publish'],
			},
		},
		default: '',
		description: 'The changelog for this version',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const documentId = this.getNodeParameter('documentId', itemIndex) as string;
	const minor = this.getNodeParameter('minor', itemIndex, false) as boolean;
	const approverIdsRaw = this.getNodeParameter('approverIds', itemIndex, '') as string;
	const changelog = this.getNodeParameter('changelog', itemIndex) as string;

	const query = `
		mutation PublishDocument($input: PublishDocumentInput!) {
			publishDocument(input: $input) {
				document {
					id
					status
					compliancePortalVisibility
					currentPublishedMajor
					currentPublishedMinor
					createdAt
					updatedAt
				}
				documentVersion {
					id
					title
					major
					minor
					status
					content
					changelog
					classification
					documentType
					publishedAt
					createdAt
					updatedAt
				}
				approvalQuorum {
					id
					status
					createdAt
					updatedAt
				}
			}
		}
	`;

	const input: Record<string, unknown> = { documentId, minor, changelog };
	if (!minor) {
		// A major publish must set approverIds explicitly: an empty list publishes
		// directly without approval, a non-empty list requests approval. It must be
		// omitted for a minor publish, which ignores approvers.
		input.approverIds = approverIdsRaw
			.split(',')
			.map(id => id.trim())
			.filter(Boolean);
	}

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
