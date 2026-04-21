// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
		displayName: 'Risk ID',
		name: 'riskId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['risk'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the risk to update',
		required: true,
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['risk'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The name of the risk',
	},
	{
		displayName: 'Category',
		name: 'category',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['risk'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The category of the risk',
	},
	{
		displayName: 'Treatment',
		name: 'treatment',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['risk'],
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
		description: 'The treatment strategy for the risk',
	},
	{
		displayName: 'Additional Fields',
		name: 'additionalFields',
		type: 'collection',
		placeholder: 'Add Field',
		default: {},
		displayOptions: {
			show: {
				resource: ['risk'],
				operation: ['update'],
			},
		},
		options: [
			{
				displayName: 'Description',
				name: 'description',
				type: 'string',
				default: '',
				description: 'The description of the risk',
			},
			{
				displayName: 'Inherent Impact',
				name: 'inherentImpact',
				type: 'number',
				typeOptions: {
					minValue: 1,
					maxValue: 5,
				},
				default: 1,
				description: 'The inherent impact of the risk (1-5)',
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
				description: 'The inherent likelihood of the risk (1-5)',
			},
			{
				displayName: 'Note',
				name: 'note',
				type: 'string',
				default: '',
				description: 'Additional notes about the risk',
			},
			{
				displayName: 'Owner ID',
				name: 'ownerId',
				type: 'string',
				default: '',
				description: 'The ID of the person who owns this risk',
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
				description: 'The residual impact of the risk (1-5)',
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
				description: 'The residual likelihood of the risk (1-5)',
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const riskId = this.getNodeParameter('riskId', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex, '') as string;
	const category = this.getNodeParameter('category', itemIndex, '') as string;
	const treatment = this.getNodeParameter('treatment', itemIndex, '') as string;
	const additionalFields = this.getNodeParameter('additionalFields', itemIndex, {}) as {
		description?: string;
		ownerId?: string;
		inherentLikelihood?: number;
		inherentImpact?: number;
		residualLikelihood?: number;
		residualImpact?: number;
		note?: string;
	};

	const query = `
		mutation UpdateRisk($input: UpdateRiskInput!) {
			updateRisk(input: $input) {
				risk {
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
				}
			}
		}
	`;

	const input: Record<string, unknown> = { id: riskId };
	if (name) input.name = name;
	if (category) input.category = category;
	if (treatment) input.treatment = treatment;
	if (additionalFields.description !== undefined) input.description = additionalFields.description === '' ? null : additionalFields.description;
	if (additionalFields.ownerId !== undefined) input.ownerId = additionalFields.ownerId === '' ? null : additionalFields.ownerId;
	if (additionalFields.inherentLikelihood !== undefined) input.inherentLikelihood = additionalFields.inherentLikelihood;
	if (additionalFields.inherentImpact !== undefined) input.inherentImpact = additionalFields.inherentImpact;
	if (additionalFields.residualLikelihood !== undefined) input.residualLikelihood = additionalFields.residualLikelihood;
	if (additionalFields.residualImpact !== undefined) input.residualImpact = additionalFields.residualImpact;
	if (additionalFields.note !== undefined) input.note = additionalFields.note === '' ? null : additionalFields.note;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
