// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
import * as archiveUserOp from './archiveUser.operation';
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
				name: 'Archive',
				value: 'archiveUser',
				description: 'Archive a user in the organization',
				action: 'Archive a user',
			},
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
	...archiveUserOp.description,
	...listUsersOp.description,
	...getUserOp.description,
	...createUserOp.description,
	...inviteUserOp.description,
	...updateUserOp.description,
	...updateMembershipOp.description,
	...removeUserOp.description,
];

export {
	archiveUserOp as archiveUser,
	listUsersOp as listUsers,
	getUserOp as getUser,
	createUserOp as createUser,
	inviteUserOp as inviteUser,
	updateUserOp as updateUser,
	updateMembershipOp as updateMembership,
	removeUserOp as removeUser,
};
