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
		displayName: 'Vendor ID',
		name: 'vendorId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['vendor'],
				operation: ['updateDataPrivacyAgreement'],
			},
		},
		default: '',
		description: 'The ID of the vendor',
		required: true,
	},
	{
		displayName: 'Valid From',
		name: 'validFrom',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['vendor'],
				operation: ['updateDataPrivacyAgreement'],
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
				resource: ['vendor'],
				operation: ['updateDataPrivacyAgreement'],
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
	const vendorId = this.getNodeParameter('vendorId', itemIndex) as string;
	const validFrom = this.getNodeParameter('validFrom', itemIndex, '') as string;
	const validUntil = this.getNodeParameter('validUntil', itemIndex, '') as string;

	const query = `
		mutation UpdateVendorDataPrivacyAgreement($input: UpdateVendorDataPrivacyAgreementInput!) {
			updateVendorDataPrivacyAgreement(input: $input) {
				vendorDataPrivacyAgreement {
					id
					validFrom
					validUntil
				}
			}
		}
	`;

	const input: Record<string, unknown> = { vendorId };
	if (validFrom) input.validFrom = validFrom;
	if (validUntil) input.validUntil = validUntil;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
