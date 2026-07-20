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
		displayName: 'Commitment Group ID',
		name: 'commitmentGroupId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['trustCenter'],
				operation: ['updateCommitmentGroup'],
			},
		},
		default: '',
		description: 'The ID of the commitment group to update',
		required: true,
	},
	{
		displayName: 'Title',
		name: 'title',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['trustCenter'],
				operation: ['updateCommitmentGroup'],
			},
		},
		default: '',
		description: 'The title of the commitment group',
	},
	{
		displayName: 'Description',
		name: 'description',
		type: 'string',
		typeOptions: {
			rows: 4,
		},
		displayOptions: {
			show: {
				resource: ['trustCenter'],
				operation: ['updateCommitmentGroup'],
			},
		},
		default: '',
		description: 'The description of the commitment group',
	},
	{
		displayName: 'Rank',
		name: 'rank',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['trustCenter'],
				operation: ['updateCommitmentGroup'],
			},
		},
		default: '',
		description: 'The rank of the commitment group for ordering',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const commitmentGroupId = this.getNodeParameter('commitmentGroupId', itemIndex) as string;
	const title = this.getNodeParameter('title', itemIndex, '') as string;
	const description = this.getNodeParameter('description', itemIndex, '') as string;
	const rank = this.getNodeParameter('rank', itemIndex, '') as string;

	const query = `
		mutation UpdateCompliancePortalCommitmentGroup($input: UpdateCompliancePortalCommitmentGroupInput!) {
			updateCompliancePortalCommitmentGroup(input: $input) {
				compliancePortalCommitmentGroup {
					id
					title
					description
					rank
					createdAt
					updatedAt
				}
			}
		}
	`;

	const input: Record<string, unknown> = { id: commitmentGroupId };
	if (title) input.title = title;
	if (description) input.description = description;
	if (rank) input.rank = Number(rank);

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
