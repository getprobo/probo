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

const show = { show: { resource: ['documentSuggestion'], operation: ['changeBlock'] } };

// These insert into an empty document when the quote is empty.
const emptyDocumentOperations = ['INSERT_BLOCK', 'INSERT_DIVIDER', 'INSERT_IMAGE', 'INSERT_FILE'];

const blockOperations = [
	{ name: 'Delete Block', value: 'DELETE_BLOCK' },
	{ name: 'Delete List Item', value: 'DELETE_LIST_ITEM' },
	{ name: 'Delete Mention', value: 'DELETE_MENTION' },
	{ name: 'Insert Block', value: 'INSERT_BLOCK' },
	{ name: 'Insert Divider', value: 'INSERT_DIVIDER' },
	{ name: 'Insert File', value: 'INSERT_FILE' },
	{ name: 'Insert Image', value: 'INSERT_IMAGE' },
	{ name: 'Insert List Item', value: 'INSERT_LIST_ITEM' },
	{ name: 'Insert Mention', value: 'INSERT_MENTION' },
	{ name: 'Nest List Item', value: 'NEST_LIST_ITEM' },
	{ name: 'Set Block', value: 'SET_BLOCK' },
	{ name: 'Set Task', value: 'SET_TASK' },
	{ name: 'Unnest List Item', value: 'UNNEST_LIST_ITEM' },
];

export const description: INodeProperties[] = [
	...changeModeFields('changeBlock'),
	{
		displayName: 'Block Operation',
		name: 'blockOperation',
		type: 'options',
		required: true,
		default: 'SET_BLOCK',
		options: blockOperations,
		displayOptions: show,
	},
	{
		displayName: 'Quote',
		name: 'quote',
		type: 'string',
		default: '',
		description: 'Text that locates the block. Empty inserts into an empty document.',
		displayOptions: show,
	},
	{
		displayName: 'Quote Occurrence',
		name: 'quoteOccurrence',
		type: 'number',
		typeOptions: { minValue: 0, numberPrecision: 0 },
		default: 0,
		displayOptions: show,
	},
	{
		displayName: 'Before',
		name: 'before',
		type: 'boolean',
		default: false,
		description: 'Whether to insert before the quote, or delete the neighboring textless block',
		displayOptions: show,
	},
	{
		displayName: 'Block',
		name: 'blockKind',
		type: 'options',
		default: 'PARAGRAPH',
		options: [
			{ name: 'Bullet List', value: 'BULLET_LIST' },
			{ name: 'Code Block', value: 'CODE_BLOCK' },
			{ name: 'Heading', value: 'HEADING' },
			{ name: 'Numbered List', value: 'ORDERED_LIST' },
			{ name: 'Paragraph', value: 'PARAGRAPH' },
			{ name: 'Quote', value: 'BLOCKQUOTE' },
			{ name: 'Task List', value: 'TASK_LIST' },
		],
		displayOptions: {
			show: {
				resource: ['documentSuggestion'],
				operation: ['changeBlock'],
				blockOperation: ['SET_BLOCK', 'INSERT_BLOCK', 'INSERT_LIST_ITEM'],
			},
		},
	},
	{
		displayName: 'Level',
		name: 'headingLevel',
		type: 'number',
		typeOptions: { minValue: 1, maxValue: 6 },
		default: 1,
		description: 'Heading level from 1 to 6',
		displayOptions: {
			show: {
				resource: ['documentSuggestion'],
				operation: ['changeBlock'],
				blockKind: ['HEADING'],
			},
		},
	},
	{
		displayName: 'Language',
		name: 'language',
		type: 'string',
		default: '',
		description: 'Code block language. mermaid draws a diagram.',
		displayOptions: {
			show: {
				resource: ['documentSuggestion'],
				operation: ['changeBlock'],
				blockKind: ['CODE_BLOCK'],
			},
		},
	},
	{
		displayName: 'Text',
		name: 'blockText',
		type: 'string',
		default: '',
		description: 'Text of the new block or list item',
		displayOptions: show,
	},
	{
		displayName: 'Items',
		name: 'items',
		type: 'json',
		default: '[]',
		description: 'JSON array of list item texts',
		displayOptions: {
			show: {
				resource: ['documentSuggestion'],
				operation: ['changeBlock'],
				blockOperation: ['INSERT_BLOCK'],
				blockKind: ['BULLET_LIST', 'ORDERED_LIST', 'TASK_LIST'],
			},
		},
	},
	{
		displayName: 'Checked',
		name: 'checked',
		type: 'boolean',
		default: false,
		displayOptions: {
			show: {
				resource: ['documentSuggestion'],
				operation: ['changeBlock'],
				blockOperation: ['SET_TASK'],
			},
		},
	},
	{
		displayName: 'File ID',
		name: 'fileId',
		type: 'string',
		default: '',
		description: 'File already uploaded in this organization',
		displayOptions: {
			show: {
				resource: ['documentSuggestion'],
				operation: ['changeBlock'],
				blockOperation: ['INSERT_IMAGE', 'INSERT_FILE'],
			},
		},
	},
	{
		displayName: 'Image URL',
		name: 'src',
		type: 'string',
		default: '',
		description: 'HTTP or HTTPS image address when no file is set',
		displayOptions: {
			show: {
				resource: ['documentSuggestion'],
				operation: ['changeBlock'],
				blockOperation: ['INSERT_IMAGE'],
			},
		},
	},
	{
		displayName: 'Alt Text',
		name: 'alt',
		type: 'string',
		default: '',
		displayOptions: {
			show: {
				resource: ['documentSuggestion'],
				operation: ['changeBlock'],
				blockOperation: ['INSERT_IMAGE'],
			},
		},
	},
	{
		displayName: 'Width',
		name: 'width',
		type: 'number',
		typeOptions: { minValue: 1, maxValue: 100 },
		default: 100,
		description: 'Image width as a percentage from 1 to 100',
		displayOptions: {
			show: {
				resource: ['documentSuggestion'],
				operation: ['changeBlock'],
				blockOperation: ['INSERT_IMAGE'],
			},
		},
	},
	{
		displayName: 'Profile ID',
		name: 'profileId',
		type: 'string',
		default: '',
		description: 'Membership profile to mention. The server fills the label.',
		displayOptions: {
			show: {
				resource: ['documentSuggestion'],
				operation: ['changeBlock'],
				blockOperation: ['INSERT_MENTION'],
			},
		},
	},
];

