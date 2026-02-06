import type { INodeProperties } from 'n8n-workflow';
import * as inviteOp from './invite.operation';
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
				resource: ['member'],
			},
		},
		options: [
			{
				name: 'Delete',
				value: 'delete',
				description: 'Remove a member from an organization',
				action: 'Delete a member',
			},
			{
				name: 'Get',
				value: 'get',
				description: 'Get a member (membership) by ID',
				action: 'Get a member',
			},
			{
				name: 'Get Many',
				value: 'getAll',
				description: 'Get many members of an organization',
				action: 'Get many members',
			},
			{
				name: 'Invite',
				value: 'invite',
				description: 'Invite a new member to an organization',
				action: 'Invite a member',
			},
			{
				name: 'Update',
				value: 'update',
				description: 'Update a member\'s role',
				action: 'Update a member',
			},
		],
		default: 'invite',
	},
	...inviteOp.description,
	...updateOp.description,
	...deleteOp.description,
	...getOp.description,
	...getAllOp.description,
];

export { inviteOp as invite, updateOp as update, deleteOp as delete, getOp as get, getAllOp as getAll };
