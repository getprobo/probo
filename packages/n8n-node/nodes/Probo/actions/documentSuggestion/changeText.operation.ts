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
import { changeModeFields, changeOccurrence, changeSelection, changeTarget } from './changeTarget';

const show = { show: { resource: ['documentSuggestion'], operation: ['changeText'] } };

export const description: INodeProperties[] = [
	...changeModeFields('changeText'),
	{
		displayName: 'Quote',
		name: 'quote',
		type: 'string',
		required: true,
		default: '',
		description: 'Text to replace. It must not overlap an existing suggestion.',
		displayOptions: show,
	},
	{
		displayName: 'Replacement',
		name: 'replacement',
		type: 'string',
		default: '',
		description: 'New text. Empty deletes the quote. A newline inside the quote becomes a line break.',
		displayOptions: show,
	},
	{
		displayName: 'Quote Occurrence',
		name: 'quoteOccurrence',
		type: 'number',
		typeOptions: { minValue: 0, numberPrecision: 0 },
		default: 0,
		description: 'Zero-based occurrence of the quote',
		displayOptions: show,
	},
];

export async function execute(this: IExecuteFunctions, itemIndex: number): Promise<INodeExecutionData> {
	const quote = this.getNodeParameter('quote', itemIndex) as string;
	const replacement = this.getNodeParameter('replacement', itemIndex, '') as string;
	const quoteOccurrence = changeOccurrence.call(this, itemIndex);
	if (quote === '') {
		throw new NodeOperationError(this.getNode(), 'Quote is required', { itemIndex });
	}

	const query = `
		mutation ChangeDocumentText($input: ChangeDocumentTextInput!) {
			changeDocumentText(input: $input) { ${changeSelection} }
		}
	`;
	const responseData = await proboApiRequest.call(this, query, {
		input: { ...changeTarget.call(this, itemIndex), quote, replacement, quoteOccurrence },
	});

	return { json: responseData, pairedItem: { item: itemIndex } };
}
