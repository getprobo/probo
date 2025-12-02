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
				resource: ['asset'],
			},
		},
		options: [
			{
				name: 'Create',
				value: 'create',
				description: 'Create a new asset',
				action: 'Create an asset',
			},
			{
				name: 'Delete',
				value: 'delete',
				description: 'Delete an asset',
				action: 'Delete an asset',
			},
			{
				name: 'Get',
				value: 'get',
				description: 'Get an asset',
				action: 'Get an asset',
			},
			{
				name: 'Get Many',
				value: 'getAll',
				description: 'Get many assets',
				action: 'Get many assets',
			},
			{
				name: 'Update',
				value: 'update',
				description: 'Update an existing asset',
				action: 'Update an asset',
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

