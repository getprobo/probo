import type { INodeProperties } from 'n8n-workflow';
import * as createOp from './create.operation';
import * as updateOp from './update.operation';
import * as deleteOp from './delete.operation';
import * as getOp from './get.operation';
import * as getAllOp from './getAll.operation';

export const description: INodeProperties[] = [
	{
		displayName: 'Operation',
		name: 'operation',
		type: 'options',
		noDataExpression: true,
		displayOptions: {
			show: {
				resource: ['measure'],
			},
		},
		options: [
			{
				name: 'Create',
				value: 'create',
				description: 'Create a new measure',
				action: 'Create a measure',
			},
			{
				name: 'Delete',
				value: 'delete',
				description: 'Delete a measure',
				action: 'Delete a measure',
			},
			{
				name: 'Get',
				value: 'get',
				description: 'Get a measure',
				action: 'Get a measure',
			},
			{
				name: 'Get Many',
				value: 'getAll',
				description: 'Get many measures',
				action: 'Get many measures',
			},
			{
				name: 'Update',
				value: 'update',
				description: 'Update an existing measure',
				action: 'Update a measure',
			},
		],
		default: 'create',
	},
	...createOp.description,
	...updateOp.description,
	...deleteOp.description,
	...getOp.description,
	...getAllOp.description,
];

export { createOp as create, updateOp as update, deleteOp as delete, getOp as get, getAllOp as getAll };

