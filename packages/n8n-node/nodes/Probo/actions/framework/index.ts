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
				resource: ['framework'],
			},
		},
		options: [
			{
				name: 'Create',
				value: 'create',
				description: 'Create a new framework',
				action: 'Create a framework',
			},
			{
				name: 'Delete',
				value: 'delete',
				description: 'Delete a framework',
				action: 'Delete a framework',
			},
			{
				name: 'Get',
				value: 'get',
				description: 'Get a framework',
				action: 'Get a framework',
			},
			{
				name: 'Get Many',
				value: 'getAll',
				description: 'Get many frameworks',
				action: 'Get many frameworks',
			},
			{
				name: 'Update',
				value: 'update',
				description: 'Update an existing framework',
				action: 'Update a framework',
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

