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
		displayName: 'Task ID',
		name: 'taskId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['task'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the task to update',
		required: true,
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['task'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The name of the task',
	},
	{
		displayName: 'Description',
		name: 'description',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['task'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The description of the task',
	},
	{
		displayName: 'State',
		name: 'state',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['task'],
				operation: ['update'],
			},
		},
		options: [
			{
				name: '(Unchanged)',
				value: '',
			},
			{
				name: 'Todo',
				value: 'TODO',
			},
			{
				name: 'In Progress',
				value: 'IN_PROGRESS',
			},
			{
				name: 'Done',
				value: 'DONE',
			},
		],
		default: '',
		description: 'The state of the task',
	},
	{
		displayName: 'Priority',
		name: 'priority',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['task'],
				operation: ['update'],
			},
		},
		options: [
			{
				name: '(Unchanged)',
				value: '',
			},
			{
				name: 'High',
				value: 'HIGH',
			},
			{
				name: 'Low',
				value: 'LOW',
			},
			{
				name: 'Medium',
				value: 'MEDIUM',
			},
			{
				name: 'Urgent',
				value: 'URGENT',
			},
		],
		default: '',
		description: 'The priority of the task',
	},
	{
		displayName: 'Rank',
		name: 'rank',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['task'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The rank of the task for ordering',
	},
	{
		displayName: 'Time Estimate',
		name: 'timeEstimate',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['task'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The time estimate for the task',
	},
	{
		displayName: 'Deadline',
		name: 'deadline',
		type: 'dateTime',
		displayOptions: {
			show: {
				resource: ['task'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The deadline for the task',
	},
	{
		displayName: 'Assigned To ID',
		name: 'assignedToId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['task'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the user assigned to this task',
	},
	{
		displayName: 'Measure ID',
		name: 'measureId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['task'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the measure this task belongs to',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const taskId = this.getNodeParameter('taskId', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex, '') as string;
	const description = this.getNodeParameter('description', itemIndex, '') as string;
	const state = this.getNodeParameter('state', itemIndex, '') as string;
	const priority = this.getNodeParameter('priority', itemIndex, '') as string;
	const rank = this.getNodeParameter('rank', itemIndex, '') as string;
	const timeEstimate = this.getNodeParameter('timeEstimate', itemIndex, '') as string;
	const deadline = this.getNodeParameter('deadline', itemIndex, '') as string;
	const assignedToId = this.getNodeParameter('assignedToId', itemIndex, '') as string;
	const measureId = this.getNodeParameter('measureId', itemIndex, '') as string;

	const query = `
		mutation UpdateTask($input: UpdateTaskInput!) {
			updateTask(input: $input) {
				task {
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
	`;

	const input: Record<string, string> = { taskId };
	if (name) input.name = name;
	if (description) input.description = description;
	if (state) input.state = state;
	if (priority) input.priority = priority;
	if (rank) input.rank = rank;
	if (timeEstimate) input.timeEstimate = timeEstimate;
	if (deadline) input.deadline = deadline;
	if (assignedToId) input.assignedToId = assignedToId;
	if (measureId) input.measureId = measureId;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
