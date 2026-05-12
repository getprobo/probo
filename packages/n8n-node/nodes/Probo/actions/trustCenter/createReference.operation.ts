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
		displayName: 'Trust Center ID',
		name: 'trustCenterId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['trustCenter'],
				operation: ['createReference'],
			},
		},
		default: '',
		description: 'The ID of the trust center',
		required: true,
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['trustCenter'],
				operation: ['createReference'],
			},
		},
		default: '',
		description: 'The name of the reference',
		required: true,
	},
	{
		displayName: 'Description',
		name: 'description',
		type: 'string',
		typeOptions: {
			rows: 4,
		},
		displayOptions: {
			show: {
				resource: ['trustCenter'],
				operation: ['createReference'],
			},
		},
		default: '',
		description: 'The description of the reference',
	},
	{
		displayName: 'Website URL',
		name: 'websiteUrl',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['trustCenter'],
				operation: ['createReference'],
			},
		},
		default: '',
		description: 'The website URL of the reference',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const trustCenterId = this.getNodeParameter('trustCenterId', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex) as string;
	const description = this.getNodeParameter('description', itemIndex, '') as string;
	const websiteUrl = this.getNodeParameter('websiteUrl', itemIndex, '') as string;

	const query = `
		mutation CreateTrustCenterReference($input: CreateTrustCenterReferenceInput!) {
			createTrustCenterReference(input: $input) {
				trustCenterReferenceEdge {
					node {
						id
						name
						description
						websiteUrl
						rank
						createdAt
						updatedAt
					}
				}
			}
		}
	`;

	const input: Record<string, unknown> = { trustCenterId, name };
	if (description) input.description = description;
	if (websiteUrl) input.websiteUrl = websiteUrl;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
