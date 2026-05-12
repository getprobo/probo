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
		displayName: 'Organization ID',
		name: 'organizationId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['asset'],
				operation: ['publish'],
			},
		},
		default: '',
		description: 'The ID of the organization whose asset list to publish',
		required: true,
	},
	{
		displayName: 'Approver IDs',
		name: 'approverIds',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['asset'],
				operation: ['publish'],
			},
		},
		default: '',
		description: 'Comma-separated list of approver profile IDs',
	},
	{
		displayName: 'Minor',
		name: 'minor',
		type: 'boolean',
		displayOptions: {
			show: {
				resource: ['asset'],
				operation: ['publish'],
			},
		},
		default: false,
		description: 'Whether to publish as a minor version. Approvers are ignored when set. The list must already have a published major version.',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const approverIds = this.getNodeParameter('approverIds', itemIndex, '') as string;
	const minor = this.getNodeParameter('minor', itemIndex, false) as boolean;

	const query = `
		mutation PublishAssetList($input: PublishAssetListInput!) {
			publishAssetList(input: $input) {
				documentEdge {
					node {
						id
						status
						currentPublishedMajor
						currentPublishedMinor
						createdAt
						updatedAt
					}
				}
				documentVersionEdge {
					node {
						id
						title
						major
						minor
						status
						classification
						documentType
						publishedAt
						createdAt
						updatedAt
					}
				}
			}
		}
	`;

	const input: Record<string, unknown> = { organizationId, minor };

	if (approverIds) {
		input.approverIds = approverIds
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
