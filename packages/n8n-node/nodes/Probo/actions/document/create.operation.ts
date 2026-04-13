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
				resource: ['document'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The ID of the organization',
		required: true,
	},
	{
		displayName: 'Title',
		name: 'title',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['document'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The title of the document',
		required: true,
	},
	{
		displayName: 'Document Type',
		name: 'documentType',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['document'],
				operation: ['create'],
			},
		},
		options: [
			{ name: 'Governance', value: 'GOVERNANCE' },
			{ name: 'Other', value: 'OTHER' },
			{ name: 'Plan', value: 'PLAN' },
			{ name: 'Policy', value: 'POLICY' },
			{ name: 'Procedure', value: 'PROCEDURE' },
			{ name: 'Record', value: 'RECORD' },
			{ name: 'Register', value: 'REGISTER' },
			{ name: 'Report', value: 'REPORT' },
			{ name: 'Template', value: 'TEMPLATE' },
		],
		default: 'POLICY',
		description: 'The type of the document',
		required: true,
	},
	{
		displayName: 'Classification',
		name: 'classification',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['document'],
				operation: ['create'],
			},
		},
		options: [
			{ name: 'Confidential', value: 'CONFIDENTIAL' },
			{ name: 'Internal', value: 'INTERNAL' },
			{ name: 'Public', value: 'PUBLIC' },
			{ name: 'Secret', value: 'SECRET' },
		],
		default: 'INTERNAL',
		description: 'The classification of the document',
		required: true,
	},
	{
		displayName: 'Content',
		name: 'content',
		type: 'string',
		typeOptions: {
			rows: 6,
		},
		displayOptions: {
			show: {
				resource: ['document'],
				operation: ['create'],
			},
		},
		default: '',
		description: 'The content of the document in markdown format',
	},
	{
		displayName: 'Additional Fields',
		name: 'additionalFields',
		type: 'collection',
		placeholder: 'Add Field',
		default: {},
		displayOptions: {
			show: {
				resource: ['document'],
				operation: ['create'],
			},
		},
		options: [
			{
				displayName: 'Trust Center Visibility',
				name: 'trustCenterVisibility',
				type: 'options',
				options: [
					{ name: 'None', value: 'NONE' },
					{ name: 'Private', value: 'PRIVATE' },
					{ name: 'Public', value: 'PUBLIC' },
				],
				default: 'NONE',
				description: 'The trust center visibility of the document',
			},
			{
				displayName: 'Default Approver IDs',
				name: 'defaultApproverIds',
				type: 'string',
				default: '',
				description: 'Comma-separated list of default approver profile IDs',
			},
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const title = this.getNodeParameter('title', itemIndex) as string;
	const documentType = this.getNodeParameter('documentType', itemIndex) as string;
	const classification = this.getNodeParameter('classification', itemIndex) as string;
	const content = this.getNodeParameter('content', itemIndex, '') as string;
	const additionalFields = this.getNodeParameter('additionalFields', itemIndex, {}) as {
		trustCenterVisibility?: string;
		defaultApproverIds?: string;
	};

	const query = `
		mutation CreateDocument($input: CreateDocumentInput!) {
			createDocument(input: $input) {
				documentEdge {
					node {
						id
						status
						trustCenterVisibility
						currentPublishedMajor
						currentPublishedMinor
						archivedAt
						createdAt
						updatedAt
					}
				}
				documentVersionEdge {
					node {
						id
						title
						major
						minor
						status
						content
						changelog
						classification
						documentType
						publishedAt
						createdAt
						updatedAt
					}
				}
			}
		}
	`;

	const input: Record<string, unknown> = {
		organizationId,
		title,
		documentType,
		classification,
	};
	if (content) input.content = content;
	if (additionalFields.trustCenterVisibility) input.trustCenterVisibility = additionalFields.trustCenterVisibility;
	if (additionalFields.defaultApproverIds) {
		input.defaultApproverIds = additionalFields.defaultApproverIds.split(',').map(id => id.trim()).filter(Boolean);
	}

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
