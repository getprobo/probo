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
import { proboApiMultipartRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Measure ID',
		name: 'measureId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['evidence'],
				operation: ['upload'],
			},
		},
		default: '',
		description: 'The ID of the measure to upload evidence for',
		required: true,
	},
	{
		displayName: 'Input Data Field Name',
		name: 'binaryPropertyName',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['evidence'],
				operation: ['upload'],
			},
		},
		default: 'data',
		description: 'The name of the input field containing the binary file data to upload',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const measureId = this.getNodeParameter('measureId', itemIndex) as string;
	const binaryPropertyName = this.getNodeParameter('binaryPropertyName', itemIndex) as string;

	const binaryData = this.helpers.assertBinaryData(itemIndex, binaryPropertyName);
	const fileBuffer = await this.helpers.getBinaryDataBuffer(itemIndex, binaryPropertyName);

	const fileName = binaryData.fileName || 'evidence';
	const mimeType = binaryData.mimeType || 'application/octet-stream';

	const query = `
		mutation UploadMeasureEvidence($input: UploadMeasureEvidenceInput!) {
			uploadMeasureEvidence(input: $input) {
				evidenceEdge {
					node {
						id
						state
						type
						description
						createdAt
						updatedAt
					}
				}
			}
		}
	`;

	const variables = {
		input: {
			measureId,
			file: null,
		},
	};

	const responseData = await proboApiMultipartRequest.call(
		this,
		query,
		variables,
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
