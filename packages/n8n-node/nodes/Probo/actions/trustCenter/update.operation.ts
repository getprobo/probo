// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the trust center to update',
		required: true,
	},
	{
		displayName: 'Active',
		name: 'active',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['trustCenter'],
				operation: ['update'],
			},
		},
		options: [
			{
				name: '(Unchanged)',
				value: '',
			},
			{
				name: 'Active',
				value: 'true',
			},
			{
				name: 'Inactive',
				value: 'false',
			},
		],
		default: '',
		description: 'Whether the trust center is active',
	},
	{
		displayName: 'Search Engine Indexing',
		name: 'searchEngineIndexing',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['trustCenter'],
				operation: ['update'],
			},
		},
		options: [
			{
				name: '(Unchanged)',
				value: '',
			},
			{
				name: 'Indexable',
				value: 'INDEXABLE',
			},
			{
				name: 'Not Indexable',
				value: 'NOT_INDEXABLE',
			},
		],
		default: '',
		description: 'Whether search engines should index the trust center',
	},
	{
		displayName: 'Description',
		name: 'description',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['trustCenter'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The description shown on the compliance page',
	},
	{
		displayName: 'Website URL',
		name: 'websiteUrl',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['trustCenter'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The website URL shown on the compliance page',
	},
	{
		displayName: 'Email',
		name: 'email',
		type: 'string',
		placeholder: 'name@example.com',
		displayOptions: {
			show: {
				resource: ['trustCenter'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The contact email shown on the compliance page',
	},
	{
		displayName: 'Headquarter Address',
		name: 'headquarterAddress',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['trustCenter'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The headquarter address shown on the compliance page',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const trustCenterId = this.getNodeParameter('trustCenterId', itemIndex) as string;
	const active = this.getNodeParameter('active', itemIndex, '') as string;
	const searchEngineIndexing = this.getNodeParameter('searchEngineIndexing', itemIndex, '') as string;
	const description = this.getNodeParameter('description', itemIndex, '') as string;
	const websiteUrl = this.getNodeParameter('websiteUrl', itemIndex, '') as string;
	const email = this.getNodeParameter('email', itemIndex, '') as string;
	const headquarterAddress = this.getNodeParameter('headquarterAddress', itemIndex, '') as string;

	const hasSettingsUpdate = active !== '' || searchEngineIndexing !== '';
	const hasBrandUpdate = description || websiteUrl || email || headquarterAddress;

	if (!hasSettingsUpdate && !hasBrandUpdate) {
		throw new Error('At least one field must be provided to update the trust center');
	}

	let responseData: Record<string, unknown> = {};

	if (hasSettingsUpdate) {
		const query = `
			mutation UpdateTrustCenter($input: UpdateTrustCenterInput!) {
				updateTrustCenter(input: $input) {
					trustCenter {
						id
						active
						searchEngineIndexing
						description
						websiteUrl
						email
						headquarterAddress
						createdAt
						updatedAt
					}
				}
			}
		`;

		const input: Record<string, unknown> = { trustCenterId };
		if (active !== '') input.active = active === 'true';
		if (searchEngineIndexing) input.searchEngineIndexing = searchEngineIndexing;

		responseData = await proboApiRequest.call(this, query, { input }) as Record<string, unknown>;
	}

	if (hasBrandUpdate) {
		const brandQuery = `
			mutation UpdateTrustCenterBrand($input: UpdateTrustCenterBrandInput!) {
				updateTrustCenterBrand(input: $input) {
					trustCenter {
						id
						active
						searchEngineIndexing
						description
						websiteUrl
						email
						headquarterAddress
						createdAt
						updatedAt
					}
				}
			}
		`;

		const brandInput: Record<string, string> = { trustCenterId };
		if (description) brandInput.description = description;
		if (websiteUrl) brandInput.websiteUrl = websiteUrl;
		if (email) brandInput.email = email;
		if (headquarterAddress) brandInput.headquarterAddress = headquarterAddress;

		responseData = await proboApiRequest.call(this, brandQuery, { input: brandInput }) as Record<string, unknown>;
	}

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
