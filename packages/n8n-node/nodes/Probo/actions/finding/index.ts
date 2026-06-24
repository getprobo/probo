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
import * as linkAuditOp from './linkAudit.operation';
import * as publishOp from './publish.operation';
import * as unlinkAuditOp from './unlinkAudit.operation';

export const description: INodeProperties[] = [
	{
		displayName: 'Operation',
		name: 'operation',
		type: 'options',
		noDataExpression: true,
		displayOptions: {
			show: {
				resource: ['finding'],
			},
		},
		options: [
			{
				name: 'Create',
				value: 'create',
				description: 'Create a new finding',
				action: 'Create a finding',
			},
			{
				name: 'Delete',
				value: 'delete',
				description: 'Delete a finding',
				action: 'Delete a finding',
			},
			{
				name: 'Get',
				value: 'get',
				description: 'Get a finding',
				action: 'Get a finding',
			},
			{
				name: 'Get Many',
				value: 'getAll',
				description: 'Get many findings',
				action: 'Get many findings',
			},
			{
				name: 'Link Audit',
				value: 'linkAudit',
				description: 'Link an audit to a finding',
				action: 'Link an audit to a finding',
			},
			{
				name: 'Publish',
				value: 'publish',
				description: 'Publish the finding list as a document',
				action: 'Publish the finding list',
			},
			{
				name: 'Unlink Audit',
				value: 'unlinkAudit',
				description: 'Unlink an audit from a finding',
				action: 'Unlink an audit from a finding',
			},
			{
				name: 'Update',
				value: 'update',
				description: 'Update an existing finding',
				action: 'Update a finding',
			},
		],
		default: 'create',
	},
	...createOp.description,
	...updateOp.description,
	...deleteOp.description,
	...getOp.description,
	...getAllOp.description,
	...linkAuditOp.description,
	...publishOp.description,
	...unlinkAuditOp.description,
];

export {
	createOp as create,
	updateOp as update,
	deleteOp as delete,
	getOp as get,
	getAllOp as getAll,
	linkAuditOp as linkAudit,
	publishOp as publish,
	unlinkAuditOp as unlinkAudit,
};
