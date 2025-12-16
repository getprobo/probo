import type { INodeProperties, IExecuteFunctions, INodeExecutionData, IDataObject } from 'n8n-workflow';
import { proboApiRequestAllItems } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Vendor ID',
		name: 'vendorId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['vendor'],
				operation: ['getAllRiskAssessments'],
			},
		},
		default: '',
		description: 'The ID of the vendor',
		required: true,
	},
	{
		displayName: 'Return All',
		name: 'returnAll',
		type: 'boolean',
		displayOptions: {
			show: {
				resource: ['vendor'],
				operation: ['getAllRiskAssessments'],
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
				operation: ['getAllRiskAssessments'],
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
				operation: ['getAllRiskAssessments'],
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
	const vendorId = this.getNodeParameter('vendorId', itemIndex) as string;
	const returnAll = this.getNodeParameter('returnAll', itemIndex) as boolean;
	const limit = this.getNodeParameter('limit', itemIndex, 50) as number;
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
		query GetVendorRiskAssessments($vendorId: ID!, $first: Int, $after: CursorKey) {
			node(id: $vendorId) {
				... on Vendor {
					riskAssessments(first: $first, after: $after) {
						edges {
							node {
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
						pageInfo {
							hasNextPage
							endCursor
						}
					}
				}
			}
		}
	`;

	const vendorRiskAssessments = await proboApiRequestAllItems.call(
		this,
		query,
		{ vendorId },
		(response) => {
			const data = response?.data as IDataObject | undefined;
			const node = data?.node as IDataObject | undefined;
			return node?.riskAssessments as IDataObject | undefined;
		},
		returnAll,
		limit,
	);

	return {
		json: { vendorRiskAssessments },
		pairedItem: { item: itemIndex },
	};
}
