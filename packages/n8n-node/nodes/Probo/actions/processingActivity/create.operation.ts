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
				resource: ['processingActivity'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The ID of the organization',
		required: true,
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['processingActivity'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The name of the processing activity',
		required: true,
	},
	{
		displayName: 'Purpose',
		name: 'purpose',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['processingActivity'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The purpose of the processing activity',
		required: true,
	},
	{
		displayName: 'Role',
		name: 'role',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['processingActivity'],
				operation: ['create'],
			},
		},
		options: [
			{
				name: 'Controller',
				value: 'CONTROLLER',
			},
			{
				name: 'Processor',
				value: 'PROCESSOR',
			},
		],
		default: 'CONTROLLER',
		description: 'The role for the processing activity',
		required: true,
	},
	{
		displayName: 'Lawful Basis',
		name: 'lawfulBasis',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['processingActivity'],
				operation: ['create'],
			},
		},
		options: [
			{
				name: 'Consent',
				value: 'CONSENT',
			},
			{
				name: 'Contractual Necessity',
				value: 'CONTRACTUAL_NECESSITY',
			},
			{
				name: 'Legal Obligation',
				value: 'LEGAL_OBLIGATION',
			},
			{
				name: 'Legitimate Interest',
				value: 'LEGITIMATE_INTEREST',
			},
			{
				name: 'Public Task',
				value: 'PUBLIC_TASK',
			},
			{
				name: 'Vital Interests',
				value: 'VITAL_INTERESTS',
			},
		],
		default: 'LEGITIMATE_INTEREST',
		description: 'The lawful basis for the processing activity',
		required: true,
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex) as string;
	const purpose = this.getNodeParameter('purpose', itemIndex) as string;
	const role = this.getNodeParameter('role', itemIndex) as string;
	const lawfulBasis = this.getNodeParameter('lawfulBasis', itemIndex) as string;

	const query = `
		mutation CreateProcessingActivity($input: CreateProcessingActivityInput!) {
			createProcessingActivity(input: $input) {
				processingActivityEdge {
					node {
						id
						name
						purpose
						role
						lawfulBasis
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
			name,
			purpose,
			role,
			lawfulBasis,
		},
	};

	const responseData = await proboApiRequest.call(this, query, variables);

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
