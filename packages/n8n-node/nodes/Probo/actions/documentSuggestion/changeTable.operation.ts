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
import { changeModeFields, changeSelection, changeTarget } from './changeTarget';

const show = { resource: ['documentSuggestion'], operation: ['changeTable'] };

export const description: INodeProperties[] = [
	...changeModeFields('changeTable'),
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
		displayOptions: { show },
	},
	{
		displayName: 'Quote',
		name: 'quote',
		type: 'string',
		required: true,
		default: '',
		description: 'Text inside a cell of the row, column, or table, or the block a new table goes beside',
		displayOptions: { show },
	},
	{
		displayName: 'Quote Occurrence',
		name: 'quoteOccurrence',
		type: 'number',
		typeOptions: { minValue: 0 },
		default: 0,
		displayOptions: { show },
	},
	{
		displayName: 'Before',
		name: 'before',
		type: 'boolean',
		default: false,
		description: 'Whether to insert before the quoted row, column, or block',
		displayOptions: { show: { ...show, tableOperation: ['INSERT_ROW', 'INSERT_COLUMN', 'INSERT_TABLE'] } },
	},
	{
		displayName: 'Cells',
		name: 'cells',
		type: 'json',
		default: '[]',
		description: 'JSON array of cell texts for the new row or column',
		displayOptions: { show: { ...show, tableOperation: ['INSERT_ROW', 'INSERT_COLUMN'] } },
	},
	{
		displayName: 'Rows',
		name: 'rows',
		type: 'json',
		default: '[["", ""], ["", ""]]',
		description: 'JSON array of rows for the new table, each a JSON array of cell texts',
		displayOptions: { show: { ...show, tableOperation: ['INSERT_TABLE'] } },
	},
	{
		displayName: 'Header Row',
		name: 'headerRow',
		type: 'boolean',
		default: false,
		displayOptions: { show: { ...show, tableOperation: ['INSERT_TABLE'] } },
	},
];

export async function execute(this: IExecuteFunctions, itemIndex: number): Promise<INodeExecutionData> {
	const operation = this.getNodeParameter('tableOperation', itemIndex) as string;
	const quote = this.getNodeParameter('quote', itemIndex) as string;
	const quoteOccurrence = this.getNodeParameter('quoteOccurrence', itemIndex, 0) as number;
	if (quote === '') {
		throw new NodeOperationError(this.getNode(), 'Quote is required', { itemIndex });
	}

	const input: Record<string, unknown> = {
		...changeTarget.call(this, itemIndex),
		operation,
		quote,
		quoteOccurrence,
	};

	if (['INSERT_ROW', 'INSERT_COLUMN', 'INSERT_TABLE'].includes(operation)) {
		input.before = this.getNodeParameter('before', itemIndex, false) as boolean;
	}

	if (operation === 'INSERT_ROW' || operation === 'INSERT_COLUMN') {
		input.cells = stringMatrix.call(this, itemIndex, 'cells', false);
	}

	if (operation === 'INSERT_TABLE') {
		input.rows = stringMatrix.call(this, itemIndex, 'rows', true);
		input.headerRow = this.getNodeParameter('headerRow', itemIndex, false) as boolean;
	}

	const query = `
		mutation ChangeDocumentTable($input: ChangeDocumentTableInput!) {
			changeDocumentTable(input: $input) { ${changeSelection} }
		}
	`;
	const responseData = await proboApiRequest.call(this, query, { input });

	return { json: responseData, pairedItem: { item: itemIndex } };
}

function stringMatrix(this: IExecuteFunctions, itemIndex: number, field: string, nested: boolean): unknown {
	const raw = this.getNodeParameter(field, itemIndex, '[]');
	let parsed: unknown = raw;
	if (typeof raw === 'string') {
		try {
			parsed = JSON.parse(raw) as unknown;
		} catch {
			throw new NodeOperationError(this.getNode(), `${field} must be valid JSON`, { itemIndex });
		}
	}
	if (!Array.isArray(parsed)) {
		throw new NodeOperationError(this.getNode(), `${field} must be a JSON array`, { itemIndex });
	}

	if (!nested) {
		return parsed.map((cell) => {
			if (typeof cell !== 'string') {
				throw new NodeOperationError(this.getNode(), `${field} entries must be strings`, { itemIndex });
			}
			return cell;
		});
	}

	return parsed.map((row) => {
		if (!Array.isArray(row)) {
			throw new NodeOperationError(this.getNode(), 'Each row must be a JSON array', { itemIndex });
		}
		return row.map((cell) => {
			if (typeof cell !== 'string') {
				throw new NodeOperationError(this.getNode(), 'Row cells must be strings', { itemIndex });
			}
			return cell;
		});
	});
}
