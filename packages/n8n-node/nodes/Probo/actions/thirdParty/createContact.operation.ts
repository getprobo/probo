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
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'ThirdParty ID',
		name: 'thirdPartyId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['thirdParty'],
				operation: ['createContact'],
			},
		},
		default: '',
		description: 'The ID of the thirdParty',
		required: true,
	},
	{
		displayName: 'Full Name',
		name: 'fullName',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['thirdParty'],
				operation: ['createContact'],
			},
		},
		default: '',
		description: 'The full name of the contact',
	},
	{
		displayName: 'Email',
		name: 'email',
		type: 'string',
		placeholder: 'name@email.com',
		displayOptions: {
			show: {
				resource: ['thirdParty'],
				operation: ['createContact'],
			},
		},
		default: '',
		description: 'The email address of the contact',
	},
	{
		displayName: 'Phone',
		name: 'phone',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['thirdParty'],
				operation: ['createContact'],
			},
		},
		default: '',
		description: 'The phone number of the contact',
	},
	{
		displayName: 'Role',
		name: 'role',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['thirdParty'],
				operation: ['createContact'],
			},
		},
		default: '',
		description: 'The role of the contact',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const thirdPartyId = this.getNodeParameter('thirdPartyId', itemIndex) as string;
	const fullName = this.getNodeParameter('fullName', itemIndex, '') as string;
	const email = this.getNodeParameter('email', itemIndex, '') as string;
	const phone = this.getNodeParameter('phone', itemIndex, '') as string;
	const role = this.getNodeParameter('role', itemIndex, '') as string;

	const query = `
		mutation CreateThirdPartyContact($input: CreateThirdPartyContactInput!) {
			createThirdPartyContact(input: $input) {
				thirdPartyContactEdge {
					node {
						id
						fullName
						email
						phone
						role
						createdAt
						updatedAt
					}
				}
			}
		}
	`;

	const input: Record<string, unknown> = { thirdPartyId };
	if (fullName) input.fullName = fullName;
	if (email) input.email = email;
	if (phone) input.phone = phone;
	if (role) input.role = role;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
