// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import type { INodeProperties, IExecuteFunctions, INodeExecutionData } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Organization ID',
		name: 'organizationId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['task'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The ID of the organization',
		required: true,
	},
	{
		displayName: 'Measure ID',
		name: 'measureId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['task'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The ID of the measure this task belongs to',
		required: true,
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['task'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The name of the task',
		required: true,
	},
	{
		displayName: 'Description',
		name: 'description',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['task'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The description of the task',
	},
	{
		displayName: 'Priority',
		name: 'priority',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['task'],
				operation: ['create'],
			},
		},
		options: [
			{
				name: 'Urgent',
				value: 'URGENT',
			},
			{
				name: 'High',
				value: 'HIGH',
			},
			{
				name: 'Medium',
				value: 'MEDIUM',
			},
			{
				name: 'Low',
				value: 'LOW',
			},
		],
		default: 'MEDIUM',
		description: 'The priority of the task',
	},
	{
		displayName: 'Time Estimate',
		name: 'timeEstimate',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['task'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The time estimate for the task',
	},
	{
		displayName: 'Assigned To ID',
		name: 'assignedToId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['task'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The ID of the user assigned to this task',
	},
	{
		displayName: 'Deadline',
		name: 'deadline',
		type: 'dateTime',
		displayOptions: {
			show: {
				resource: ['task'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The deadline for the task',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const measureId = this.getNodeParameter('measureId', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex) as string;
	const description = this.getNodeParameter('description', itemIndex, '') as string;
	const priority = this.getNodeParameter('priority', itemIndex, '') as string;
	const timeEstimate = this.getNodeParameter('timeEstimate', itemIndex, '') as string;
	const assignedToId = this.getNodeParameter('assignedToId', itemIndex, '') as string;
	const deadline = this.getNodeParameter('deadline', itemIndex, '') as string;

	const query = `
		mutation CreateTask($input: CreateTaskInput!) {
			createTask(input: $input) {
				taskEdge {
					node {
						id
						name
						description
						state
						priority
						timeEstimate
						deadline
						createdAt
						updatedAt
					}
				}
			}
		}
	`;

	const variables = {
		input: {
			organizationId,
			measureId,
			name,
			...(description && { description }),
			...(priority && { priority }),
			...(timeEstimate && { timeEstimate }),
			...(assignedToId && { assignedToId }),
			...(deadline && { deadline }),
		},
	};

	const responseData = await proboApiRequest.call(this, query, variables);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
