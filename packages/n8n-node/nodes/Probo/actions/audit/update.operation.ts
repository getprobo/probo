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
		displayName: 'Audit ID',
		name: 'id',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['audit'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the audit to update',
		required: true,
	},
	{
		displayName: 'Update Fields',
		name: 'updateFields',
		type: 'collection',
		placeholder: 'Add Field',
		default: {},
		displayOptions: {
			show: {
				resource: ['audit'],
				operation: ['update'],
			},
		},
		options: [
			{
				displayName: 'Name',
				name: 'name',
				type: 'string',
				default: '',
				description: 'The name of the audit',
			},
			{
				displayName: 'State',
				name: 'state',
				type: 'options',
				options: [
					{
						name: 'Completed',
						value: 'COMPLETED',
					},
					{
						name: 'In Progress',
						value: 'IN_PROGRESS',
					},
					{
						name: 'Not Started',
						value: 'NOT_STARTED',
					},
					{
						name: 'Outdated',
						value: 'OUTDATED',
					},
					{
						name: 'Rejected',
						value: 'REJECTED',
					},
				],
				default: 'NOT_STARTED',
				description: 'The state of the audit',
			},
			{
				displayName: 'Valid From',
				name: 'validFrom',
				type: 'dateTime',
				default: '',
				description: 'The start date of the audit validity period',
			},
			{
				displayName: 'Valid Until',
				name: 'validUntil',
				type: 'dateTime',
				default: '',
				description: 'The end date of the audit validity period',
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const id = this.getNodeParameter('id', itemIndex) as string;
	const updateFields = this.getNodeParameter('updateFields', itemIndex, {}) as {
		name?: string;
		state?: string;
		validFrom?: string;
		validUntil?: string;
	};

	const query = `
		mutation UpdateAudit($input: UpdateAuditInput!) {
			updateAudit(input: $input) {
				audit {
					id
					name
					state
					validFrom
					validUntil
					reportUrl
					trustCenterVisibility
					createdAt
					updatedAt
				}
			}
		}
	`;

	const input: Record<string, unknown> = { id };
	if (updateFields.name) input.name = updateFields.name;
	if (updateFields.state) input.state = updateFields.state;
	if (updateFields.validFrom) input.validFrom = updateFields.validFrom;
	if (updateFields.validUntil) input.validUntil = updateFields.validUntil;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
