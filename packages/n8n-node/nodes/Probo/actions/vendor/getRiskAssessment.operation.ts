import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Vendor Risk Assessment ID',
		name: 'vendorRiskAssessmentId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['vendor'],
				operation: ['getRiskAssessment'],
			},
		},
		default: '',
		description: 'The ID of the vendor risk assessment',
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
				operation: ['getRiskAssessment'],
			},
		},
		options: [
			{
				displayName: 'Include Vendor',
				name: 'includeVendor',
				type: 'boolean',
				default: false,
				description: 'Whether to include vendor in the response',
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const vendorRiskAssessmentId = this.getNodeParameter('vendorRiskAssessmentId', itemIndex) as string;
	const options = this.getNodeParameter('options', itemIndex, {}) as {
		includeVendor?: boolean;
	};

	const vendorFragment = options.includeVendor
		? `vendor {
			id
			name
		}`
		: '';

	const query = `
		query GetVendorRiskAssessment($vendorRiskAssessmentId: ID!) {
			node(id: $vendorRiskAssessmentId) {
				... on VendorRiskAssessment {
					id
					expiresAt
					dataSensitivity
					businessImpact
					notes
					${vendorFragment}
					createdAt
					updatedAt
				}
			}
		}
	`;

	const variables = {
		vendorRiskAssessmentId,
	};

	const responseData = await proboApiRequest.call(this, query, variables);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
