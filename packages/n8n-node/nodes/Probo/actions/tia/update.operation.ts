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
		displayName: 'TIA ID',
		name: 'id',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['tia'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the TIA to update',
		required: true,
	},
	{
		displayName: 'Data Subjects',
		name: 'dataSubjects',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['tia'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The data subjects involved in the transfer',
	},
	{
		displayName: 'Legal Mechanism',
		name: 'legalMechanism',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['tia'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The legal mechanism for the transfer',
	},
	{
		displayName: 'Transfer',
		name: 'transfer',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['tia'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The transfer details',
	},
	{
		displayName: 'Local Law Risk',
		name: 'localLawRisk',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['tia'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The local law risk assessment',
	},
	{
		displayName: 'Supplementary Measures',
		name: 'supplementaryMeasures',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['tia'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The supplementary measures for the transfer',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const id = this.getNodeParameter('id', itemIndex) as string;
	const dataSubjects = this.getNodeParameter('dataSubjects', itemIndex, '') as string;
	const legalMechanism = this.getNodeParameter('legalMechanism', itemIndex, '') as string;
	const transfer = this.getNodeParameter('transfer', itemIndex, '') as string;
	const localLawRisk = this.getNodeParameter('localLawRisk', itemIndex, '') as string;
	const supplementaryMeasures = this.getNodeParameter('supplementaryMeasures', itemIndex, '') as string;

	const query = `
		mutation UpdateTransferImpactAssessment($input: UpdateTransferImpactAssessmentInput!) {
			updateTransferImpactAssessment(input: $input) {
				transferImpactAssessment {
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
	`;

	const input: Record<string, string> = { id };
	if (dataSubjects) input.dataSubjects = dataSubjects;
	if (legalMechanism) input.legalMechanism = legalMechanism;
	if (transfer) input.transfer = transfer;
	if (localLawRisk) input.localLawRisk = localLawRisk;
	if (supplementaryMeasures) input.supplementaryMeasures = supplementaryMeasures;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
