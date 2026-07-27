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

import type { IDataObject, ILoadOptionsFunctions, INodePropertyOptions } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

const organizationSourcesQuery = `
	query AccessReviewSources($organizationId: ID!, $after: CursorKey) {
		organization: node(id: $organizationId) {
			... on Organization {
				accessReviewSources(first: 100, after: $after) {
					edges {
						node {
							id
							name
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

const campaignOrganizationQuery = `
	query AccessReviewCampaignOrganization($accessReviewCampaignId: ID!) {
		campaign: node(id: $accessReviewCampaignId) {
			... on AccessReviewCampaign {
				organization {
					id
				}
			}
		}
	}
`;

function mapSources(responseData: IDataObject): INodePropertyOptions[] {
	const edges = getAccessReviewSourcesConnection(responseData)?.edges as
		| Array<{ node: IDataObject }>
		| undefined;

	return (edges ?? [])
		.map((edge) => edge.node)
		.filter((node): node is IDataObject => node !== undefined && typeof node.id === 'string')
		.map((node) => ({
			name: String(node.name ?? node.id),
			value: String(node.id),
		}));
}

function getAccessReviewSourcesConnection(responseData: IDataObject): IDataObject | undefined {
	const data = responseData.data as IDataObject | undefined;
	const organization = data?.organization as IDataObject | undefined;

	return organization?.accessReviewSources as IDataObject | undefined;
}

function getPageInfo(responseData: IDataObject): { hasNextPage: boolean; endCursor?: string } {
	const pageInfo = getAccessReviewSourcesConnection(responseData)?.pageInfo as
		| IDataObject
		| undefined;

	return {
		hasNextPage: pageInfo?.hasNextPage === true,
		endCursor: typeof pageInfo?.endCursor === 'string' ? pageInfo.endCursor : undefined,
	};
}

async function resolveOrganizationId(
	this: ILoadOptionsFunctions,
): Promise<string | undefined> {
	const organizationId = this.getCurrentNodeParameter('organizationId') as string | undefined;
	if (organizationId) {
		return organizationId;
	}

	const accessReviewCampaignId = this.getCurrentNodeParameter('accessReviewCampaignId') as string | undefined;
	if (!accessReviewCampaignId) {
		return undefined;
	}

	const responseData = await proboApiRequest.call(this, campaignOrganizationQuery, {
		accessReviewCampaignId,
	});
	const data = responseData.data as IDataObject | undefined;
	const campaign = data?.campaign as IDataObject | undefined;
	const organization = campaign?.organization as IDataObject | undefined;

	return typeof organization?.id === 'string' ? organization.id : undefined;
}

export async function getAccessReviewSources(
	this: ILoadOptionsFunctions,
): Promise<INodePropertyOptions[]> {
	const organizationId = await resolveOrganizationId.call(this);
	if (!organizationId) {
		return [];
	}

	const options: INodePropertyOptions[] = [];
	let after: string | undefined;

	do {
		const responseData = await proboApiRequest.call(this, organizationSourcesQuery, {
			organizationId,
			after,
		});

		options.push(...mapSources(responseData));

		const pageInfo = getPageInfo(responseData);
		after = pageInfo.hasNextPage ? pageInfo.endCursor : undefined;
	} while (after);

	return options;
}
