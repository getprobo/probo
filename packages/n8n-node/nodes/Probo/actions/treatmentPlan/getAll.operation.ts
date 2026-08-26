// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

const treatmentPlanNodeFields = `
	id
	treatment
	inherentLikelihood
	inherentImpact
	inherentRiskScore
	residualLikelihood
	residualImpact
	residualRiskScore
	netLikelihood
	netImpact
	netRiskScore
	risk {
		id
		name
		category
	}
	riskAnalysis {
		id
		name
	}
	owner {
		id
		fullName
	}
	createdAt
	updatedAt
`;

function treatmentPlansQuery(parentType: 'Organization' | 'Risk' | 'RiskAnalysis'): string {
	return `
		query GetTreatmentPlans($id: ID!, $first: Int, $after: CursorKey, $orderBy: TreatmentPlanOrder, $filter: TreatmentPlanFilter) {
			node(id: $id) {
				... on ${parentType} {
					treatmentPlans(first: $first, after: $after, orderBy: $orderBy, filter: $filter) {
						edges {
							node {
								${treatmentPlanNodeFields}
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
}

export const description: INodeProperties[] = [
	{
		displayName: 'Organization ID',
		name: 'organizationId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['treatmentPlan'],
				operation: ['getAll'],
			},
		},
		default: '',
		description: 'The ID of the organization. Ignored when Risk ID or Risk Analysis ID is set.',
		required: true,
	},
	{
		displayName: 'Risk ID',
		name: 'riskId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['treatmentPlan'],
				operation: ['getAll'],
			},
		},
		default: '',
		description: 'List treatment plans for this risk. Cannot be set together with Risk Analysis ID.',
	},
	{
		displayName: 'Risk Analysis ID',
		name: 'riskAnalysisId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['treatmentPlan'],
				operation: ['getAll'],
			},
		},
		default: '',
		description: 'List treatment plans for this risk analysis. Cannot be set together with Risk ID.',
	},
	{
		displayName: 'Return All',
		name: 'returnAll',
		type: 'boolean',
		displayOptions: {
			show: {
				resource: ['treatmentPlan'],
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
				resource: ['treatmentPlan'],
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
		displayName: 'Order By',
		name: 'orderBy',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['treatmentPlan'],
				operation: ['getAll'],
			},
		},
		options: [
			{ name: 'Category', value: 'CATEGORY' },
			{ name: 'Created At', value: 'CREATED_AT' },
			{ name: 'Inherent Risk Score', value: 'INHERENT_RISK_SCORE' },
			{ name: 'None', value: '' },
			{ name: 'Residual Risk Score', value: 'RESIDUAL_RISK_SCORE' },
			{ name: 'Treatment', value: 'TREATMENT' },
		],
		default: '',
		description: 'Field to order treatment plans by',
	},
	{
		displayName: 'Order Direction',
		name: 'orderDirection',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['treatmentPlan'],
				operation: ['getAll'],
			},
		},
		options: [
			{ name: 'Ascending', value: 'ASC' },
			{ name: 'Descending', value: 'DESC' },
		],
		default: 'DESC',
		description: 'Sort direction. Used when Order By is set.',
	},
	{
		displayName: 'Score Type',
		name: 'scoreType',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['treatmentPlan'],
				operation: ['getAll'],
			},
		},
		options: [
			{ name: 'Inherent', value: 'INHERENT' },
			{ name: 'Net', value: 'NET' },
			{ name: 'None', value: '' },
			{ name: 'Residual', value: 'RESIDUAL' },
		],
		default: '',
		description: 'Filter by matrix score type. Requires Likelihood and Impact.',
	},
	{
		displayName: 'Likelihood',
		name: 'likelihood',
		type: 'number',
		displayOptions: {
			show: {
				resource: ['treatmentPlan'],
				operation: ['getAll'],
			},
		},
		default: 0,
		description: 'Filter by likelihood. Requires Score Type and Impact.',
	},
	{
		displayName: 'Impact',
		name: 'impact',
		type: 'number',
		displayOptions: {
			show: {
				resource: ['treatmentPlan'],
				operation: ['getAll'],
			},
		},
		default: 0,
		description: 'Filter by impact. Requires Score Type and Likelihood.',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const riskId = this.getNodeParameter('riskId', itemIndex, '') as string;
	const riskAnalysisId = this.getNodeParameter('riskAnalysisId', itemIndex, '') as string;
	const returnAll = this.getNodeParameter('returnAll', itemIndex) as boolean;
	const limit = this.getNodeParameter('limit', itemIndex, 50) as number;
	const orderBy = this.getNodeParameter('orderBy', itemIndex, '') as string;
	const orderDirection = this.getNodeParameter('orderDirection', itemIndex, 'DESC') as string;
	const scoreType = this.getNodeParameter('scoreType', itemIndex, '') as string;
	const likelihood = this.getNodeParameter('likelihood', itemIndex, 0) as number;
	const impact = this.getNodeParameter('impact', itemIndex, 0) as number;

	if (riskId !== '' && riskAnalysisId !== '') {
		throw new Error('Risk ID and Risk Analysis ID cannot be set together');
	}

	const hasCell = scoreType !== '' || likelihood > 0 || impact > 0;
	if (hasCell && (scoreType === '' || likelihood <= 0 || impact <= 0)) {
		throw new Error('Score Type, Likelihood, and Impact must be set together');
	}

	let parentType: 'Organization' | 'Risk' | 'RiskAnalysis' = 'Organization';
	let id = organizationId;
	if (riskId !== '') {
		parentType = 'Risk';
		id = riskId;
	} else if (riskAnalysisId !== '') {
		parentType = 'RiskAnalysis';
		id = riskAnalysisId;
	}

	const variables: IDataObject = { id };
	if (orderBy !== '') {
		variables.orderBy = { field: orderBy, direction: orderDirection };
	}
	if (hasCell) {
		variables.filter = { scoreType, likelihood, impact };
	}

	const treatmentPlans = await proboApiRequestAllItems.call(
		this,
		treatmentPlansQuery(parentType),
		variables,
		(response) => {
			const data = response?.data as IDataObject | undefined;
			const node = data?.node as IDataObject | undefined;
			return node?.treatmentPlans as IDataObject | undefined;
		},
		returnAll,
		limit,
	);

	return {
		json: { treatmentPlans },
		pairedItem: { item: itemIndex },
	};
}
