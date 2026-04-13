// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the document to update',
		required: true,
	},
	{
		displayName: 'Update Fields',
		name: 'updateFields',
		type: 'collection',
		placeholder: 'Add Field',
		default: {},
		displayOptions: {
			show: {
				resource: ['document'],
				operation: ['update'],
			},
		},
		options: [
			{
				displayName: 'Trust Center Visibility',
				name: 'trustCenterVisibility',
				type: 'options',
				options: [
					{ name: 'None', value: 'NONE' },
					{ name: 'Private', value: 'PRIVATE' },
					{ name: 'Public', value: 'PUBLIC' },
				],
				default: 'NONE',
				description: 'The trust center visibility of the document',
			},
			{
				displayName: 'Default Approver IDs',
				name: 'defaultApproverIds',
				type: 'string',
				default: '',
				description: 'Comma-separated list of default approver profile IDs',
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const documentId = this.getNodeParameter('documentId', itemIndex) as string;
	const updateFields = this.getNodeParameter('updateFields', itemIndex, {}) as {
		trustCenterVisibility?: string;
		defaultApproverIds?: string;
	};

	const query = `
		mutation UpdateDocument($input: UpdateDocumentInput!) {
			updateDocument(input: $input) {
				document {
					id
					status
					trustCenterVisibility
					currentPublishedMajor
					currentPublishedMinor
					archivedAt
					createdAt
					updatedAt
				}
			}
		}
	`;

	const input: Record<string, unknown> = { id: documentId };
	if (updateFields.trustCenterVisibility !== undefined) input.trustCenterVisibility = updateFields.trustCenterVisibility;
	if (updateFields.defaultApproverIds !== undefined && updateFields.defaultApproverIds !== '') {
		input.defaultApproverIds = updateFields.defaultApproverIds.split(',').map(id => id.trim()).filter(Boolean);
	}

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
