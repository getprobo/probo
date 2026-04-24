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

import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Organization ID',
		name: 'organizationId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['thirdParty'],
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
				resource: ['thirdParty'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The name of the thirdParty',
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
				resource: ['thirdParty'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The description of the thirdParty',
	},
	{
		displayName: 'Category',
		name: 'category',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['thirdParty'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The category of the thirdParty',
	},
	{
		displayName: 'Website URL',
		name: 'websiteUrl',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['thirdParty'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The website URL of the thirdParty',
	},
	{
		displayName: 'Legal Name',
		name: 'legalName',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['thirdParty'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The legal name of the thirdParty',
	},
	{
		displayName: 'Headquarter Address',
		name: 'headquarterAddress',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['thirdParty'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The headquarter address of the thirdParty',
	},
	{
		displayName: 'Business Owner ID',
		name: 'businessOwnerId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['thirdParty'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The ID of the business owner (People ID)',
	},
	{
		displayName: 'Security Owner ID',
		name: 'securityOwnerId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['thirdParty'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The ID of the security owner (People ID)',
	},
	{
		displayName: 'Additional Fields',
		name: 'additionalFields',
		type: 'collection',
		placeholder: 'Add Field',
		default: {},
		displayOptions: {
			show: {
				resource: ['thirdParty'],
				operation: ['create'],
			},
		},
		options: [
			{
				displayName: 'Business Associate Agreement URL',
				name: 'businessAssociateAgreementUrl',
				type: 'string',
				default: '',
			},
			{
				displayName: 'Certifications',
				name: 'certifications',
				type: 'string',
				default: '',
				description: 'Comma-separated list of certifications',
			},
			{
				displayName: 'Countries',
				name: 'countries',
				type: 'string',
				default: '',
				description: 'Comma-separated list of country codes',
			},
			{
				displayName: 'Data Processing Agreement URL',
				name: 'dataProcessingAgreementUrl',
				type: 'string',
				default: '',
			},
			{
				displayName: 'Privacy Policy URL',
				name: 'privacyPolicyUrl',
				type: 'string',
				default: '',
			},
			{
				displayName: 'Security Page URL',
				name: 'securityPageUrl',
				type: 'string',
				default: '',
			},
			{
				displayName: 'Service Level Agreement URL',
				name: 'serviceLevelAgreementUrl',
				type: 'string',
				default: '',
			},
			{
				displayName: 'Status Page URL',
				name: 'statusPageUrl',
				type: 'string',
				default: '',
				description: 'The status page URL of the thirdParty',
			},
			{
				displayName: 'Subprocessors List URL',
				name: 'subprocessorsListUrl',
				type: 'string',
				default: '',
			},
			{
				displayName: 'Terms of Service URL',
				name: 'termsOfServiceUrl',
				type: 'string',
				default: '',
			},
			{
				displayName: 'Trust Page URL',
				name: 'trustPageUrl',
				type: 'string',
				default: '',
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex) as string;
	const description = this.getNodeParameter('description', itemIndex, '') as string;
	const category = this.getNodeParameter('category', itemIndex, '') as string;
	const websiteUrl = this.getNodeParameter('websiteUrl', itemIndex, '') as string;
	const legalName = this.getNodeParameter('legalName', itemIndex, '') as string;
	const headquarterAddress = this.getNodeParameter('headquarterAddress', itemIndex, '') as string;
	const businessOwnerId = this.getNodeParameter('businessOwnerId', itemIndex, '') as string;
	const securityOwnerId = this.getNodeParameter('securityOwnerId', itemIndex, '') as string;
	const additionalFields = this.getNodeParameter('additionalFields', itemIndex, {}) as {
		statusPageUrl?: string;
		termsOfServiceUrl?: string;
		privacyPolicyUrl?: string;
		serviceLevelAgreementUrl?: string;
		dataProcessingAgreementUrl?: string;
		businessAssociateAgreementUrl?: string;
		subprocessorsListUrl?: string;
		securityPageUrl?: string;
		trustPageUrl?: string;
		certifications?: string;
		countries?: string;
	};

	const query = `
		mutation CreateThirdParty($input: CreateThirdPartyInput!) {
			createThirdParty(input: $input) {
				thirdPartyEdge {
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
						createdAt
						updatedAt
					}
				}
			}
		}
	`;

	const input: Record<string, unknown> = {
		organizationId,
		name,
	};
	if (description) input.description = description;
	if (category) input.category = category;
	if (websiteUrl) input.websiteUrl = websiteUrl;
	if (legalName) input.legalName = legalName;
	if (headquarterAddress) input.headquarterAddress = headquarterAddress;
	if (businessOwnerId) input.businessOwnerId = businessOwnerId;
	if (securityOwnerId) input.securityOwnerId = securityOwnerId;
	if (additionalFields.statusPageUrl) input.statusPageUrl = additionalFields.statusPageUrl;
	if (additionalFields.termsOfServiceUrl) input.termsOfServiceUrl = additionalFields.termsOfServiceUrl;
	if (additionalFields.privacyPolicyUrl) input.privacyPolicyUrl = additionalFields.privacyPolicyUrl;
	if (additionalFields.serviceLevelAgreementUrl) input.serviceLevelAgreementUrl = additionalFields.serviceLevelAgreementUrl;
	if (additionalFields.dataProcessingAgreementUrl) input.dataProcessingAgreementUrl = additionalFields.dataProcessingAgreementUrl;
	if (additionalFields.businessAssociateAgreementUrl) input.businessAssociateAgreementUrl = additionalFields.businessAssociateAgreementUrl;
	if (additionalFields.subprocessorsListUrl) input.subprocessorsListUrl = additionalFields.subprocessorsListUrl;
	if (additionalFields.securityPageUrl) input.securityPageUrl = additionalFields.securityPageUrl;
	if (additionalFields.trustPageUrl) input.trustPageUrl = additionalFields.trustPageUrl;
	if (additionalFields.certifications) {
		input.certifications = additionalFields.certifications.split(',').map((c) => c.trim()).filter(Boolean);
	}
	if (additionalFields.countries) {
		input.countries = additionalFields.countries.split(',').map((c) => c.trim()).filter(Boolean);
	}

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}

