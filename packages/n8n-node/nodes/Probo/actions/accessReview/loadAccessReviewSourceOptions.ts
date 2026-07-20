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
	query AccessReviewSources($organizationId: ID!) {
		organization: node(id: $organizationId) {
			... on Organization {
				accessReviewSources(first: 500) {
					edges {
						node {
							id
							name
						}
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
	const data = responseData.data as IDataObject | undefined;
	const organization = data?.organization as IDataObject | undefined;
	const accessReviewSources = organization?.accessReviewSources as IDataObject | undefined;
	const edges = accessReviewSources?.edges as Array<{ node: IDataObject }> | undefined;

	return (edges ?? [])
		.map((edge) => edge.node)
		.filter((node): node is IDataObject => node !== undefined && typeof node.id === 'string')
		.map((node) => ({
			name: String(node.name ?? node.id),
			value: String(node.id),
		}));
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

	const responseData = await proboApiRequest.call(this, organizationSourcesQuery, {
		organizationId,
	});

	return mapSources(responseData);
}
