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
import { proboApiMultipartRequest } from '../../GenericFunctions';
import { internalControlLocator } from '../../locators';

export const description: INodeProperties[] = [
	internalControlLocator({
		resource: ['evidence'],
		operation: ['createLink'],
	}),
	{
		displayName: 'URL',
		name: 'url',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['evidence'],
				operation: ['createLink'],
			},
		},
		default: '',
		description: 'HTTP or HTTPS address written into the .uri file',
		required: true,
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['evidence'],
				operation: ['createLink'],
			},
		},
		default: '',
		description: 'File name without the .uri extension. The node uploads that file through the evidence upload.',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const internalControlId = this.getNodeParameter('internalControlId', itemIndex, '', {
		extractValue: true,
	}) as string;
	const url = this.getNodeParameter('url', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex, '') as string;

	const file = linkEvidenceFile(name, url);
	if ('error' in file) {
		throw new NodeOperationError(this.getNode(), file.error, { itemIndex });
	}

	const query = `
		mutation UploadInternalControlEvidence($input: UploadInternalControlEvidenceInput!) {
			uploadInternalControlEvidence(input: $input) {
				evidenceEdge {
					node {
						id
						state
						type
						description
						createdAt
						updatedAt
						file {
							fileName
							mimeType
						}
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
				internalControlId,
				file: null,
			},
		},
		'variables.input.file',
		file.content,
		file.fileName,
		'text/uri-list',
	);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}

function linkEvidenceFile(
	name: string,
	rawURL: string,
): { content: Buffer; fileName: string } | { error: string } {
	const trimmedName = name.trim();
	const trimmedURL = rawURL.trim();

	if (trimmedName === '') {
		return { error: 'Name is required' };
	}

	if (/[\r\n]/.test(trimmedName)) {
		return { error: 'Name must be a single line' };
	}

	let parsed: URL;
	try {
		parsed = new URL(trimmedURL);
	} catch {
		return { error: 'URL must be an HTTP or HTTPS address' };
	}

	if ((parsed.protocol !== 'http:' && parsed.protocol !== 'https:') || parsed.host === '') {
		return { error: 'URL must be an HTTP or HTTPS address' };
	}

	return {
		content: Buffer.from(trimmedURL),
		fileName: `${trimmedName}.uri`,
	};
}
