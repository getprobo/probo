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
		displayName: 'Approval Decision ID',
		name: 'approvalDecisionId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['document'],
				operation: ['getApprovalDecision'],
			},
		},
		default: '',
		description: 'The ID of the document version approval decision',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const approvalDecisionId = this.getNodeParameter('approvalDecisionId', itemIndex) as string;

	const query = `
		query GetDocumentVersionApprovalDecision($approvalDecisionId: ID!) {
			node(id: $approvalDecisionId) {
				... on DocumentVersionApprovalDecision {
					id
					state
					comment
					decidedAt
					approver {
						id
						fullName
					}
					createdAt
					updatedAt
				}
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, { approvalDecisionId });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
