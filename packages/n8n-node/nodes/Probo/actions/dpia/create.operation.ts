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
		displayName: 'Processing Activity ID',
		name: 'processingActivityId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['dpia'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The ID of the processing activity',
		required: true,
	},
	{
		displayName: 'Description',
		name: 'description',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['dpia'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The description of the DPIA',
		required: true,
	},
	{
		displayName: 'Necessity and Proportionality',
		name: 'necessityAndProportionality',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['dpia'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The necessity and proportionality assessment',
		required: true,
	},
	{
		displayName: 'Potential Risk',
		name: 'potentialRisk',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['dpia'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The potential risk assessment',
		required: true,
	},
	{
		displayName: 'Mitigations',
		name: 'mitigations',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['dpia'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The mitigations for the identified risks',
		required: true,
	},
	{
		displayName: 'Residual Risk',
		name: 'residualRisk',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['dpia'],
				operation: ['create'],
			},
		},
		options: [
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
		default: 'LOW',
		description: 'The residual risk level',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const processingActivityId = this.getNodeParameter('processingActivityId', itemIndex) as string;
	const description = this.getNodeParameter('description', itemIndex) as string;
	const necessityAndProportionality = this.getNodeParameter('necessityAndProportionality', itemIndex) as string;
	const potentialRisk = this.getNodeParameter('potentialRisk', itemIndex) as string;
	const mitigations = this.getNodeParameter('mitigations', itemIndex) as string;
	const residualRisk = this.getNodeParameter('residualRisk', itemIndex) as string;

	const query = `
		mutation CreateDataProtectionImpactAssessment($input: CreateDataProtectionImpactAssessmentInput!) {
			createDataProtectionImpactAssessment(input: $input) {
				dataProtectionImpactAssessmentEdge {
					node {
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
		}
	`;

	const variables = {
		input: {
			processingActivityId,
			description,
			necessityAndProportionality,
			potentialRisk,
			mitigations,
			residualRisk,
		},
	};

	const responseData = await proboApiRequest.call(this, query, variables);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
