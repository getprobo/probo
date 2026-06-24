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

import type { INodeProperties } from 'n8n-workflow';
import * as createOp from './create.operation';
import * as updateOp from './update.operation';
import * as deleteOp from './delete.operation';
import * as getOp from './get.operation';
import * as getAllOp from './getAll.operation';
import * as linkMeasureOp from './linkMeasure.operation';
import * as unlinkMeasureOp from './unlinkMeasure.operation';
import * as linkDocumentOp from './linkDocument.operation';
import * as unlinkDocumentOp from './unlinkDocument.operation';
import * as linkAuditOp from './linkAudit.operation';
import * as unlinkAuditOp from './unlinkAudit.operation';
import * as linkObligationOp from './linkObligation.operation';
import * as unlinkObligationOp from './unlinkObligation.operation';

export const description: INodeProperties[] = [
	{
		displayName: 'Operation',
		name: 'operation',
		type: 'options',
		noDataExpression: true,
		displayOptions: {
			show: {
				resource: ['control'],
			},
		},
		options: [
			{
				name: 'Create',
				value: 'create',
				description: 'Create a new control',
				action: 'Create a control',
			},
			{
				name: 'Delete',
				value: 'delete',
				description: 'Delete a control',
				action: 'Delete a control',
			},
			{
				name: 'Get',
				value: 'get',
				description: 'Get a control',
				action: 'Get a control',
			},
			{
				name: 'Get Many',
				value: 'getAll',
				description: 'Get many controls',
				action: 'Get many controls',
			},
			{
				name: 'Link Audit',
				value: 'linkAudit',
				description: 'Link an audit to a control',
				action: 'Link an audit to a control',
			},
			{
				name: 'Link Document',
				value: 'linkDocument',
				description: 'Link a document to a control',
				action: 'Link a document to a control',
			},
			{
				name: 'Link Measure',
				value: 'linkMeasure',
				description: 'Link a measure to a control',
				action: 'Link a measure to a control',
			},
			{
				name: 'Link Obligation',
				value: 'linkObligation',
				description: 'Link an obligation to a control',
				action: 'Link an obligation to a control',
			},
			{
				name: 'Unlink Audit',
				value: 'unlinkAudit',
				description: 'Unlink an audit from a control',
				action: 'Unlink an audit from a control',
			},
			{
				name: 'Unlink Document',
				value: 'unlinkDocument',
				description: 'Unlink a document from a control',
				action: 'Unlink a document from a control',
			},
			{
				name: 'Unlink Measure',
				value: 'unlinkMeasure',
				description: 'Unlink a measure from a control',
				action: 'Unlink a measure from a control',
			},
			{
				name: 'Unlink Obligation',
				value: 'unlinkObligation',
				description: 'Unlink an obligation from a control',
				action: 'Unlink an obligation from a control',
			},
			{
				name: 'Update',
				value: 'update',
				description: 'Update an existing control',
				action: 'Update a control',
			},
		],
		default: 'create',
	},
	...createOp.description,
	...updateOp.description,
	...deleteOp.description,
	...getOp.description,
	...getAllOp.description,
	...linkMeasureOp.description,
	...unlinkMeasureOp.description,
	...linkDocumentOp.description,
	...unlinkDocumentOp.description,
	...linkAuditOp.description,
	...unlinkAuditOp.description,
	...linkObligationOp.description,
	...unlinkObligationOp.description,
];

export {
	createOp as create,
	updateOp as update,
	deleteOp as delete,
	getOp as get,
	getAllOp as getAll,
	linkMeasureOp as linkMeasure,
	unlinkMeasureOp as unlinkMeasure,
	linkDocumentOp as linkDocument,
	unlinkDocumentOp as unlinkDocument,
	linkAuditOp as linkAudit,
	unlinkAuditOp as unlinkAudit,
	linkObligationOp as linkObligation,
	unlinkObligationOp as unlinkObligation,
};
