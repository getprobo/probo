// Copyright (c) 2025 Probo Inc <hello@probo.com>.
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

import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';
import { thirdPartyLocator, organizationLocator } from '../../locators';

export const description: INodeProperties[] = [
	organizationLocator({
		resource: ['thirdParty'],
		operation: ['get'],
		required: false,
		description: 'Organization to search in. Leave this empty when entering an ID.',
	}),
	thirdPartyLocator({
		resource: ['thirdParty'],
		operation: ['get'],
	}),
	{
		displayName: 'Options',
		name: 'options',
		type: 'collection',
		placeholder: 'Add Option',
		default: {},
		displayOptions: {
			show: {
				resource: ['thirdParty'],
				operation: ['get'],
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
				displayName: 'Include Administrators',
				name: 'includeAdministrators',
				type: 'boolean',
				default: false,
				description: 'Whether to include administrators details',
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const thirdPartyId = this.getNodeParameter('thirdPartyId', itemIndex, '', { extractValue: true }) as string;
	const options = this.getNodeParameter('options', itemIndex, {}) as {
		includeOrganization?: boolean;
		includeAdministrators?: boolean;
	};

	const organizationFragment = options.includeOrganization
		? `organization {
			id
			name
		}`
		: '';

	const administratorsFragment = options.includeAdministrators
		? `administrators {
			id
			fullName
			emailAddress
		}`
		: '';

	const query = `
		query GetThirdParty($thirdPartyId: ID!) {
			node(id: $thirdPartyId) {
				... on ThirdParty {
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
					${organizationFragment}
					${administratorsFragment}
					createdAt
					updatedAt
				}
			}
		}
	`;

	const variables = {
		thirdPartyId,
	};

	const responseData = await proboApiRequest.call(this, query, variables);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}

