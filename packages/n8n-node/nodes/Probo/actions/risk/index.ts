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
				resource: ['risk'],
			},
		},
		options: [
			{
				name: 'Create',
				value: 'create',
				description: 'Create a new risk',
				action: 'Create a risk',
			},
			{
				name: 'Delete',
				value: 'delete',
				description: 'Delete a risk',
				action: 'Delete a risk',
			},
			{
				name: 'Get',
				value: 'get',
				description: 'Get a risk',
				action: 'Get a risk',
			},
			{
				name: 'Get Many',
				value: 'getAll',
				description: 'Get many risks',
				action: 'Get many risks',
			},
			{
				name: 'Link Document',
				value: 'linkDocument',
				description: 'Link a document to a risk',
				action: 'Link a document to a risk',
			},
			{
				name: 'Link Measure',
				value: 'linkMeasure',
				description: 'Link a measure to a risk',
				action: 'Link a measure to a risk',
			},
			{
				name: 'Link Obligation',
				value: 'linkObligation',
				description: 'Link an obligation to a risk',
				action: 'Link an obligation to a risk',
			},
			{
				name: 'Unlink Document',
				value: 'unlinkDocument',
				description: 'Unlink a document from a risk',
				action: 'Unlink a document from a risk',
			},
			{
				name: 'Unlink Measure',
				value: 'unlinkMeasure',
				description: 'Unlink a measure from a risk',
				action: 'Unlink a measure from a risk',
			},
			{
				name: 'Unlink Obligation',
				value: 'unlinkObligation',
				description: 'Unlink an obligation from a risk',
				action: 'Unlink an obligation from a risk',
			},
			{
				name: 'Update',
				value: 'update',
				description: 'Update an existing risk',
				action: 'Update a risk',
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
	linkObligationOp as linkObligation,
	unlinkObligationOp as unlinkObligation,
};
