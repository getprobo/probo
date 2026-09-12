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
import * as createOp from './create.operation';
import * as createCredentialOp from './createCredential.operation';
import * as deleteOp from './delete.operation';
import * as disableOp from './disable.operation';
import * as getOp from './get.operation';
import * as listOp from './list.operation';
import * as listCredentialsOp from './listCredentials.operation';
import * as revokeCredentialOp from './revokeCredential.operation';
import * as updateOp from './update.operation';

export const description: INodeProperties[] = [
	{
		displayName: 'Operation',
		name: 'operation',
		type: 'options',
		noDataExpression: true,
		displayOptions: {
			show: {
				resource: ['serviceAccount'],
			},
		},
		options: [
			{
				name: 'Create',
				value: 'create',
				description: 'Create a service account',
				action: 'Create a service account',
			},
			{
				name: 'Create Credential',
				value: 'createCredential',
				description: 'Create a credential and return its token',
				action: 'Create a service account credential',
			},
			{
				name: 'Delete',
				value: 'delete',
				description: 'Delete a service account',
				action: 'Delete a service account',
			},
			{
				name: 'Disable',
				value: 'disable',
				description: 'Disable a service account',
				action: 'Disable a service account',
			},
			{
				name: 'Get',
				value: 'get',
				description: 'Get a service account',
				action: 'Get a service account',
			},
			{
				name: 'List',
				value: 'list',
				description: 'List service accounts',
				action: 'List service accounts',
			},
			{
				name: 'List Credentials',
				value: 'listCredentials',
				description: 'List credentials for a service account',
				action: 'List service account credentials',
			},
			{
				name: 'Revoke Credential',
				value: 'revokeCredential',
				description: 'Revoke a service account credential',
				action: 'Revoke a service account credential',
			},
			{
				name: 'Update',
				value: 'update',
				description: 'Update a service account',
				action: 'Update a service account',
			},
		],
		default: 'list',
	},
	...createOp.description,
	...createCredentialOp.description,
	...deleteOp.description,
	...disableOp.description,
	...getOp.description,
	...listOp.description,
	...listCredentialsOp.description,
	...revokeCredentialOp.description,
	...updateOp.description,
];

export {
	createOp as create,
	createCredentialOp as createCredential,
	deleteOp as delete,
	disableOp as disable,
	getOp as get,
	listOp as list,
	listCredentialsOp as listCredentials,
	revokeCredentialOp as revokeCredential,
	updateOp as update,
};
