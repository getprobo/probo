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
		displayName: 'Document ID',
		name: 'documentId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['document'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the document to update',
		required: true,
	},
	{
		displayName: 'Update Fields',
		name: 'updateFields',
		type: 'collection',
		placeholder: 'Add Field',
		default: {},
		displayOptions: {
			show: {
				resource: ['document'],
				operation: ['update'],
			},
		},
		options: [
			{
				displayName: 'Classification',
				name: 'classification',
				type: 'options',
				options: [
					{ name: 'Confidential', value: 'CONFIDENTIAL' },
					{ name: 'Internal', value: 'INTERNAL' },
					{ name: 'Public', value: 'PUBLIC' },
					{ name: 'Secret', value: 'SECRET' },
				],
				default: 'INTERNAL',
				description: 'The classification of the document. Updating it edits the current draft version, creating one from the latest published version if none exists.',
			},
			{
				displayName: 'Content',
				name: 'content',
				type: 'string',
				typeOptions: {
					rows: 6,
				},
				default: '',
				description: 'The body of the document as a ProseMirror document JSON string. Updating it edits the current draft version, creating one from the latest published version if none exists.',
			},
			{
				displayName: 'Default Approver IDs',
				name: 'defaultApproverIds',
				type: 'string',
				default: '',
				description: 'Comma-separated list of default approver profile IDs',
			},
			{
				displayName: 'Document Type',
				name: 'documentType',
				type: 'options',
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
				description: 'The type of the document. Updating it edits the current draft version, creating one from the latest published version if none exists.',
			},
			{
				displayName: 'Title',
				name: 'title',
				type: 'string',
				default: '',
				description: 'The title of the document. Updating it edits the current draft version, creating one from the latest published version if none exists.',
			},
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
		],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const documentId = this.getNodeParameter('documentId', itemIndex) as string;
	const updateFields = this.getNodeParameter('updateFields', itemIndex, {}) as {
		title?: string;
		content?: string;
		classification?: string;
		documentType?: string;
		trustCenterVisibility?: string;
		defaultApproverIds?: string;
	};

	const query = `
		mutation UpdateDocument($input: UpdateDocumentInput!) {
			updateDocument(input: $input) {
				document {
					id
					status
					trustCenterVisibility
					currentPublishedMajor
					currentPublishedMinor
					archivedAt
					createdAt
					updatedAt
				}
				documentVersion {
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
	`;

	const input: Record<string, unknown> = { id: documentId };
	if (updateFields.title !== undefined && updateFields.title !== '') input.title = updateFields.title;
	if (updateFields.content !== undefined && updateFields.content !== '') input.content = updateFields.content;
	if (updateFields.classification !== undefined) input.classification = updateFields.classification;
	if (updateFields.documentType !== undefined) input.documentType = updateFields.documentType;
	if (updateFields.trustCenterVisibility !== undefined) input.trustCenterVisibility = updateFields.trustCenterVisibility;
	if (updateFields.defaultApproverIds !== undefined && updateFields.defaultApproverIds !== '') {
		input.defaultApproverIds = updateFields.defaultApproverIds.split(',').map(id => id.trim()).filter(Boolean);
	}

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
