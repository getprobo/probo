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

import type { INodeProperties, IExecuteFunctions, INodeExecutionData, IDataObject } from 'n8n-workflow';
import { proboApiRequestAllItems } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Organization ID',
		name: 'organizationId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['thirdParty'],
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
				resource: ['thirdParty'],
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
				resource: ['thirdParty'],
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
				resource: ['thirdParty'],
				operation: ['getAll'],
			},
		},
		options: [
			{
				displayName: 'Filter by Level',
				name: 'filterLevel',
				type: 'number',
				default: 0,
				description: 'Filter by third party level (1 = direct, 2+ = indirect, 0 = no filter)',
			},
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
		filterLevel?: number;
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

	const filterVariable = options.filterLevel ? ', $filter: ThirdPartyFilter' : '';
	const filterArgument = options.filterLevel ? ', filter: $filter' : '';

	const query = `
		query GetThirdParties($organizationId: ID!, $first: Int, $after: CursorKey${filterVariable}) {
			node(id: $organizationId) {
				... on Organization {
					thirdParties(first: $first, after: $after${filterArgument}) {
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
								level
								ancestors {
									id
									name
								}
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

	const variables: IDataObject = { organizationId };
	if (options.filterLevel) {
		variables.filter = { level: options.filterLevel };
	}

	const thirdParties = await proboApiRequestAllItems.call(
		this,
		query,
		variables,
		(response) => {
			const data = response?.data as IDataObject | undefined;
			const node = data?.node as IDataObject | undefined;
			return node?.thirdParties as IDataObject | undefined;
		},
		returnAll,
		limit,
	);

	return {
		json: { thirdParties },
		pairedItem: { item: itemIndex },
	};
}

