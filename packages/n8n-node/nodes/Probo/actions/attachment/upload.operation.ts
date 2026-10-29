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
import { organizationLocator } from '../../locators';
import { NodeOperationError } from 'n8n-workflow';
import { proboApiMultipartRequest, readResourceId } from '../../GenericFunctions';
import { resolveAttachmentFile } from './attachmentFile';

export const description: INodeProperties[] = [
	organizationLocator({
		resource: ['attachment'],
		operation: ['upload'],
		description: 'The ID of the organization that will own the file',
	}),
	{
		displayName: 'Input Data Field Name',
		name: 'binaryPropertyName',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['attachment'],
				operation: ['upload'],
			},
		},
		default: 'data',
		description:
			'The name of the input field containing the binary file. Provide a supported filename or MIME type.',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = readResourceId(this.getNodeParameter('organizationId', itemIndex));
	const binaryPropertyName = this.getNodeParameter('binaryPropertyName', itemIndex) as string;

	const binaryData = this.helpers.assertBinaryData(itemIndex, binaryPropertyName);
	const fileBuffer = await this.helpers.getBinaryDataBuffer(itemIndex, binaryPropertyName);

	let fileName: string;
	let mimeType: string;
	try {
		({ fileName, mimeType } = resolveAttachmentFile(binaryData.fileName, binaryData.mimeType));
	} catch (error) {
		const message = error instanceof Error ? error.message : 'Cannot upload attachment';
		throw new NodeOperationError(this.getNode(), message, { itemIndex });
	}

	const query = `
		mutation UploadFile($input: UploadFileInput!) {
			uploadFile(input: $input) {
				file {
					id
					fileName
					mimeType
					size
				}
			}
		}
	`;

	const input = {
		organizationId,
		file: null,
	};

	const responseData = await proboApiMultipartRequest.call(
		this,
		query,
		{ input },
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
