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
import { NodeOperationError } from 'n8n-workflow';
import { proboApiMultipartRequest, toPeriod } from '../../GenericFunctions';
import { organizationLocator, thirdPartyLocator } from '../../locators';

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
	organizationLocator({
		resource: ['thirdParty'],
		operation: ['uploadAgreement'],
		required: false,
		description: 'Organization to search in. Leave this empty when entering an ID.',
	}),
	thirdPartyLocator({
		resource: ['thirdParty'],
		operation: ['uploadAgreement'],
	}),
	{
		displayName: 'Type',
		name: 'agreementType',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['thirdParty'],
				operation: ['uploadAgreement'],
			},
		},
		options: agreementTypes,
		default: 'OTHER',
		required: true,
		description: 'Kind of agreement',
	},
	{
		displayName: 'Input Data Field Name',
		name: 'binaryPropertyName',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['thirdParty'],
				operation: ['uploadAgreement'],
			},
		},
		default: 'data',
		required: true,
		description: 'The name of the input field containing the PDF file',
	},
	{
		displayName: 'Valid From',
		name: 'validFrom',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['thirdParty'],
				operation: ['uploadAgreement'],
			},
		},
		default: '',
		description: 'Start of the validity period (ISO 8601). Leave empty for no start date.',
	},
	{
		displayName: 'Valid Until',
		name: 'validUntil',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['thirdParty'],
				operation: ['uploadAgreement'],
			},
		},
		default: '',
		description: 'End of the validity period (ISO 8601). Leave empty for no end date.',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const thirdPartyId = this.getNodeParameter('thirdPartyId', itemIndex, '', { extractValue: true }) as string;
	const agreementType = this.getNodeParameter('agreementType', itemIndex) as string;
	const binaryPropertyName = this.getNodeParameter('binaryPropertyName', itemIndex) as string;
	const validFrom = this.getNodeParameter('validFrom', itemIndex, '') as string;
	const validUntil = this.getNodeParameter('validUntil', itemIndex, '') as string;

	const binaryData = this.helpers.assertBinaryData(itemIndex, binaryPropertyName);
	const fileBuffer = await this.helpers.getBinaryDataBuffer(itemIndex, binaryPropertyName);
	const fileName = binaryData.fileName || 'agreement.pdf';
	const mimeType = binaryData.mimeType || 'application/pdf';

	if (mimeType !== 'application/pdf' && !fileName.toLowerCase().endsWith('.pdf')) {
		throw new NodeOperationError(this.getNode(), 'Third party agreements accept PDF files only');
	}

	const query = `
		mutation UploadThirdPartyAgreement($input: UploadThirdPartyAgreementInput!) {
			uploadThirdPartyAgreement(input: $input) {
				thirdPartyAgreementEdge {
					node {
						id
						type
						validity {
							start
							end
						}
						file {
							fileName
							downloadUrl
						}
						createdAt
						updatedAt
					}
				}
			}
		}
	`;

	const input: Record<string, unknown> = {
		thirdPartyId,
		type: agreementType,
		file: null,
	};
	const validity = toPeriod(validFrom, validUntil);
	if (validity) {
		input.validity = validity;
	}

	const responseData = await proboApiMultipartRequest.call(
		this,
		query,
		{ input },
		'variables.input.file',
		fileBuffer,
		fileName,
		mimeType === 'application/pdf' ? mimeType : 'application/pdf',
	);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
