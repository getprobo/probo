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

import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'DPIA ID',
		name: 'id',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['dpia'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the DPIA to update',
		required: true,
	},
	{
		displayName: 'Description',
		name: 'description',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['dpia'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The description of the DPIA',
	},
	{
		displayName: 'Necessity and Proportionality',
		name: 'necessityAndProportionality',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['dpia'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The necessity and proportionality assessment',
	},
	{
		displayName: 'Potential Risk',
		name: 'potentialRisk',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['dpia'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The potential risk assessment',
	},
	{
		displayName: 'Mitigations',
		name: 'mitigations',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['dpia'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The mitigations for the identified risks',
	},
	{
		displayName: 'Residual Risk',
		name: 'residualRisk',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['dpia'],
				operation: ['update'],
			},
		},
		options: [
			{
				name: '(Unchanged)',
				value: '',
			},
			{
				name: 'Low',
				value: 'LOW',
			},
			{
				name: 'Medium',
				value: 'MEDIUM',
			},
			{
				name: 'High',
				value: 'HIGH',
			},
		],
		default: '',
		description: 'The residual risk level',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const id = this.getNodeParameter('id', itemIndex) as string;
	const description = this.getNodeParameter('description', itemIndex, '') as string;
	const necessityAndProportionality = this.getNodeParameter('necessityAndProportionality', itemIndex, '') as string;
	const potentialRisk = this.getNodeParameter('potentialRisk', itemIndex, '') as string;
	const mitigations = this.getNodeParameter('mitigations', itemIndex, '') as string;
	const residualRisk = this.getNodeParameter('residualRisk', itemIndex, '') as string;

	const query = `
		mutation UpdateDataProtectionImpactAssessment($input: UpdateDataProtectionImpactAssessmentInput!) {
			updateDataProtectionImpactAssessment(input: $input) {
				dataProtectionImpactAssessment {
					id
					description
					necessityAndProportionality
					potentialRisk
					mitigations
					residualRisk
					createdAt
					updatedAt
				}
			}
		}
	`;

	const input: Record<string, string> = { id };
	if (description) input.description = description;
	if (necessityAndProportionality) input.necessityAndProportionality = necessityAndProportionality;
	if (potentialRisk) input.potentialRisk = potentialRisk;
	if (mitigations) input.mitigations = mitigations;
	if (residualRisk) input.residualRisk = residualRisk;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
