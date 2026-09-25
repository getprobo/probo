// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
import * as createOp from './create.operation';
import * as updateOp from './update.operation';
import * as deleteOp from './delete.operation';
import * as getOp from './get.operation';
import * as getAllOp from './getAll.operation';
import * as linkToLinearOp from './linkToLinear.operation';
import * as listLinearIssuesOp from './listLinearIssues.operation';
import * as listLinearTeamsOp from './listLinearTeams.operation';
import * as publishToLinearOp from './publishToLinear.operation';
import * as unlinkExternalOp from './unlinkExternal.operation';

export const description: INodeProperties[] = [
	{
		displayName: 'Operation',
		name: 'operation',
		type: 'options',
		noDataExpression: true,
		displayOptions: {
			show: {
				resource: ['task'],
			},
		},
		options: [
			{
				name: 'Create',
				value: 'create',
				description: 'Create a new task',
				action: 'Create a task',
			},
			{
				name: 'Delete',
				value: 'delete',
				description: 'Delete a task',
				action: 'Delete a task',
			},
			{
				name: 'Get',
				value: 'get',
				description: 'Get a task',
				action: 'Get a task',
			},
			{
				name: 'Get Many',
				value: 'getAll',
				description: 'Get many tasks',
				action: 'Get many tasks',
			},
			{
				name: 'Link To Linear',
				value: 'linkToLinear',
				description: 'Link a task to an existing Linear issue. This updates the Probo task with the Linear issue.',
				action: 'Link a task to linear',
			},
			{
				name: 'List Linear Issues',
				value: 'listLinearIssues',
				description: 'Search Linear issues in a team',
				action: 'List linear issues',
			},
			{
				name: 'List Linear Teams',
				value: 'listLinearTeams',
				description: 'Search Linear teams for an organization',
				action: 'List linear teams',
			},
			{
				name: 'Publish To Linear',
				value: 'publishToLinear',
				description: 'Publish a task as a new Linear issue',
				action: 'Publish a task to linear',
			},
			{
				name: 'Unlink External',
				value: 'unlinkExternal',
				description: 'Unlink a task from its external issue',
				action: 'Unlink a task external issue',
			},
			{
				name: 'Update',
				value: 'update',
				description: 'Update an existing task',
				action: 'Update a task',
			},
		],
		default: 'create',
	},
	...createOp.description,
	...updateOp.description,
	...deleteOp.description,
	...getOp.description,
	...getAllOp.description,
	...linkToLinearOp.description,
	...listLinearIssuesOp.description,
	...listLinearTeamsOp.description,
	...publishToLinearOp.description,
	...unlinkExternalOp.description,
];

export {
	createOp as create,
	updateOp as update,
	deleteOp as delete,
	getOp as get,
	getAllOp as getAll,
	linkToLinearOp as linkToLinear,
	listLinearIssuesOp as listLinearIssues,
	listLinearTeamsOp as listLinearTeams,
	publishToLinearOp as publishToLinear,
	unlinkExternalOp as unlinkExternal,
};
