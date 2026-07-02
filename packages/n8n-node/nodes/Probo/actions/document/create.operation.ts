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
