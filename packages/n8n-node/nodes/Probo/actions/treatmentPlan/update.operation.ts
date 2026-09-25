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

import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Treatment Plan ID',
		name: 'treatmentPlanId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['treatmentPlan'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the treatment plan to update',
		required: true,
	},
	{
		displayName: 'Treatment',
		name: 'treatment',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['treatmentPlan'],
				operation: ['update'],
			},
		},
		options: [
			{ name: '(Unchanged)', value: '' },
			{
				name: 'Accepted',
				value: 'ACCEPTED',
			},
			{
				name: 'Avoided',
				value: 'AVOIDED',
			},
			{
				name: 'Mitigated',
				value: 'MITIGATED',
			},
			{
				name: 'Transferred',
				value: 'TRANSFERRED',
			},
		],
		default: '',
		description: 'The treatment strategy',
	},
	{
		displayName: 'Additional Fields',
		name: 'additionalFields',
		type: 'collection',
		placeholder: 'Add Field',
		default: {},
		displayOptions: {
			show: {
				resource: ['treatmentPlan'],
				operation: ['update'],
			},
		},
		options: [
			{
				displayName: 'Inherent Impact',
				name: 'inherentImpact',
				type: 'number',
				typeOptions: {
					minValue: 1,
					maxValue: 5,
				},
				default: 1,
				description: 'The inherent impact (1-5)',
			},
			{
				displayName: 'Inherent Likelihood',
				name: 'inherentLikelihood',
				type: 'number',
				typeOptions: {
					minValue: 1,
					maxValue: 5,
				},
				default: 1,
				description: 'The inherent likelihood (1-5)',
			},
			{
				displayName: 'Owner ID',
				name: 'ownerId',
				type: 'string',
				default: '',
				description: 'The ID of the person who owns this treatment plan',
			},
			{
				displayName: 'Residual Impact',
				name: 'residualImpact',
				type: 'number',
				typeOptions: {
					minValue: 1,
					maxValue: 5,
				},
				default: 1,
				description: 'The residual impact (1-5)',
			},
			{
				displayName: 'Residual Likelihood',
				name: 'residualLikelihood',
				type: 'number',
				typeOptions: {
					minValue: 1,
					maxValue: 5,
				},
				default: 1,
				description: 'The residual likelihood (1-5)',
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const treatmentPlanId = this.getNodeParameter('treatmentPlanId', itemIndex) as string;
	const treatment = this.getNodeParameter('treatment', itemIndex, '') as string;
	const additionalFields = this.getNodeParameter('additionalFields', itemIndex, {}) as {
		ownerId?: string;
		inherentLikelihood?: number;
		inherentImpact?: number;
		residualLikelihood?: number;
		residualImpact?: number;
	};

	const query = `
		mutation UpdateTreatmentPlan($input: UpdateTreatmentPlanInput!) {
			updateTreatmentPlan(input: $input) {
				treatmentPlan {
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
					createdAt
					updatedAt
				}
			}
		}
	`;

	const input: Record<string, unknown> = { id: treatmentPlanId };
	if (treatment) input.treatment = treatment;
	if (additionalFields.ownerId) input.ownerId = additionalFields.ownerId;
	if (additionalFields.inherentLikelihood !== undefined) input.inherentLikelihood = additionalFields.inherentLikelihood;
	if (additionalFields.inherentImpact !== undefined) input.inherentImpact = additionalFields.inherentImpact;
	if (additionalFields.residualLikelihood !== undefined) input.residualLikelihood = additionalFields.residualLikelihood;
	if (additionalFields.residualImpact !== undefined) input.residualImpact = additionalFields.residualImpact;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
