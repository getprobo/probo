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
		displayName: 'Finding ID',
		name: 'findingId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['finding'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the finding to update',
		required: true,
	},
	{
		displayName: 'Additional Fields',
		name: 'additionalFields',
		type: 'collection',
		placeholder: 'Add Field',
		default: {},
		displayOptions: {
			show: {
				resource: ['finding'],
				operation: ['update'],
			},
		},
		options: [
			{
				displayName: 'Corrective Action',
				name: 'correctiveAction',
				type: 'string',
				default: '',
				description: 'The corrective action for the finding',
			},
			{
				displayName: 'Description',
				name: 'description',
				type: 'string',
				default: '',
				description: 'The description of the finding',
			},
			{
				displayName: 'Due Date',
				name: 'dueDate',
				type: 'string',
				default: '',
				description: 'The due date for the finding (ISO 8601 format)',
			},
			{
				displayName: 'Effectiveness Check',
				name: 'effectivenessCheck',
				type: 'string',
				default: '',
				description: 'The effectiveness check for the finding',
			},
			{
				displayName: 'Identified On',
				name: 'identifiedOn',
				type: 'string',
				default: '',
				description: 'The date the finding was identified (ISO 8601 format)',
			},
			{
				displayName: 'Owner ID',
				name: 'ownerId',
				type: 'string',
				default: '',
				description: 'The ID of the person who owns this finding',
			},
			{
				displayName: 'Priority',
				name: 'priority',
				type: 'options',
				options: [
					{
						name: '(Unchanged)',
						value: '',
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
						name: 'High',
						value: 'HIGH',
					},
				],
				default: '',
				description: 'The priority of the finding',
			},
			{
				displayName: 'Risk ID',
				name: 'riskId',
				type: 'string',
				default: '',
				description: 'The ID of the associated risk',
			},
			{
				displayName: 'Root Cause',
				name: 'rootCause',
				type: 'string',
				default: '',
				description: 'The root cause of the finding',
			},
			{
				displayName: 'Source',
				name: 'source',
				type: 'string',
				default: '',
				description: 'The source of the finding',
			},
			{
				displayName: 'Status',
				name: 'status',
				type: 'options',
				options: [
					{
						name: '(Unchanged)',
						value: '',
					},
					{
						name: 'Closed',
						value: 'CLOSED',
					},
					{
						name: 'False Positive',
						value: 'FALSE_POSITIVE',
					},
					{
						name: 'In Progress',
						value: 'IN_PROGRESS',
					},
					{
						name: 'Mitigated',
						value: 'MITIGATED',
					},
					{
						name: 'Open',
						value: 'OPEN',
					},
					{
						name: 'Risk Accepted',
						value: 'RISK_ACCEPTED',
					},
				],
				default: '',
				description: 'The status of the finding',
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const findingId = this.getNodeParameter('findingId', itemIndex) as string;
	const additionalFields = this.getNodeParameter('additionalFields', itemIndex, {}) as {
		description?: string;
		source?: string;
		identifiedOn?: string;
		rootCause?: string;
		correctiveAction?: string;
		ownerId?: string;
		dueDate?: string;
		status?: string;
		priority?: string;
		riskId?: string;
		effectivenessCheck?: string;
	};

	const query = `
		mutation UpdateFinding($input: UpdateFindingInput!) {
			updateFinding(input: $input) {
				finding {
					id
					kind
					description
					source
					identifiedOn
					rootCause
					correctiveAction
					dueDate
					status
					priority
					effectivenessCheck
					createdAt
					updatedAt
				}
			}
		}
	`;

	const input: Record<string, unknown> = { id: findingId };
	if (additionalFields.description !== undefined) input.description = additionalFields.description === '' ? null : additionalFields.description;
	if (additionalFields.source !== undefined) input.source = additionalFields.source === '' ? null : additionalFields.source;
	if (additionalFields.identifiedOn !== undefined) input.identifiedOn = additionalFields.identifiedOn === '' ? null : additionalFields.identifiedOn;
	if (additionalFields.rootCause !== undefined) input.rootCause = additionalFields.rootCause === '' ? null : additionalFields.rootCause;
	if (additionalFields.correctiveAction !== undefined) input.correctiveAction = additionalFields.correctiveAction === '' ? null : additionalFields.correctiveAction;
	if (additionalFields.ownerId !== undefined) input.ownerId = additionalFields.ownerId === '' ? null : additionalFields.ownerId;
	if (additionalFields.dueDate !== undefined) input.dueDate = additionalFields.dueDate === '' ? null : additionalFields.dueDate;
	if (additionalFields.status !== undefined) input.status = additionalFields.status;
	if (additionalFields.priority !== undefined) input.priority = additionalFields.priority;
	if (additionalFields.riskId !== undefined) input.riskId = additionalFields.riskId === '' ? null : additionalFields.riskId;
	if (additionalFields.effectivenessCheck !== undefined) input.effectivenessCheck = additionalFields.effectivenessCheck === '' ? null : additionalFields.effectivenessCheck;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
