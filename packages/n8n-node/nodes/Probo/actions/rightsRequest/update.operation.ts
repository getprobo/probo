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
		displayName: 'Rights Request ID',
		name: 'id',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['rightsRequest'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the rights request to update',
		required: true,
	},
	{
		displayName: 'Request Type',
		name: 'requestType',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['rightsRequest'],
				operation: ['update'],
			},
		},
		options: [
			{
				name: '(Unchanged)',
				value: '',
			},
			{
				name: 'Access',
				value: 'ACCESS',
			},
			{
				name: 'Deletion',
				value: 'DELETION',
			},
			{
				name: 'Portability',
				value: 'PORTABILITY',
			},
		],
		default: '',
		description: 'The type of rights request',
	},
	{
		displayName: 'Request State',
		name: 'requestState',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['rightsRequest'],
				operation: ['update'],
			},
		},
		options: [
			{
				name: '(Unchanged)',
				value: '',
			},
			{
				name: 'To Do',
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
		description: 'The state of the rights request',
	},
	{
		displayName: 'Data Subject',
		name: 'dataSubject',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['rightsRequest'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The data subject of the rights request',
	},
	{
		displayName: 'Contact',
		name: 'contact',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['rightsRequest'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The contact for the rights request',
	},
	{
		displayName: 'Details',
		name: 'details',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['rightsRequest'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The details of the rights request',
	},
	{
		displayName: 'Deadline',
		name: 'deadline',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['rightsRequest'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The deadline for the rights request',
	},
	{
		displayName: 'Action Taken',
		name: 'actionTaken',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['rightsRequest'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The action taken for the rights request',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const id = this.getNodeParameter('id', itemIndex) as string;
	const requestType = this.getNodeParameter('requestType', itemIndex, '') as string;
	const requestState = this.getNodeParameter('requestState', itemIndex, '') as string;
	const dataSubject = this.getNodeParameter('dataSubject', itemIndex, '') as string;
	const contact = this.getNodeParameter('contact', itemIndex, '') as string;
	const details = this.getNodeParameter('details', itemIndex, '') as string;
	const deadline = this.getNodeParameter('deadline', itemIndex, '') as string;
	const actionTaken = this.getNodeParameter('actionTaken', itemIndex, '') as string;

	const query = `
		mutation UpdateRightsRequest($input: UpdateRightsRequestInput!) {
			updateRightsRequest(input: $input) {
				rightsRequest {
					id
					requestType
					requestState
					dataSubject
					contact
					details
					deadline
					actionTaken
					createdAt
					updatedAt
				}
			}
		}
	`;

	const input: Record<string, string> = { id };
	if (requestType) input.requestType = requestType;
	if (requestState) input.requestState = requestState;
	if (dataSubject) input.dataSubject = dataSubject;
	if (contact) input.contact = contact;
	if (details) input.details = details;
	if (deadline) input.deadline = deadline;
	if (actionTaken) input.actionTaken = actionTaken;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
