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

const addFormatOptions = { show: { resource: ['documentSuggestion'], operation: ['addFormat'] } };

export const description: INodeProperties[] = [
	{
		displayName: 'Suggestion ID',
		name: 'suggestionId',
		type: 'string',
		required: true,
		default: '',
		displayOptions: addFormatOptions,
	},
	{
		displayName: 'Mark',
		name: 'formatMark',
		type: 'options',
		required: true,
		default: 'BOLD',
		options: [
			{ name: 'Bold', value: 'BOLD' },
			{ name: 'Code', value: 'CODE' },
			{ name: 'Italic', value: 'ITALIC' },
			{ name: 'Link', value: 'LINK' },
			{ name: 'Strike', value: 'STRIKE' },
			{ name: 'Underline', value: 'UNDERLINE' },
		],
		displayOptions: addFormatOptions,
	},
	{
		displayName: 'Action',
		name: 'formatAction',
		type: 'options',
		required: true,
		default: 'ADD',
		options: [
			{ name: 'Add', value: 'ADD' },
			{ name: 'Remove', value: 'REMOVE' },
		],
		displayOptions: addFormatOptions,
	},
	{
		displayName: 'Quote',
		name: 'quote',
		type: 'string',
		required: true,
		default: '',
		description: 'Text to format. It must not overlap another suggestion.',
		displayOptions: addFormatOptions,
	},
	{
		displayName: 'Quote Occurrence',
		name: 'quoteOccurrence',
		type: 'number',
		typeOptions: { minValue: 0 },
		default: 0,
		description: 'Zero-based occurrence of the quote',
		displayOptions: addFormatOptions,
	},
	{
		displayName: 'Link Address',
		name: 'href',
		type: 'string',
		default: '',
		description: 'Address of the suggested link',
		displayOptions: {
			show: {
				resource: ['documentSuggestion'],
				operation: ['addFormat'],
				formatMark: ['LINK'],
				formatAction: ['ADD'],
			},
		},
	},
];

export async function execute(this: IExecuteFunctions, itemIndex: number): Promise<INodeExecutionData> {
	const suggestionId = this.getNodeParameter('suggestionId', itemIndex) as string;
	const mark = this.getNodeParameter('formatMark', itemIndex) as string;
	const action = this.getNodeParameter('formatAction', itemIndex) as string;
	const quote = this.getNodeParameter('quote', itemIndex) as string;
	const quoteOccurrence = this.getNodeParameter('quoteOccurrence', itemIndex, 0) as number;
	if (quote === '') {
		throw new NodeOperationError(this.getNode(), 'Quote is required', { itemIndex });
	}

	const input: Record<string, unknown> = { suggestionId, mark, action, quote, quoteOccurrence };
	if (mark === 'LINK' && action === 'ADD') {
		const href = (this.getNodeParameter('href', itemIndex, '') as string).trim();
		if (href === '') {
			throw new NodeOperationError(this.getNode(), 'Link address is required', { itemIndex });
		}
		input.href = href;
	}

	const query = `
		mutation AddDocumentSuggestionFormatChange($input: AddDocumentSuggestionFormatChangeInput!) {
			addDocumentSuggestionFormatChange(input: $input) { documentSuggestion { id status } }
		}
	`;
	const responseData = await proboApiRequest.call(this, query, { input });
	return { json: responseData, pairedItem: { item: itemIndex } };
}
