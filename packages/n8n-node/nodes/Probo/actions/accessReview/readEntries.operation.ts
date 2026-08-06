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

import type { IDataObject, IExecuteFunctions, INodeExecutionData, INodeProperties } from 'n8n-workflow';
import { proboApiRequestAllItems } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Access Review Campaign ID',
		name: 'accessReviewCampaignId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['accessReview'],
				operation: ['readEntries'],
			},
		},
		default: '',
		description: 'The ID of the access review campaign',
		required: true,
	},
	{
		displayName: 'Return All',
		name: 'returnAll',
		type: 'boolean',
		displayOptions: {
			show: {
				resource: ['accessReview'],
				operation: ['readEntries'],
			},
		},
		default: false,
		description: 'Whether to return all results or only up to a given limit',
	},
	{
		displayName: 'Limit',
		name: 'limit',
		type: 'number',
		displayOptions: {
			show: {
				resource: ['accessReview'],
				operation: ['readEntries'],
				returnAll: [false],
			},
		},
		typeOptions: {
			minValue: 1,
		},
		default: 50,
		description: 'Max number of results to return',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const accessReviewCampaignId = this.getNodeParameter('accessReviewCampaignId', itemIndex) as string;
	const returnAll = this.getNodeParameter('returnAll', itemIndex) as boolean;
	const limit = this.getNodeParameter('limit', itemIndex, 50) as number;

	const query = `
		query ReadAccessReviewCampaignEntries(
			$accessReviewCampaignId: ID!
			$first: Int
			$after: CursorKey
		) {
			node(id: $accessReviewCampaignId) {
				... on AccessReviewCampaign {
					entries(first: $first, after: $after) {
						edges {
							node {
								id
								email
								fullName
								roles
								jobTitle
								isAdmin
								active
								mfaStatus
								authMethod
								accountType
								lastLogin
								accountCreatedAt
								externalId
								incrementalTag
								flags
								flagReasons
								decision
								decisionNote
								decidedBy
								decidedAt
								campaignSource {
									id
									name
								}
								createdAt
								updatedAt
							}
						}
						pageInfo {
							hasNextPage
							endCursor
						}
					}
				}
			}
		}
	`;

	const accessReviewEntries = await proboApiRequestAllItems.call(
		this,
		query,
		{ accessReviewCampaignId },
		(response) => {
			const data = response?.data as IDataObject | undefined;
			const node = data?.node as IDataObject | undefined;
			return node?.entries as IDataObject | undefined;
		},
		returnAll,
		limit,
	);

	return {
		json: { accessReviewEntries },
		pairedItem: { item: itemIndex },
	};
}
