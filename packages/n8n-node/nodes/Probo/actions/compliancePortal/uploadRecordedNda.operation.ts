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

import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiMultipartRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Access ID',
		name: 'id',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['compliancePortal'],
				operation: ['uploadRecordedNda'],
			},
		},
		default: '',
		description: 'The ID of the compliance portal access',
		required: true,
	},
	{
		displayName: 'Signed At',
		name: 'signedAt',
		type: 'dateTime',
		displayOptions: {
			show: {
				resource: ['compliancePortal'],
				operation: ['uploadRecordedNda'],
			},
		},
		default: '',
		description: 'When the NDA was signed in the other system',
		required: true,
	},
	{
		displayName: 'Input Data Field Name',
		name: 'binaryPropertyName',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['compliancePortal'],
				operation: ['uploadRecordedNda'],
			},
		},
		default: 'data',
		description: 'The name of the input field containing the signed NDA PDF',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const id = this.getNodeParameter('id', itemIndex) as string;
	const signedAt = this.getNodeParameter('signedAt', itemIndex) as string;
	const binaryPropertyName = this.getNodeParameter('binaryPropertyName', itemIndex) as string;

	const binaryData = this.helpers.assertBinaryData(itemIndex, binaryPropertyName);
	const fileBuffer = await this.helpers.getBinaryDataBuffer(itemIndex, binaryPropertyName);

	const fileName = binaryData.fileName || 'signed-nda.pdf';
	const mimeType = binaryData.mimeType || 'application/pdf';

	const query = `
		mutation UploadCompliancePortalAccessRecordedNDA($input: UploadCompliancePortalAccessRecordedNDAInput!) {
			uploadCompliancePortalAccessRecordedNDA(input: $input) {
				compliancePortalAccess {
					id
					ndaBypassed
					recordedNdaSignedAt
					recordedNda {
						fileName
						downloadUrl
					}
				}
			}
		}
	`;

	const responseData = await proboApiMultipartRequest.call(
		this,
		query,
		{
			input: {
				compliancePortalAccessId: id,
				fileName,
				file: null,
				signedAt,
			},
		},
		'variables.input.file',
		fileBuffer,
		fileName,
		mimeType,
	);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
