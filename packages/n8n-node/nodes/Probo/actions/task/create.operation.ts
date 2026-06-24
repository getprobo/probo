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
