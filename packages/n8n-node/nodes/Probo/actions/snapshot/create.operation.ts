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
		displayName: 'Organization ID',
		name: 'organizationId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['snapshot'],
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
				resource: ['snapshot'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The name of the snapshot',
		required: true,
	},
	{
		displayName: 'Description',
		name: 'description',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['snapshot'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The description of the snapshot',
	},
	{
		displayName: 'Type',
		name: 'type',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['snapshot'],
				operation: ['create'],
			},
		},
		options: [
			{
				name: 'Assets',
				value: 'ASSETS',
			},
			{
				name: 'Findings',
				value: 'FINDINGS',
			},
			{
				name: 'Obligations',
				value: 'OBLIGATIONS',
			},
			{
				name: 'Processing Activities',
				value: 'PROCESSING_ACTIVITIES',
			},
			{
				name: 'Risks',
				value: 'RISKS',
			},
			{
				name: 'Statements of Applicability',
				value: 'STATEMENTS_OF_APPLICABILITY',
			},
			{
				name: 'Vendors',
				value: 'VENDORS',
			},
		],
		default: 'RISKS',
		description: 'The type of snapshot',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex) as string;
	const description = this.getNodeParameter('description', itemIndex, '') as string;
	const type = this.getNodeParameter('type', itemIndex) as string;

	const query = `
		mutation CreateSnapshot($input: CreateSnapshotInput!) {
			createSnapshot(input: $input) {
				snapshotEdge {
					node {
						id
						name
						description
						type
						createdAt
					}
				}
			}
		}
	`;

	const variables = {
		input: {
			organizationId,
			name,
			...(description && { description }),
			type,
		},
	};

	const responseData = await proboApiRequest.call(this, query, variables);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
