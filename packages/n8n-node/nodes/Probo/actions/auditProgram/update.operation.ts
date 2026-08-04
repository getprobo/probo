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
		displayName: 'Audit Program ID',
		name: 'id',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['auditProgram'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the audit program to update',
		required: true,
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['auditProgram'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The name of the audit program',
	},
	{
		displayName: 'Valid From',
		name: 'validFrom',
		type: 'dateTime',
		displayOptions: {
			show: {
				resource: ['auditProgram'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The start date of the audit program validity period',
	},
	{
		displayName: 'Valid Until',
		name: 'validUntil',
		type: 'dateTime',
		displayOptions: {
			show: {
				resource: ['auditProgram'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The end date of the audit program validity period',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const id = this.getNodeParameter('id', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex, '') as string;
	const validFrom = this.getNodeParameter('validFrom', itemIndex, '') as string;
	const validUntil = this.getNodeParameter('validUntil', itemIndex, '') as string;

	const query = `
		mutation UpdateAuditProgram($input: UpdateAuditProgramInput!) {
			updateAuditProgram(input: $input) {
				auditProgram {
					id
					name
					validFrom
					validUntil
					createdAt
					updatedAt
				}
			}
		}
	`;

	const input: Record<string, unknown> = { id };
	if (name) input.name = name;
	if (validFrom) input.validFrom = validFrom;
	if (validUntil) input.validUntil = validUntil;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
