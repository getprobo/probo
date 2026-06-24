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
		displayName: 'ThirdParty ID',
		name: 'thirdPartyId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['thirdParty'],
				operation: ['updateBusinessAssociateAgreement'],
			},
		},
		default: '',
		description: 'The ID of the thirdParty',
		required: true,
	},
	{
		displayName: 'Valid From',
		name: 'validFrom',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['thirdParty'],
				operation: ['updateBusinessAssociateAgreement'],
			},
		},
		default: '',
		description: 'The start date of the agreement validity (ISO 8601)',
	},
	{
		displayName: 'Valid Until',
		name: 'validUntil',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['thirdParty'],
				operation: ['updateBusinessAssociateAgreement'],
			},
		},
		default: '',
		description: 'The end date of the agreement validity (ISO 8601)',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const thirdPartyId = this.getNodeParameter('thirdPartyId', itemIndex) as string;
	const validFrom = this.getNodeParameter('validFrom', itemIndex, '') as string;
	const validUntil = this.getNodeParameter('validUntil', itemIndex, '') as string;

	const query = `
		mutation UpdateThirdPartyBusinessAssociateAgreement($input: UpdateThirdPartyBusinessAssociateAgreementInput!) {
			updateThirdPartyBusinessAssociateAgreement(input: $input) {
				thirdPartyBusinessAssociateAgreement {
					id
					validFrom
					validUntil
				}
			}
		}
	`;

	const input: Record<string, unknown> = { thirdPartyId };
	if (validFrom) input.validFrom = validFrom;
	if (validUntil) input.validUntil = validUntil;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
