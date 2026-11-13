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

const addTableOptions = { show: { resource: ['documentSuggestion'], operation: ['addTable'] } };

export const description: INodeProperties[] = [
	{
		displayName: 'Suggestion ID',
		name: 'suggestionId',
		type: 'string',
		required: true,
		default: '',
		displayOptions: addTableOptions,
	},
	{
		displayName: 'Table Operation',
		name: 'tableOperation',
		type: 'options',
		required: true,
		default: 'INSERT_ROW',
		options: [
			{ name: 'Delete Column', value: 'DELETE_COLUMN' },
			{ name: 'Delete Row', value: 'DELETE_ROW' },
			{ name: 'Delete Table', value: 'DELETE_TABLE' },
			{ name: 'Insert Column', value: 'INSERT_COLUMN' },
			{ name: 'Insert Row', value: 'INSERT_ROW' },
			{ name: 'Insert Table', value: 'INSERT_TABLE' },
		],
		displayOptions: addTableOptions,
	},
	{
		displayName: 'Quote',
		name: 'quote',
		type: 'string',
		required: true,
		default: '',
		description:
			'Text inside a cell of the row, column, or table to change. For Insert Table, text of the top-level block the new table goes beside.',
		displayOptions: addTableOptions,
	},
	{
		displayName: 'Quote Occurrence',
		name: 'quoteOccurrence',
		type: 'number',
		typeOptions: { minValue: 0 },
		default: 0,
		description: 'Zero-based occurrence of the quote',
		displayOptions: addTableOptions,
	},
	{
		displayName: 'Before',
		name: 'before',
		type: 'boolean',
		default: false,
		description: 'Whether to insert before the quoted row, column, or block instead of after it',
		displayOptions: {
			show: {
				resource: ['documentSuggestion'],
				operation: ['addTable'],
				tableOperation: ['INSERT_ROW', 'INSERT_COLUMN', 'INSERT_TABLE'],
			},
		},
	},
	{
		displayName: 'Cells',
		name: 'cells',
		type: 'json',
		default: '[]',
		description:
			'JSON array of cell texts: the new row left to right, or the new column top to bottom. An empty array leaves the cells empty.',
		displayOptions: {
			show: {
				resource: ['documentSuggestion'],
				operation: ['addTable'],
				tableOperation: ['INSERT_ROW', 'INSERT_COLUMN'],
			},
		},
	},
	{
		displayName: 'Rows',
		name: 'rows',
		type: 'json',
		default: '[["", ""], ["", ""]]',
		description: 'JSON array of rows for the new table, each a JSON array of cell texts with the same length',
		displayOptions: {
			show: {
				resource: ['documentSuggestion'],
				operation: ['addTable'],
				tableOperation: ['INSERT_TABLE'],
			},
		},
	},
	{
		displayName: 'Header Row',
		name: 'headerRow',
		type: 'boolean',
		default: false,
		description: "Whether to make the new table's first row a header row",
		displayOptions: {
			show: {
				resource: ['documentSuggestion'],
				operation: ['addTable'],
				tableOperation: ['INSERT_TABLE'],
			},
		},
	},
];

export async function execute(this: IExecuteFunctions, itemIndex: number): Promise<INodeExecutionData> {
	const suggestionId = this.getNodeParameter('suggestionId', itemIndex) as string;
	const operation = this.getNodeParameter('tableOperation', itemIndex) as string;
	const quote = this.getNodeParameter('quote', itemIndex) as string;
	const quoteOccurrence = this.getNodeParameter('quoteOccurrence', itemIndex, 0) as number;
	if (quote === '') {
		throw new NodeOperationError(this.getNode(), 'Quote is required', { itemIndex });
	}

	const input: Record<string, unknown> = { suggestionId, operation, quote, quoteOccurrence };

	if (['INSERT_ROW', 'INSERT_COLUMN', 'INSERT_TABLE'].includes(operation)) {
		input.before = this.getNodeParameter('before', itemIndex, false) as boolean;
	}

	if (operation === 'INSERT_ROW' || operation === 'INSERT_COLUMN') {
		const cells = jsonArray.call(this, itemIndex, this.getNodeParameter('cells', itemIndex, '[]'), 'Cells');
		if (cells.length > 0) {
			input.cells = cells.map(cell => stringCell.call(this, itemIndex, cell, 'Cells'));
		}
	}

	if (operation === 'INSERT_TABLE') {
		const rows = jsonArray.call(this, itemIndex, this.getNodeParameter('rows', itemIndex, '[]'), 'Rows');
		if (rows.length === 0) {
			throw new NodeOperationError(this.getNode(), 'Rows must hold at least one row', { itemIndex });
		}
		input.rows = rows.map((row) => {
			if (!Array.isArray(row)) {
				throw new NodeOperationError(this.getNode(), 'Each row must be a JSON array', { itemIndex });
			}
			return row.map(cell => stringCell.call(this, itemIndex, cell, 'Rows'));
		});
		input.headerRow = this.getNodeParameter('headerRow', itemIndex, false) as boolean;
	}

	const query = `
		mutation AddDocumentSuggestionTableChange($input: AddDocumentSuggestionTableChangeInput!) {
			addDocumentSuggestionTableChange(input: $input) { documentSuggestion { id status } }
		}
	`;
	const responseData = await proboApiRequest.call(this, query, { input });
	return { json: responseData, pairedItem: { item: itemIndex } };
}

function jsonArray(this: IExecuteFunctions, itemIndex: number, value: unknown, field: string): unknown[] {
	let parsed = value;
	if (typeof value === 'string') {
		if (value.trim() === '') {
			return [];
		}
		try {
			parsed = JSON.parse(value) as unknown;
		} catch {
			throw new NodeOperationError(this.getNode(), `Invalid JSON in ${field}`, { itemIndex });
		}
	}
	if (parsed === undefined || parsed === null) {
		return [];
	}
	if (!Array.isArray(parsed)) {
		throw new NodeOperationError(this.getNode(), `${field} must be a JSON array`, { itemIndex });
	}
	return parsed;
}

function stringCell(this: IExecuteFunctions, itemIndex: number, value: unknown, field: string): string {
	if (typeof value !== 'string') {
		throw new NodeOperationError(this.getNode(), `${field} cells must be strings`, { itemIndex });
	}
	return value;
}