export async function execute(this: IExecuteFunctions, itemIndex: number): Promise<INodeExecutionData> {
	const operation = this.getNodeParameter('blockOperation', itemIndex) as string;
	const quote = this.getNodeParameter('quote', itemIndex, '') as string;
	const quoteOccurrence = changeOccurrence.call(this, itemIndex);
	if (quote.trim() === '' && !emptyDocumentOperations.includes(operation)) {
		throw new NodeOperationError(this.getNode(), 'Quote is required for this operation', { itemIndex });
	}
	const input: Record<string, unknown> = {
		...changeTarget.call(this, itemIndex),
		operation,
		quote,
		quoteOccurrence,
	};

	// An omitted before deletes the quoted block; a set one deletes its neighbor.
	if (this.getNodeParameter('before', itemIndex, false) as boolean) {
		input.before = true;
	}

	if (['SET_BLOCK', 'INSERT_BLOCK', 'INSERT_LIST_ITEM'].includes(operation)) {
		input.block = this.getNodeParameter('blockKind', itemIndex) as string;
	}

	if (input.block === 'HEADING') {
		input.level = this.getNodeParameter('headingLevel', itemIndex, 1) as number;
	}

	if (input.block === 'CODE_BLOCK') {
		const language = this.getNodeParameter('language', itemIndex, '') as string;
		if (language !== '') {
			input.language = language;
		}
	}

	const text = this.getNodeParameter('blockText', itemIndex, '') as string;
	if (text !== '') {
		input.text = text;
	}

	if (operation === 'INSERT_BLOCK' && ['BULLET_LIST', 'ORDERED_LIST', 'TASK_LIST'].includes(String(input.block))) {
		const raw = this.getNodeParameter('items', itemIndex, '[]');
		let parsed: unknown = raw;
		if (typeof raw === 'string') {
			try {
				parsed = JSON.parse(raw) as unknown;
			} catch {
				throw new NodeOperationError(this.getNode(), 'Items must be valid JSON', { itemIndex });
			}
		}
		if (!Array.isArray(parsed) || parsed.some((item) => typeof item !== 'string')) {
			throw new NodeOperationError(this.getNode(), 'Items must be a JSON array of strings', { itemIndex });
		}
		input.items = parsed;
	}

	if (operation === 'SET_TASK') {
		input.checked = this.getNodeParameter('checked', itemIndex, false) as boolean;
	}

	if (operation === 'INSERT_IMAGE' || operation === 'INSERT_FILE') {
		const fileId = this.getNodeParameter('fileId', itemIndex, '') as string;
		if (fileId !== '') {
			input.fileId = fileId;
		}
	}

	if (operation === 'INSERT_IMAGE') {
		const src = this.getNodeParameter('src', itemIndex, '') as string;
		const alt = this.getNodeParameter('alt', itemIndex, '') as string;
		if (src !== '') {
			input.src = src;
		}
		if (alt !== '') {
			input.alt = alt;
		}
		input.width = this.getNodeParameter('width', itemIndex, 100) as number;
	}

	if (operation === 'INSERT_MENTION') {
		const profileId = this.getNodeParameter('profileId', itemIndex, '') as string;
		if (profileId === '') {
			throw new NodeOperationError(this.getNode(), 'Profile ID is required', { itemIndex });
		}
		input.profileId = profileId;
	}

	const query = `
		mutation ChangeDocumentBlock($input: ChangeDocumentBlockInput!) {
			changeDocumentBlock(input: $input) { ${changeSelection} }
		}
	`;
	const responseData = await proboApiRequest.call(this, query, { input });

	return { json: responseData, pairedItem: { item: itemIndex } };
}
