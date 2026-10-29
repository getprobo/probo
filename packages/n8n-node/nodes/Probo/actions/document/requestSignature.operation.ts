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
import { stampGraphqlNode, withParentReference } from '../../output';
import { documentVersionLocator, personLocator, organizationLocator, documentLocator } from '../../locators';

export const description: INodeProperties[] = [
	organizationLocator({
		resource: ['document'],
		operation: ['requestSignature'],
		required: false,
		description: 'Organization to search in. Leave this empty when entering an ID.',
	}),
	documentLocator({
		resource: ['document'],
		operation: ['requestSignature'],
		required: false,
		description: 'Document to search versions in. Leave this empty when entering a version ID.',
	}),
		documentVersionLocator({
		resource: ['document'],
		operation: ['requestSignature'],
	}),
	personLocator({
		name: 'signatoryId',
		displayName: 'Signatory',
		resource: ['document'],
		operation: ['requestSignature'],
	}),
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const documentVersionId = this.getNodeParameter('documentVersionId', itemIndex, '', { extractValue: true }) as string;
	const signatoryId = this.getNodeParameter('signatoryId', itemIndex, '', { extractValue: true }) as string;

	const query = `
		mutation RequestSignature($input: RequestSignatureInput!) {
			requestSignature(input: $input) {
				documentVersionSignatureEdge {
					node {
						id
						state
						signedAt
						requestedAt
						createdAt
						updatedAt
						documentVersion {
							id
							document {
								id
							}
						}
						signedBy {
							id
							fullName
							emailAddress
						}
					}
				}
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, {
		input: { documentVersionId, signatoryId },
	});

	return {
		json: stampGraphqlNode(
			responseData,
			['data', 'requestSignature', 'documentVersionSignatureEdge', 'node'],
			(node) => withParentReference(node, { versionId: documentVersionId }),
		),
		pairedItem: { item: itemIndex },
	};
}
