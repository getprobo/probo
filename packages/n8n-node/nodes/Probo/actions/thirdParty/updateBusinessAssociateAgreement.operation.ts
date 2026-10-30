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
import { proboApiRequest, toPeriod } from '../../GenericFunctions';
import { thirdPartyLocator, organizationLocator } from '../../locators';

export const description: INodeProperties[] = [
	organizationLocator({
		resource: ['thirdParty'],
		operation: ['updateBusinessAssociateAgreement'],
		required: false,
		description: 'Organization to search in. Leave this empty when entering an ID.',
	}),
	thirdPartyLocator({
		resource: ['thirdParty'],
		operation: ['updateBusinessAssociateAgreement'],
	}),
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
	const thirdPartyId = this.getNodeParameter('thirdPartyId', itemIndex, '', { extractValue: true }) as string;
	const validFrom = this.getNodeParameter('validFrom', itemIndex, '') as string;
	const validUntil = this.getNodeParameter('validUntil', itemIndex, '') as string;

	const query = `
		mutation UpdateThirdPartyBusinessAssociateAgreement($input: UpdateThirdPartyBusinessAssociateAgreementInput!) {
			updateThirdPartyBusinessAssociateAgreement(input: $input) {
				thirdPartyBusinessAssociateAgreement {
					id
					validity {
						start
						end
					}
				}
			}
		}
	`;

	const input: Record<string, unknown> = { thirdPartyId };
	const validity = toPeriod(validFrom, validUntil);
	if (validity) input.validity = validity;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
