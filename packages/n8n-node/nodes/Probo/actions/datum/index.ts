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
				resource: ['datum'],
			},
		},
		options: [
			{
				name: 'Create',
				value: 'create',
				description: 'Create a new datum',
				action: 'Create a datum',
			},
			{
				name: 'Delete',
				value: 'delete',
				description: 'Delete a datum',
				action: 'Delete a datum',
			},
			{
				name: 'Get',
				value: 'get',
				description: 'Get a datum',
				action: 'Get a datum',
			},
			{
				name: 'Get Many',
				value: 'getAll',
				description: 'Get many data',
				action: 'Get many data',
			},
			{
				name: 'Update',
				value: 'update',
				description: 'Update an existing datum',
				action: 'Update a datum',
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

