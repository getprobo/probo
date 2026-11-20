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

import type { IExecuteFunctions, INodeExecutionData, INodeProperties } from 'n8n-workflow';
import { proboApiRequest, toPeriod } from '../../GenericFunctions';

const agreementTypes = [
	{ name: 'Terms (CGV/CGU)', value: 'TERMS' },
	{ name: 'Master Service Agreement', value: 'MASTER_SERVICE' },
	{ name: 'Business Associate Agreement', value: 'BUSINESS_ASSOCIATE' },
	{ name: 'Data Privacy Agreement', value: 'DATA_PRIVACY' },
	{ name: 'Amendment', value: 'AMENDMENT' },
	{ name: 'Annex', value: 'ANNEX' },
	{ name: 'Other', value: 'OTHER' },
];

export const description: INodeProperties[] = [
	{
		displayName: 'Agreement ID',
		name: 'agreementId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['thirdParty'],
				operation: ['updateAgreement'],
			},
		},
		default: '',
		required: true,
		description: 'The ID of the third party agreement to update',
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['thirdParty'],
				operation: ['updateAgreement'],
			},
		},
		default: '',
		description: 'New name. Leave empty to keep the current name.',
	},
	{
		displayName: 'Type',
		name: 'agreementType',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['thirdParty'],
				operation: ['updateAgreement'],
			},
		},
		options: [{ name: 'Keep Current', value: '' }, ...agreementTypes],
		default: '',
		description: 'Kind of agreement. Keep current to leave it unchanged.',
	},
	{
		displayName: 'Valid From',
		name: 'validFrom',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['thirdParty'],
				operation: ['updateAgreement'],
			},
		},
		default: '',
		description: 'Start of the validity period (ISO 8601). Leave empty to keep the current date.',
	},
	{
		displayName: 'Valid Until',
		name: 'validUntil',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['thirdParty'],
				operation: ['updateAgreement'],
			},
		},
		default: '',
		description: 'End of the validity period (ISO 8601). Leave empty to keep the current date.',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const agreementId = this.getNodeParameter('agreementId', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex, '') as string;
	const agreementType = this.getNodeParameter('agreementType', itemIndex, '') as string;
	const validFrom = this.getNodeParameter('validFrom', itemIndex, '') as string;
	const validUntil = this.getNodeParameter('validUntil', itemIndex, '') as string;

	const query = `
		mutation UpdateThirdPartyAgreement($input: UpdateThirdPartyAgreementInput!) {
			updateThirdPartyAgreement(input: $input) {
				thirdPartyAgreement {
					id
					name
					type
					validity {
						start
						end
					}
				}
			}
		}
	`;

	const input: Record<string, unknown> = { id: agreementId };
	if (name) {
		input.name = name;
	}
	if (agreementType) {
		input.type = agreementType;
	}
	const validity = toPeriod(validFrom, validUntil);
	if (validity) {
		input.validity = validity;
	}

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
