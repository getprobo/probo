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
import { accessReviewSourceIdsUpdateField } from './sources.fields';

export const description: INodeProperties[] = [
	{
		displayName: 'Access Review Campaign ID',
		name: 'accessReviewCampaignId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['accessReview'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The ID of the access review campaign to update',
		required: true,
	},
	{
		displayName: 'Name',
		name: 'name',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['accessReview'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The name of the access review campaign',
	},
	{
		displayName: 'Description',
		name: 'description',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['accessReview'],
				operation: ['update'],
			},
		},
		default: '',
		description: 'The description of the access review campaign',
	},
	{
		displayName: 'Additional Fields',
		name: 'additionalFields',
		type: 'collection',
		placeholder: 'Add Field',
		default: {},
		displayOptions: {
			show: {
				resource: ['accessReview'],
				operation: ['update'],
			},
		},
		options: [accessReviewSourceIdsUpdateField],
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const accessReviewCampaignId = this.getNodeParameter('accessReviewCampaignId', itemIndex) as string;
	const name = this.getNodeParameter('name', itemIndex, '') as string;
	const description = this.getNodeParameter('description', itemIndex, '') as string;
	const additionalFields = this.getNodeParameter('additionalFields', itemIndex, {}) as {
		accessReviewSourceIds?: string[];
	};

	const query = `
		mutation UpdateAccessReviewCampaign($input: UpdateAccessReviewCampaignInput!) {
			updateAccessReviewCampaign(input: $input) {
				accessReviewCampaign {
					id
					name
					description
					status
					startedAt
					completedAt
					createdAt
					updatedAt
				}
			}
		}
	`;

	const input: Record<string, string | string[]> = { accessReviewCampaignId };
	if (name) input.name = name;
	if (description) input.description = description;
	if (additionalFields.accessReviewSourceIds !== undefined) {
		input.accessReviewSourceIds = additionalFields.accessReviewSourceIds;
	}

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}
