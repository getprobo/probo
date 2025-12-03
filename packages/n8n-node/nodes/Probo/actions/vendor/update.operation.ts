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
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the vendor to update',
		required: true,
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['vendor'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The name of the vendor',
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
				resource: ['vendor'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The description of the vendor',
	},
	{
		displayName: 'Category',
		name: 'category',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['vendor'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The category of the vendor',
	},
	{
		displayName: 'Website URL',
		name: 'websiteUrl',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['vendor'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The website URL of the vendor',
	},
	{
		displayName: 'Legal Name',
		name: 'legalName',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['vendor'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The legal name of the vendor',
	},
	{
		displayName: 'Headquarter Address',
		name: 'headquarterAddress',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['vendor'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The headquarter address of the vendor',
	},
	{
		displayName: 'Business Owner ID',
		name: 'businessOwnerId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['vendor'],
				operation: ['update'],
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
				resource: ['vendor'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the security owner (People ID)',
	},
	{
		displayName: 'Show on Trust Center',
		name: 'showOnTrustCenter',
		type: 'boolean',
		displayOptions: {
			show: {
				resource: ['vendor'],
				operation: ['update'],
			},
		},
		default: false,
		description: 'Whether to show the vendor on the trust center',
	},
	{
		displayName: 'Additional Fields',
		name: 'additionalFields',
		type: 'collection',
		placeholder: 'Add Field',
		default: {},
		displayOptions: {
			show: {
				resource: ['vendor'],
				operation: ['update'],
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
				description: 'The status page URL of the vendor',
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
	const vendorId = this.getNodeParameter('vendorId', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex, '') as string;
	const description = this.getNodeParameter('description', itemIndex, '') as string;
	const category = this.getNodeParameter('category', itemIndex, '') as string;
	const websiteUrl = this.getNodeParameter('websiteUrl', itemIndex, '') as string;
	const legalName = this.getNodeParameter('legalName', itemIndex, '') as string;
	const headquarterAddress = this.getNodeParameter('headquarterAddress', itemIndex, '') as string;
	const businessOwnerId = this.getNodeParameter('businessOwnerId', itemIndex, '') as string;
	const securityOwnerId = this.getNodeParameter('securityOwnerId', itemIndex, '') as string;
	const showOnTrustCenter = this.getNodeParameter('showOnTrustCenter', itemIndex) as boolean | undefined;
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
		mutation UpdateVendor($input: UpdateVendorInput!) {
			updateVendor(input: $input) {
				vendor {
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
	`;

	const input: Record<string, unknown> = { id: vendorId };
	if (name) input.name = name;
	if (description !== undefined) input.description = description === '' ? null : description;
	if (category) input.category = category;
	if (websiteUrl !== undefined) input.websiteUrl = websiteUrl === '' ? null : websiteUrl;
	if (legalName !== undefined) input.legalName = legalName === '' ? null : legalName;
	if (headquarterAddress !== undefined) input.headquarterAddress = headquarterAddress === '' ? null : headquarterAddress;
	if (businessOwnerId !== undefined) input.businessOwnerId = businessOwnerId === '' ? null : businessOwnerId;
	if (securityOwnerId !== undefined) input.securityOwnerId = securityOwnerId === '' ? null : securityOwnerId;
	if (showOnTrustCenter !== undefined) input.showOnTrustCenter = showOnTrustCenter;
	if (additionalFields.statusPageUrl !== undefined) input.statusPageUrl = additionalFields.statusPageUrl === '' ? null : additionalFields.statusPageUrl;
	if (additionalFields.termsOfServiceUrl !== undefined) input.termsOfServiceUrl = additionalFields.termsOfServiceUrl === '' ? null : additionalFields.termsOfServiceUrl;
	if (additionalFields.privacyPolicyUrl !== undefined) input.privacyPolicyUrl = additionalFields.privacyPolicyUrl === '' ? null : additionalFields.privacyPolicyUrl;
	if (additionalFields.serviceLevelAgreementUrl !== undefined) input.serviceLevelAgreementUrl = additionalFields.serviceLevelAgreementUrl === '' ? null : additionalFields.serviceLevelAgreementUrl;
	if (additionalFields.dataProcessingAgreementUrl !== undefined) input.dataProcessingAgreementUrl = additionalFields.dataProcessingAgreementUrl === '' ? null : additionalFields.dataProcessingAgreementUrl;
	if (additionalFields.businessAssociateAgreementUrl !== undefined) input.businessAssociateAgreementUrl = additionalFields.businessAssociateAgreementUrl === '' ? null : additionalFields.businessAssociateAgreementUrl;
	if (additionalFields.subprocessorsListUrl !== undefined) input.subprocessorsListUrl = additionalFields.subprocessorsListUrl === '' ? null : additionalFields.subprocessorsListUrl;
	if (additionalFields.securityPageUrl !== undefined) input.securityPageUrl = additionalFields.securityPageUrl === '' ? null : additionalFields.securityPageUrl;
	if (additionalFields.trustPageUrl !== undefined) input.trustPageUrl = additionalFields.trustPageUrl === '' ? null : additionalFields.trustPageUrl;
	if (additionalFields.certifications !== undefined) {
		if (additionalFields.certifications === '') {
			input.certifications = [];
		} else {
			input.certifications = additionalFields.certifications.split(',').map((c) => c.trim()).filter(Boolean);
		}
	}
	if (additionalFields.countries !== undefined) {
		if (additionalFields.countries === '') {
			input.countries = [];
		} else {
			input.countries = additionalFields.countries.split(',').map((c) => c.trim()).filter(Boolean);
		}
	}

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}

