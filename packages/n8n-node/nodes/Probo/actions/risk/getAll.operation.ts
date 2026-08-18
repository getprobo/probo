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

const riskNodeFields = `
	id
	name
	description
	category
	treatment
	inherentLikelihood
	inherentImpact
	inherentRiskScore
	residualLikelihood
	residualImpact
	residualRiskScore
	note
	createdAt
	updatedAt
`;

export const description: INodeProperties[] = [
	{
		displayName: 'Organization ID',
		name: 'organizationId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['risk'],
				operation: ['getAll'],
			},
		},
		default: '',
		description: 'The ID of the organization. Ignored when Risk Analysis ID is set.',
		required: true,
	},
	{
		displayName: 'Risk Analysis ID',
		name: 'riskAnalysisId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['risk'],
				operation: ['getAll'],
			},
		},
		default: '',
		description: 'List unplanned scenario-linked risks on this analysis',
	},
	{
		displayName: 'Return All',
		name: 'returnAll',
		type: 'boolean',
		displayOptions: {
			show: {
				resource: ['risk'],
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
				resource: ['risk'],
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
		displayName: 'Search',
		name: 'query',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['risk'],
				operation: ['getAll'],
			},
		},
		default: '',
		description: 'Filter risks by search query',
	},
	{
		displayName: 'Order By',
		name: 'orderBy',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['risk'],
				operation: ['getAll'],
			},
		},
		options: [
			{ name: 'Category', value: 'CATEGORY' },
			{ name: 'Created At', value: 'CREATED_AT' },
			{ name: 'Inherent Risk Score', value: 'INHERENT_RISK_SCORE' },
			{ name: 'Name', value: 'NAME' },
			{ name: 'None', value: '' },
			{ name: 'Residual Risk Score', value: 'RESIDUAL_RISK_SCORE' },
			{ name: 'Treatment', value: 'TREATMENT' },
		],
		default: '',
		description: 'Field to order risks by',
	},
	{
		displayName: 'Order Direction',
		name: 'orderDirection',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['risk'],
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
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const riskAnalysisId = this.getNodeParameter('riskAnalysisId', itemIndex, '') as string;
	const returnAll = this.getNodeParameter('returnAll', itemIndex) as boolean;
	const limit = this.getNodeParameter('limit', itemIndex, 50) as number;
	const queryFilter = this.getNodeParameter('query', itemIndex, '') as string;
	const orderBy = this.getNodeParameter('orderBy', itemIndex, '') as string;
	const orderDirection = this.getNodeParameter('orderDirection', itemIndex, 'DESC') as string;

	const byAnalysis = riskAnalysisId !== '';
	const id = byAnalysis ? riskAnalysisId : organizationId;
	const connectionField = byAnalysis ? 'scenarioRisks' : 'risks';
	const parentType = byAnalysis ? 'RiskAnalysis' : 'Organization';

	const query = `
		query GetRisks($id: ID!, $first: Int, $after: CursorKey, $orderBy: RiskOrder, $filter: RiskFilter) {
			node(id: $id) {
				... on ${parentType} {
					${connectionField}(first: $first, after: $after, orderBy: $orderBy, filter: $filter) {
						edges {
							node {
								${riskNodeFields}
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

	const variables: IDataObject = { id };
	if (orderBy !== '') {
		variables.orderBy = { field: orderBy, direction: orderDirection };
	}
	if (queryFilter !== '') {
		variables.filter = { query: queryFilter };
	}

	const risks = await proboApiRequestAllItems.call(
		this,
		query,
		variables,
		(response) => {
			const data = response?.data as IDataObject | undefined;
			const node = data?.node as IDataObject | undefined;
			return node?.[connectionField] as IDataObject | undefined;
		},
		returnAll,
		limit,
	);

	return {
		json: { risks },
		pairedItem: { item: itemIndex },
	};
}
