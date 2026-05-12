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
				resource: ['finding'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The ID of the organization',
		required: true,
	},
	{
		displayName: 'Kind',
		name: 'kind',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['finding'],
				operation: ['create'],
			},
		},
		options: [
			{
				name: 'Minor Nonconformity',
				value: 'MINOR_NONCONFORMITY',
			},
			{
				name: 'Major Nonconformity',
				value: 'MAJOR_NONCONFORMITY',
			},
			{
				name: 'Observation',
				value: 'OBSERVATION',
			},
			{
				name: 'Exception',
				value: 'EXCEPTION',
			},
		],
		default: 'MINOR_NONCONFORMITY',
		description: 'The kind of finding',
		required: true,
	},
	{
		displayName: 'Description',
		name: 'description',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['finding'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The description of the finding',
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
				operation: ['create'],
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
				default: 'MEDIUM',
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
				default: 'OPEN',
				description: 'The status of the finding',
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const kind = this.getNodeParameter('kind', itemIndex) as string;
	const description = this.getNodeParameter('description', itemIndex) as string;
	const additionalFields = this.getNodeParameter('additionalFields', itemIndex, {}) as {
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
		mutation CreateFinding($input: CreateFindingInput!) {
			createFinding(input: $input) {
				findingEdge {
					node {
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
		}
	`;

	const input: Record<string, unknown> = {
		organizationId,
		kind,
		description,
	};
	if (additionalFields.source) input.source = additionalFields.source;
	if (additionalFields.identifiedOn) input.identifiedOn = additionalFields.identifiedOn;
	if (additionalFields.rootCause) input.rootCause = additionalFields.rootCause;
	if (additionalFields.correctiveAction) input.correctiveAction = additionalFields.correctiveAction;
	if (additionalFields.ownerId) input.ownerId = additionalFields.ownerId;
	if (additionalFields.dueDate) input.dueDate = additionalFields.dueDate;
	if (additionalFields.status) input.status = additionalFields.status;
	if (additionalFields.priority) input.priority = additionalFields.priority;
	if (additionalFields.riskId) input.riskId = additionalFields.riskId;
	if (additionalFields.effectivenessCheck) input.effectivenessCheck = additionalFields.effectivenessCheck;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
