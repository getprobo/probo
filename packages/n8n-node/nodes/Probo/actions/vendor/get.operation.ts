import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Vendor ID',
		name: 'vendorId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['vendor'],
				operation: ['get'],
			},
		},
		default: '',
		description: 'The ID of the vendor',
		required: true,
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
	const vendorId = this.getNodeParameter('vendorId', itemIndex) as string;
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
			primaryEmailAddress
		}`
		: '';

	const securityOwnerFragment = options.includeSecurityOwner
		? `securityOwner {
			id
			fullName
			primaryEmailAddress
		}`
		: '';

	const query = `
		query GetVendor($vendorId: ID!) {
			node(id: $vendorId) {
				... on Vendor {
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
		}
	`;

	const variables = {
		vendorId,
	};

	const responseData = await proboApiRequest.call(this, query, variables);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}

