import type { INodeProperties } from 'n8n-workflow';
import * as createOp from './create.operation';
import * as updateOp from './update.operation';
import * as deleteOp from './delete.operation';
import * as getOp from './get.operation';
import * as getAllOp from './getAll.operation';
import * as uploadReportOp from './uploadReport.operation';
import * as deleteReportOp from './deleteReport.operation';

export const description: INodeProperties[] = [
	{
		displayName: 'Operation',
		name: 'operation',
		type: 'options',
		noDataExpression: true,
		displayOptions: {
			show: {
				resource: ['audit'],
			},
		},
		options: [
			{
				name: 'Create',
				value: 'create',
				description: 'Create a new audit',
				action: 'Create an audit',
			},
			{
				name: 'Delete',
				value: 'delete',
				description: 'Delete an audit',
				action: 'Delete an audit',
			},
			{
				name: 'Delete File',
				value: 'deleteReport',
				description: 'Delete an audit file',
				action: 'Delete an audit file',
			},
			{
				name: 'Get',
				value: 'get',
				description: 'Get an audit',
				action: 'Get an audit',
			},
			{
				name: 'Get Many',
				value: 'getAll',
				description: 'Get many audits',
				action: 'Get many audits',
			},
			{
				name: 'Update',
				value: 'update',
				description: 'Update an existing audit',
				action: 'Update an audit',
			},
			{
				name: 'Upload Report File',
				value: 'uploadReport',
				description: 'Upload a report file for an audit',
				action: 'Upload an audit report file',
			},
		],
		default: 'create',
	},
	...createOp.description,
	...updateOp.description,
	...deleteOp.description,
	...getOp.description,
	...getAllOp.description,
	...uploadReportOp.description,
	...deleteReportOp.description,
];

export {
	createOp as create,
	updateOp as update,
	deleteOp as delete,
	getOp as get,
	getAllOp as getAll,
	uploadReportOp as uploadReport,
	deleteReportOp as deleteReport,
};
