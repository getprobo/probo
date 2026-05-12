// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

import type { INodeProperties, IExecuteFunctions, INodeExecutionData, IDataObject } from 'n8n-workflow';
import { proboApiRequestAllItems } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Organization ID',
		name: 'organizationId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['vendor'],
				operation: ['getAll'],
			},
		},
		default: '',
		description: 'The ID of the organization',
		required: true,
	},
	{
		displayName: 'Return All',
		name: 'returnAll',
		type: 'boolean',
		displayOptions: {
			show: {
				resource: ['vendor'],
				operation: ['getAll'],
			},
		},
		default: false,
		description: 'Whether to return all results or only up to a given limit',
	},
	{
		displayName: 'Limit',
		name: 'limit',
		type: 'number',
		displayOptions: {
			show: {
				resource: ['vendor'],
				operation: ['getAll'],
				returnAll: [false],
			},
		},
		typeOptions: {
			minValue: 1,
		},
		default: 50,
		description: 'Max number of results to return',
	},
	{
		displayName: 'Options',
		name: 'options',
		type: 'collection',
		placeholder: 'Add Option',
		default: {},
		displayOptions: {
			show: {
				resource: ['vendor'],
				operation: ['getAll'],
			},
		},
		options: [
			{
				displayName: 'Include Organization',
				name: 'includeOrganization',
				type: 'boolean',
				default: false,
				description: 'Whether to include organization in the response',
			},
			{
				displayName: 'Include Business Owner',
				name: 'includeBusinessOwner',
				type: 'boolean',
				default: false,
				description: 'Whether to include business owner in the response',
			},
			{
				displayName: 'Include Security Owner',
				name: 'includeSecurityOwner',
				type: 'boolean',
				default: false,
				description: 'Whether to include security owner in the response',
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const returnAll = this.getNodeParameter('returnAll', itemIndex) as boolean;
	const limit = this.getNodeParameter('limit', itemIndex, 50) as number;
	const options = this.getNodeParameter('options', itemIndex, {}) as {
		includeOrganization?: boolean;
		includeBusinessOwner?: boolean;
		includeSecurityOwner?: boolean;
	};

	const organizationFragment = options.includeOrganization
		? `organization {
			id
			name
		}`
		: '';

	const businessOwnerFragment = options.includeBusinessOwner
		? `businessOwner {
			id
			fullName
			emailAddress
		}`
		: '';

	const securityOwnerFragment = options.includeSecurityOwner
		? `securityOwner {
			id
			fullName
			emailAddress
		}`
		: '';

	const query = `
		query GetVendors($organizationId: ID!, $first: Int, $after: CursorKey) {
			node(id: $organizationId) {
				... on Organization {
					vendors(first: $first, after: $after) {
						edges {
							node {
								id
								name
								description
								category
								websiteUrl
								legalName
								headquarterAddress
								statusPageUrl
								termsOfServiceUrl
								privacyPolicyUrl
								serviceLevelAgreementUrl
								dataProcessingAgreementUrl
								businessAssociateAgreementUrl
								subprocessorsListUrl
								securityPageUrl
								trustPageUrl
								certifications
								countries
								showOnTrustCenter
								${organizationFragment}
								${businessOwnerFragment}
								${securityOwnerFragment}
								createdAt
								updatedAt
							}
						}
						pageInfo {
							hasNextPage
							endCursor
						}
					}
				}
			}
		}
	`;

	const vendors = await proboApiRequestAllItems.call(
		this,
		query,
		{ organizationId },
		(response) => {
			const data = response?.data as IDataObject | undefined;
			const node = data?.node as IDataObject | undefined;
			return node?.vendors as IDataObject | undefined;
		},
		returnAll,
		limit,
	);

	return {
		json: { vendors },
		pairedItem: { item: itemIndex },
	};
}

