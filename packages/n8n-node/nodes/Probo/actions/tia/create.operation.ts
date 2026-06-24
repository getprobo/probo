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
		displayName: 'Processing Activity ID',
		name: 'processingActivityId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['tia'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The ID of the processing activity',
		required: true,
	},
	{
		displayName: 'Data Subjects',
		name: 'dataSubjects',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['tia'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The data subjects involved in the transfer',
		required: true,
	},
	{
		displayName: 'Legal Mechanism',
		name: 'legalMechanism',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['tia'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The legal mechanism for the transfer',
		required: true,
	},
	{
		displayName: 'Transfer',
		name: 'transfer',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['tia'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The transfer details',
		required: true,
	},
	{
		displayName: 'Local Law Risk',
		name: 'localLawRisk',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['tia'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The local law risk assessment',
		required: true,
	},
	{
		displayName: 'Supplementary Measures',
		name: 'supplementaryMeasures',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['tia'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The supplementary measures for the transfer',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const processingActivityId = this.getNodeParameter('processingActivityId', itemIndex) as string;
	const dataSubjects = this.getNodeParameter('dataSubjects', itemIndex) as string;
	const legalMechanism = this.getNodeParameter('legalMechanism', itemIndex) as string;
	const transfer = this.getNodeParameter('transfer', itemIndex) as string;
	const localLawRisk = this.getNodeParameter('localLawRisk', itemIndex) as string;
	const supplementaryMeasures = this.getNodeParameter('supplementaryMeasures', itemIndex) as string;

	const query = `
		mutation CreateTransferImpactAssessment($input: CreateTransferImpactAssessmentInput!) {
			createTransferImpactAssessment(input: $input) {
				transferImpactAssessmentEdge {
					node {
						id
						dataSubjects
						legalMechanism
						transfer
						localLawRisk
						supplementaryMeasures
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
			dataSubjects,
			legalMechanism,
			transfer,
			localLawRisk,
			supplementaryMeasures,
		},
	};

	const responseData = await proboApiRequest.call(this, query, variables);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
