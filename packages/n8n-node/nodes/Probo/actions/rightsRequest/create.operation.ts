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
				resource: ['rightsRequest'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The ID of the organization',
		required: true,
	},
	{
		displayName: 'Request Type',
		name: 'requestType',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['rightsRequest'],
				operation: ['create'],
			},
		},
		options: [
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
		default: 'ACCESS',
		description: 'The type of rights request',
		required: true,
	},
	{
		displayName: 'Request State',
		name: 'requestState',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['rightsRequest'],
				operation: ['create'],
			},
		},
		options: [
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
		default: 'TODO',
		description: 'The state of the rights request',
		required: true,
	},
	{
		displayName: 'Data Subject',
		name: 'dataSubject',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['rightsRequest'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The data subject of the rights request',
		required: true,
	},
	{
		displayName: 'Contact',
		name: 'contact',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['rightsRequest'],
				operation: ['create'],
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
				operation: ['create'],
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
				operation: ['create'],
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
				operation: ['create'],
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
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const requestType = this.getNodeParameter('requestType', itemIndex) as string;
	const requestState = this.getNodeParameter('requestState', itemIndex) as string;
	const dataSubject = this.getNodeParameter('dataSubject', itemIndex) as string;
	const contact = this.getNodeParameter('contact', itemIndex, '') as string;
	const details = this.getNodeParameter('details', itemIndex, '') as string;
	const deadline = this.getNodeParameter('deadline', itemIndex, '') as string;
	const actionTaken = this.getNodeParameter('actionTaken', itemIndex, '') as string;

	const query = `
		mutation CreateRightsRequest($input: CreateRightsRequestInput!) {
			createRightsRequest(input: $input) {
				rightsRequestEdge {
					node {
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
		}
	`;

	const variables = {
		input: {
			organizationId,
			requestType,
			requestState,
			dataSubject,
			...(contact && { contact }),
			...(details && { details }),
			...(deadline && { deadline }),
			...(actionTaken && { actionTaken }),
		},
	};

	const responseData = await proboApiRequest.call(this, query, variables);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
