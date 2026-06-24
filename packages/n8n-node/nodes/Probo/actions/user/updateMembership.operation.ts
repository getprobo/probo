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

import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboConnectApiRequest } from '../../GenericFunctions';

const roleOptions = [
	{ name: 'Owner', value: 'OWNER' },
	{ name: 'Admin', value: 'ADMIN' },
	{ name: 'Employee', value: 'EMPLOYEE' },
	{ name: 'Viewer', value: 'VIEWER' },
	{ name: 'Auditor', value: 'AUDITOR' },
];

export const description: INodeProperties[] = [
	{
		displayName: 'Organization ID',
		name: 'organizationId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['user'],
				operation: ['updateMembership'],
			},
		},
		default: '',
		description: 'The ID of the organization',
		required: true,
	},
	{
		displayName: 'Membership ID',
		name: 'membershipId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['user'],
				operation: ['updateMembership'],
			},
		},
		default: '',
		description: 'The ID of the membership to update',
		required: true,
	},
	{
		displayName: 'Role',
		name: 'role',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['user'],
				operation: ['updateMembership'],
			},
		},
		options: roleOptions,
		default: 'EMPLOYEE',
		description: 'New role for the membership',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const membershipId = this.getNodeParameter('membershipId', itemIndex) as string;
	const role = this.getNodeParameter('role', itemIndex) as string;

	const query = `
		mutation UpdateMembership($input: UpdateMembershipInput!) {
			updateMembership(input: $input) {
				membership {
					id
					role
					createdAt
				}
			}
		}
	`;

	const input = { organizationId, membershipId, role };
	const responseData = await proboConnectApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
