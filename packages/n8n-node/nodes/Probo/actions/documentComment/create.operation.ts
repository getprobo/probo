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
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Document ID',
		name: 'documentId',
		type: 'string',
		required: true,
		default: '',
		displayOptions: { show: { resource: ['documentComment'], operation: ['create'] } },
	},
	{
		displayName: 'Content',
		name: 'content',
		type: 'string',
		typeOptions: { rows: 4 },
		required: true,
		default: '',
		displayOptions: { show: { resource: ['documentComment'], operation: ['create'] } },
	},
	{
		displayName: 'Quote',
		name: 'quote',
		type: 'string',
		default: '',
		description: 'Document text to anchor. Leave empty when commenting on a suggestion.',
		displayOptions: { show: { resource: ['documentComment'], operation: ['create'] } },
	},
	{
		displayName: 'Quote Occurrence',
		name: 'quoteOccurrence',
		type: 'number',
		typeOptions: { minValue: 0 },
		default: 0,
		displayOptions: { show: { resource: ['documentComment'], operation: ['create'] } },
	},
	{
		displayName: 'Suggestion ID',
		name: 'suggestionId',
		type: 'string',
		default: '',
		description: 'Open suggestion to comment on',
		displayOptions: { show: { resource: ['documentComment'], operation: ['create'] } },
	},
];

export async function execute(this: IExecuteFunctions, itemIndex: number): Promise<INodeExecutionData> {
	const documentId = this.getNodeParameter('documentId', itemIndex) as string;
	const content = this.getNodeParameter('content', itemIndex) as string;
	const quote = this.getNodeParameter('quote', itemIndex, '') as string;
	const quoteOccurrence = this.getNodeParameter('quoteOccurrence', itemIndex, 0) as number;
	const suggestionId = this.getNodeParameter('suggestionId', itemIndex, '') as string;
	if (!quote && !suggestionId) {
		throw new NodeOperationError(
			this.getNode(),
			'Provide a Quote or a Suggestion ID to anchor the comment',
			{ itemIndex },
		);
	}
	const input: Record<string, string | number> = { documentId, content };
	if (quote) {
		input.quote = quote;
		input.quoteOccurrence = quoteOccurrence;
	}
	if (suggestionId) {
		input.suggestionId = suggestionId;
	}

	const query = `
		mutation CreateDocumentCommentThread($input: CreateDocumentCommentThreadInput!) {
			createDocumentCommentThread(input: $input) {
				documentCommentThreadEdge { node { id anchorQuote suggestion { id } } }
			}
		}
	`;
	const responseData = await proboApiRequest.call(this, query, { input });
	return { json: responseData, pairedItem: { item: itemIndex } };
}
