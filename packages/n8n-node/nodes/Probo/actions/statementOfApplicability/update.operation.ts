// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
		displayName: 'Statement of Applicability ID',
		name: 'id',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['statementOfApplicability'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the statement of applicability to update',
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
				resource: ['statementOfApplicability'],
				operation: ['update'],
			},
		},
		options: [
			{
				displayName: 'Name',
				name: 'name',
				type: 'string',
				default: '',
				description: 'The name of the statement of applicability',
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
	const id = this.getNodeParameter('id', itemIndex) as string;
	const updateFields = this.getNodeParameter('updateFields', itemIndex, {}) as {
		name?: string;
		defaultApproverIds?: string;
	};

	const query = `
		mutation UpdateStatementOfApplicability($input: UpdateStatementOfApplicabilityInput!) {
			updateStatementOfApplicability(input: $input) {
				statementOfApplicability {
					id
					name
					createdAt
					updatedAt
				}
			}
		}
	`;

	const input: Record<string, unknown> = { id };
	if (updateFields.name) input.name = updateFields.name;
	if (updateFields.defaultApproverIds) {
		input.defaultApproverIds = updateFields.defaultApproverIds
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
