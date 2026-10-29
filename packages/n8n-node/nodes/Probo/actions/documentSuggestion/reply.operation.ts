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
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Document ID',
		name: 'documentId',
		type: 'string',
		required: true,
		default: '',
		displayOptions: { show: { resource: ['documentSuggestion'], operation: ['reply'] } },
	},
	{
		displayName: 'Suggestion ID',
		name: 'suggestionId',
		type: 'string',
		required: true,
		default: '',
		displayOptions: { show: { resource: ['documentSuggestion'], operation: ['reply'] } },
	},
	{
		displayName: 'Content',
		name: 'content',
		type: 'string',
		typeOptions: { rows: 4 },
		required: true,
		default: '',
		displayOptions: { show: { resource: ['documentSuggestion'], operation: ['reply'] } },
	},
];

export async function execute(this: IExecuteFunctions, itemIndex: number): Promise<INodeExecutionData> {
	const documentId = this.getNodeParameter('documentId', itemIndex) as string;
	const suggestionId = this.getNodeParameter('suggestionId', itemIndex) as string;
	const content = this.getNodeParameter('content', itemIndex) as string;
	const query = `
		mutation ReplyDocumentSuggestion($input: CreateDocumentCommentThreadInput!) {
			createDocumentCommentThread(input: $input) {
				documentCommentThreadEdge { node { id suggestion { id } } }
			}
		}
	`;
	const responseData = await proboApiRequest.call(this, query, {
		input: { documentId, suggestionId, content },
	});
	return { json: responseData, pairedItem: { item: itemIndex } };
}
