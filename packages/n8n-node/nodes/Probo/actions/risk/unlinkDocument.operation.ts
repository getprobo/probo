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
import { organizationLocator, documentLocator, riskLocator } from '../../locators';
import { proboApiRequest, readResourceId } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	riskLocator({
		resource: ['risk'],
		operation: ['unlinkDocument'],
	}),
	organizationLocator({
		resource: ['risk'],
		operation: ['unlinkDocument'],
		required: false,
		description: 'Organization to search documents in. Leave this empty when entering a document ID.',
	}),
	documentLocator({
		resource: ['risk'],
		operation: ['unlinkDocument'],
		description: 'The ID of the document to unlink',
	}),
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const riskId = readResourceId(this.getNodeParameter('riskId', itemIndex));
	const documentId = readResourceId(this.getNodeParameter('documentId', itemIndex));

	const query = `
		mutation DeleteRiskDocumentMapping($input: DeleteRiskDocumentMappingInput!) {
			deleteRiskDocumentMapping(input: $input) {
				deletedRiskId
				deletedDocumentId
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, { input: { riskId, documentId } });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
