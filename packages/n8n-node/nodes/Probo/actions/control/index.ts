// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
import * as linkSnapshotOp from './linkSnapshot.operation';
import * as unlinkSnapshotOp from './unlinkSnapshot.operation';

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
				name: 'Link Snapshot',
				value: 'linkSnapshot',
				description: 'Link a snapshot to a control',
				action: 'Link a snapshot to a control',
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
				name: 'Unlink Snapshot',
				value: 'unlinkSnapshot',
				description: 'Unlink a snapshot from a control',
				action: 'Unlink a snapshot from a control',
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
	...linkSnapshotOp.description,
	...unlinkSnapshotOp.description,
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
	linkSnapshotOp as linkSnapshot,
	unlinkSnapshotOp as unlinkSnapshot,
};
