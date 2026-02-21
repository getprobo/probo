import type { INodeProperties } from 'n8n-workflow';
import * as listUsersOp from './listUsers.operation';
import * as getUserOp from './getUser.operation';
import * as createUserOp from './createUser.operation';
import * as inviteUserOp from './inviteUser.operation';
import * as updateUserOp from './updateUser.operation';
import * as updateMembershipOp from './updateMembership.operation';
import * as removeUserOp from './removeUser.operation';

export const description: INodeProperties[] = [
	{
		displayName: 'Operation',
		name: 'operation',
		type: 'options',
		noDataExpression: true,
		displayOptions: {
			show: {
				resource: ['user'],
			},
		},
		options: [
			{
				name: 'Create',
				value: 'createUser',
				description: 'Create a new user in the organization',
				action: 'Create a user',
			},
			{
				name: 'Get',
				value: 'getUser',
				description: 'Get a user (profile) by ID',
				action: 'Get a user',
			},
			{
				name: 'Invite',
				value: 'inviteUser',
				description: 'Invite a user to the organization',
				action: 'Invite a user',
			},
			{
				name: 'List',
				value: 'listUsers',
				description: 'List all users in the organization',
				action: 'List users',
			},
			{
				name: 'Remove',
				value: 'removeUser',
				description: 'Remove a user from the organization',
				action: 'Remove a user',
			},
			{
				name: 'Update',
				value: 'updateUser',
				description: 'Update a user (profile)',
				action: 'Update a user',
			},
			{
				name: 'Update Membership',
				value: 'updateMembership',
				description: 'Update a user\'s membership role',
				action: 'Update membership role',
			},
		],
		default: 'listUsers',
	},
	...listUsersOp.description,
	...getUserOp.description,
	...createUserOp.description,
	...inviteUserOp.description,
	...updateUserOp.description,
	...updateMembershipOp.description,
	...removeUserOp.description,
];

export {
	listUsersOp as listUsers,
	getUserOp as getUser,
	createUserOp as createUser,
	inviteUserOp as inviteUser,
	updateUserOp as updateUser,
	updateMembershipOp as updateMembership,
	removeUserOp as removeUser,
};
