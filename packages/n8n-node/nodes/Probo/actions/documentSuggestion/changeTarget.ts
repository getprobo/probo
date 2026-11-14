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

import { NodeOperationError, type IExecuteFunctions, type INodeProperties } from 'n8n-workflow';

export function changeModeFields(operation: string): INodeProperties[] {
	const show = { resource: ['documentSuggestion'], operation: [operation] };

	return [
		{
			displayName: 'Mode',
			name: 'changeMode',
			type: 'options',
			required: true,
			default: 'APPLY',
			options: [
				{ name: 'Apply', value: 'APPLY', description: 'Write the change into the document' },
				{ name: 'Suggest', value: 'SUGGEST', description: 'Store the change on an open suggestion' },
			],
			displayOptions: { show },
		},
		{
			displayName: 'Document ID',
			name: 'documentId',
			type: 'string',
			required: true,
			default: '',
			description: 'Document to write the change into',
			displayOptions: { show: { ...show, changeMode: ['APPLY'] } },
		},
		{
			displayName: 'Suggestion ID',
			name: 'suggestionId',
			type: 'string',
			required: true,
			default: '',
			description: 'Open suggestion that stores the change',
			displayOptions: { show: { ...show, changeMode: ['SUGGEST'] } },
		},
	];
}

export function changeTarget(this: IExecuteFunctions, itemIndex: number): Record<string, unknown> {
	const mode = this.getNodeParameter('changeMode', itemIndex) as string;
	const input: Record<string, unknown> = { mode };

	if (mode === 'APPLY') {
		const documentId = this.getNodeParameter('documentId', itemIndex, '') as string;
		if (documentId === '') {
			throw new NodeOperationError(this.getNode(), 'Document ID is required when Mode is Apply', { itemIndex });
		}
		input.documentId = documentId;
	}

	if (mode === 'SUGGEST') {
		const suggestionId = this.getNodeParameter('suggestionId', itemIndex, '') as string;
		if (suggestionId === '') {
			throw new NodeOperationError(this.getNode(), 'Suggestion ID is required when Mode is Suggest', { itemIndex });
		}
		input.suggestionId = suggestionId;
	}

	return input;
}

export function changeOccurrence(this: IExecuteFunctions, itemIndex: number): number {
	const occurrence = this.getNodeParameter('quoteOccurrence', itemIndex, 0) as number;
	if (!Number.isInteger(occurrence) || occurrence < 0) {
		throw new NodeOperationError(this.getNode(), 'Quote Occurrence must be a whole number of 0 or more', { itemIndex });
	}

	return occurrence;
}

export const changeSelection = 'document { id } documentSuggestion { id status }';
