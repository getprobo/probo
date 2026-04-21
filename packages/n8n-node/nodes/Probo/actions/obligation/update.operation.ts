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
		displayName: 'Obligation ID',
		name: 'id',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['obligation'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the obligation to update',
		required: true,
	},
	{
		displayName: 'Area',
		name: 'area',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['obligation'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The area of the obligation',
	},
	{
		displayName: 'Source',
		name: 'source',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['obligation'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The source of the obligation',
	},
	{
		displayName: 'Requirement',
		name: 'requirement',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['obligation'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The requirement of the obligation',
	},
	{
		displayName: 'Actions to Be Implemented',
		name: 'actionsToBeImplemented',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['obligation'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The actions to be implemented for the obligation',
	},
	{
		displayName: 'Regulator',
		name: 'regulator',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['obligation'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The regulator of the obligation',
	},
	{
		displayName: 'Owner ID',
		name: 'ownerId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['obligation'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the owner',
	},
	{
		displayName: 'Last Review Date',
		name: 'lastReviewDate',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['obligation'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The last review date of the obligation',
	},
	{
		displayName: 'Due Date',
		name: 'dueDate',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['obligation'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The due date of the obligation',
	},
	{
		displayName: 'Status',
		name: 'status',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['obligation'],
				operation: ['update'],
			},
		},
		options: [
			{
				name: '(Unchanged)',
				value: '',
			},
			{
				name: 'Non Compliant',
				value: 'NON_COMPLIANT',
			},
			{
				name: 'Partially Compliant',
				value: 'PARTIALLY_COMPLIANT',
			},
			{
				name: 'Compliant',
				value: 'COMPLIANT',
			},
		],
		default: '',
		description: 'The status of the obligation',
	},
	{
		displayName: 'Type',
		name: 'type',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['obligation'],
				operation: ['update'],
			},
		},
		options: [
			{
				name: '(Unchanged)',
				value: '',
			},
			{
				name: 'Legal',
				value: 'LEGAL',
			},
			{
				name: 'Contractual',
				value: 'CONTRACTUAL',
			},
		],
		default: '',
		description: 'The type of the obligation',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const id = this.getNodeParameter('id', itemIndex) as string;
	const area = this.getNodeParameter('area', itemIndex, '') as string;
	const source = this.getNodeParameter('source', itemIndex, '') as string;
	const requirement = this.getNodeParameter('requirement', itemIndex, '') as string;
	const actionsToBeImplemented = this.getNodeParameter('actionsToBeImplemented', itemIndex, '') as string;
	const regulator = this.getNodeParameter('regulator', itemIndex, '') as string;
	const ownerId = this.getNodeParameter('ownerId', itemIndex, '') as string;
	const lastReviewDate = this.getNodeParameter('lastReviewDate', itemIndex, '') as string;
	const dueDate = this.getNodeParameter('dueDate', itemIndex, '') as string;
	const status = this.getNodeParameter('status', itemIndex, '') as string;
	const type = this.getNodeParameter('type', itemIndex, '') as string;

	const query = `
		mutation UpdateObligation($input: UpdateObligationInput!) {
			updateObligation(input: $input) {
				obligation {
					id
					area
					source
					requirement
					actionsToBeImplemented
					regulator
					lastReviewDate
					dueDate
					status
					type
					createdAt
					updatedAt
				}
			}
		}
	`;

	const input: Record<string, string> = { id };
	if (area) input.area = area;
	if (source) input.source = source;
	if (requirement) input.requirement = requirement;
	if (actionsToBeImplemented) input.actionsToBeImplemented = actionsToBeImplemented;
	if (regulator) input.regulator = regulator;
	if (ownerId) input.ownerId = ownerId;
	if (lastReviewDate) input.lastReviewDate = lastReviewDate;
	if (dueDate) input.dueDate = dueDate;
	if (status) input.status = status;
	if (type) input.type = type;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
