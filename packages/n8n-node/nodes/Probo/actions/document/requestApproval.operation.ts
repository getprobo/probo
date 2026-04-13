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
				operation: ['requestApproval'],
			},
		},
		default: '',
		description: 'The ID of the document',
		required: true,
	},
	{
		displayName: 'Approver IDs',
		name: 'approverIds',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['document'],
				operation: ['requestApproval'],
			},
		},
		default: '',
		description: 'Comma-separated list of approver profile IDs',
		required: true,
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
				operation: ['requestApproval'],
			},
		},
		default: '',
		description: 'The changelog for this version',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const documentId = this.getNodeParameter('documentId', itemIndex) as string;
	const approverIds = this.getNodeParameter('approverIds', itemIndex) as string;
	const changelog = this.getNodeParameter('changelog', itemIndex, '') as string;

	const query = `
		mutation RequestDocumentVersionApproval($input: RequestDocumentVersionApprovalInput!) {
			requestDocumentVersionApproval(input: $input) {
				approvalQuorum {
					id
					createdAt
					updatedAt
				}
			}
		}
	`;

	const input: Record<string, unknown> = {
		documentId,
		approverIds: approverIds.split(',').map(id => id.trim()).filter(Boolean),
	};
	if (changelog) input.changelog = changelog;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
