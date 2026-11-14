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

import { NodeOperationError, type IExecuteFunctions, type INodeExecutionData, type INodeProperties } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

const startOptions = { show: { resource: ['documentSuggestion'], operation: ['start'] } };

export const description: INodeProperties[] = [
	{
		displayName: 'Document ID',
		name: 'documentId',
		type: 'string',
		required: true,
		default: '',
		displayOptions: startOptions,
	},
	{
		displayName: 'Quote',
		name: 'quote',
		type: 'string',
		default: '',
		description: 'Text to replace. Leave empty to only allocate an ID.',
		displayOptions: startOptions,
	},
	{
		displayName: 'Replacement',
		name: 'replacement',
		type: 'string',
		default: '',
		description: 'Suggested text. Empty deletes the quote.',
		displayOptions: startOptions,
	},
	{
		displayName: 'Comment',
		name: 'comment',
		type: 'string',
		typeOptions: { rows: 4 },
		default: '',
		description: 'Plain-text note attached to the redline',
		displayOptions: startOptions,
	},
	{
		displayName: 'Quote Occurrence',
		name: 'quoteOccurrence',
		type: 'number',
		typeOptions: { minValue: 0 },
		default: 0,
		description: 'Zero-based occurrence of the quote',
		displayOptions: startOptions,
	},
];

export async function execute(this: IExecuteFunctions, itemIndex: number): Promise<INodeExecutionData> {
	const documentId = this.getNodeParameter('documentId', itemIndex) as string;
	const quote = this.getNodeParameter('quote', itemIndex, '') as string;
	const replacement = this.getNodeParameter('replacement', itemIndex, '') as string;
	const comment = this.getNodeParameter('comment', itemIndex, '') as string;
	const quoteOccurrence = this.getNodeParameter('quoteOccurrence', itemIndex, 0) as number;
	if (quote === '' && (comment !== '' || quoteOccurrence !== 0)) {
		throw new NodeOperationError(this.getNode(), 'Quote is required when Replacement, Comment, or Quote Occurrence is set');
	}
	const input: Record<string, unknown> = { documentId };
	if (quote !== '') {
		input.quote = quote;
		input.replacement = replacement;
	}
	if (comment !== '') {
		input.comment = comment;
	}
	if (quoteOccurrence !== 0) {
		input.quoteOccurrence = quoteOccurrence;
	}
	const query = `
		mutation StartDocumentSuggestion($input: StartDocumentSuggestionInput!) {
			startDocumentSuggestion(input: $input) { documentSuggestionEdge { node { id status } } }
		}
	`;
	const responseData = await proboApiRequest.call(this, query, { input });
	return { json: responseData, pairedItem: { item: itemIndex } };
}
