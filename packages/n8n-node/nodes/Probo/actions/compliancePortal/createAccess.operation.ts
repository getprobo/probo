// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
		displayName: 'Compliance Portal ID',
		name: 'compliancePortalId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['compliancePortal'],
				operation: ['createAccess'],
			},
		},
		default: '',
		description: 'The ID of the compliance portal',
		required: true,
	},
	{
		displayName: 'Profile ID',
		name: 'profileId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['compliancePortal'],
				operation: ['createAccess'],
			},
		},
		default: '',
		description: 'Existing organization profile ID. Provide exactly one of Profile ID or Email.',
	},
	{
		displayName: 'Email',
		name: 'email',
		type: 'string',
		placeholder: 'name@email.com',
		displayOptions: {
			show: {
				resource: ['compliancePortal'],
				operation: ['createAccess'],
			},
		},
		default: '',
		description: 'Visitor email address. Provide exactly one of Profile ID or Email.',
	},
	{
		displayName: 'Document IDs',
		name: 'documentIds',
		type: 'string',
		typeOptions: {
			multipleValues: true,
		},
		displayOptions: {
			show: {
				resource: ['compliancePortal'],
				operation: ['createAccess'],
			},
		},
		default: [],
		description: 'Document IDs to grant immediately',
	},
	{
		displayName: 'Report IDs',
		name: 'reportIds',
		type: 'string',
		typeOptions: {
			multipleValues: true,
		},
		displayOptions: {
			show: {
				resource: ['compliancePortal'],
				operation: ['createAccess'],
			},
		},
		default: [],
		description: 'Report file IDs to grant immediately',
	},
	{
		displayName: 'Compliance Portal File IDs',
		name: 'compliancePortalFileIds',
		type: 'string',
		typeOptions: {
			multipleValues: true,
		},
		displayOptions: {
			show: {
				resource: ['compliancePortal'],
				operation: ['createAccess'],
			},
		},
		default: [],
		description: 'Compliance portal file IDs to grant immediately',
	},
];

function stringList(value: unknown): string[] {
	if (!Array.isArray(value)) {
		return [];
	}

	return value.filter((item): item is string => typeof item === 'string' && item !== '');
}

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const compliancePortalId = this.getNodeParameter('compliancePortalId', itemIndex) as string;
	const profileId = this.getNodeParameter('profileId', itemIndex, '') as string;
	const email = this.getNodeParameter('email', itemIndex, '') as string;
	const documentIds = stringList(this.getNodeParameter('documentIds', itemIndex, []));
	const reportIds = stringList(this.getNodeParameter('reportIds', itemIndex, []));
	const compliancePortalFileIds = stringList(
		this.getNodeParameter('compliancePortalFileIds', itemIndex, []),
	);

	const query = `
		mutation CreateCompliancePortalAccess($input: CreateCompliancePortalAccessInput!) {
			createCompliancePortalAccess(input: $input) {
				compliancePortalAccessEdge {
					node {
						id
						state
						authenticatedAt
						createdAt
						updatedAt
					}
				}
			}
		}
	`;

	const input: Record<string, unknown> = { compliancePortalId };
	if (profileId) input.profileId = profileId;
	if (email) input.email = email;
	if (documentIds.length > 0) input.documents = documentIds;
	if (reportIds.length > 0) input.reports = reportIds;
	if (compliancePortalFileIds.length > 0) input.compliancePortalFiles = compliancePortalFileIds;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
