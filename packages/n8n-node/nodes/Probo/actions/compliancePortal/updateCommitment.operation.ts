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

const commitmentIconOptions = [
	{ name: '(Unchanged)', value: '' },
	{ name: 'Lock Key', value: 'LOCK_KEY' },
	{ name: 'Eye Slash', value: 'EYE_SLASH' },
	{ name: 'Fingerprint', value: 'FINGERPRINT' },
	{ name: 'Shield Warning', value: 'SHIELD_WARNING' },
	{ name: 'Shield Check', value: 'SHIELD_CHECK' },
	{ name: 'Siren', value: 'SIREN' },
	{ name: 'Key', value: 'KEY' },
	{ name: 'Lock', value: 'LOCK' },
	{ name: 'Cloud', value: 'CLOUD' },
	{ name: 'Database', value: 'DATABASE' },
	{ name: 'Globe', value: 'GLOBE' },
	{ name: 'Eye', value: 'EYE' },
	{ name: 'Users', value: 'USERS' },
	{ name: 'Certificate', value: 'CERTIFICATE' },
	{ name: 'Gavel', value: 'GAVEL' },
	{ name: 'Heartbeat', value: 'HEARTBEAT' },
	{ name: 'Bell', value: 'BELL' },
	{ name: 'Bug', value: 'BUG' },
	{ name: 'Code', value: 'CODE' },
	{ name: 'Server', value: 'SERVER' },
];

export const description: INodeProperties[] = [
	{
		displayName: 'Commitment ID',
		name: 'commitmentId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['trustCenter'],
				operation: ['updateCommitment'],
			},
		},
		default: '',
		description: 'The ID of the commitment to update',
		required: true,
	},
	{
		displayName: 'Icon',
		name: 'icon',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['trustCenter'],
				operation: ['updateCommitment'],
			},
		},
		options: commitmentIconOptions,
		default: '',
		description: 'The icon of the commitment',
	},
	{
		displayName: 'Eyebrow',
		name: 'eyebrow',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['trustCenter'],
				operation: ['updateCommitment'],
			},
		},
		default: '',
		description: 'The eyebrow text of the commitment',
	},
	{
		displayName: 'Title',
		name: 'title',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['trustCenter'],
				operation: ['updateCommitment'],
			},
		},
		default: '',
		description: 'The title of the commitment',
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
				operation: ['updateCommitment'],
			},
		},
		default: '',
		description: 'The description of the commitment',
	},
	{
		displayName: 'Rank',
		name: 'rank',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['trustCenter'],
				operation: ['updateCommitment'],
			},
		},
		default: '',
		description: 'The rank of the commitment for ordering',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const commitmentId = this.getNodeParameter('commitmentId', itemIndex) as string;
	const icon = this.getNodeParameter('icon', itemIndex, '') as string;
	const eyebrow = this.getNodeParameter('eyebrow', itemIndex, '') as string;
	const title = this.getNodeParameter('title', itemIndex, '') as string;
	const description = this.getNodeParameter('description', itemIndex, '') as string;
	const rank = this.getNodeParameter('rank', itemIndex, '') as string;

	const query = `
		mutation UpdateCompliancePortalCommitment($input: UpdateCompliancePortalCommitmentInput!) {
			updateCompliancePortalCommitment(input: $input) {
				compliancePortalCommitment {
					id
					icon
					eyebrow
					title
					description
					rank
					createdAt
					updatedAt
				}
			}
		}
	`;

	const input: Record<string, unknown> = { id: commitmentId };
	if (icon) input.icon = icon;
	if (eyebrow) input.eyebrow = eyebrow;
	if (title) input.title = title;
	if (description) input.description = description;
	if (rank) input.rank = Number(rank);

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
