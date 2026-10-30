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

import {
  createOperationDescriptor,
  Environment,
  getSingularSelector,
  Network,
  RecordSource,
  Store,
} from "relay-runtime";
import { describe, expect, it } from "vitest";

import accessEntrySectionFragment from "#/__generated__/core/AccessEntrySection_source.graphql";
import accessEntrySourceSectionFragment from "#/__generated__/core/AccessEntrySourceSection_source.graphql";
import addCampaignSourceMutation from "#/__generated__/core/AddCampaignSourceDialogMutation.graphql";
import campaignDetailPageQuery, {
  type CampaignDetailPageQuery$data,
} from "#/__generated__/core/CampaignDetailPageQuery.graphql";

const campaignId = "campaign-1";

describe("AddCampaignSourceDialogMutation", () => {
  it("stores every field the campaign page reads for the added source", () => {
    const environment = new Environment({
      network: Network.create(() => {
        throw new Error("unexpected network request");
      }),
      store: new Store(new RecordSource()),
    });

    const pageOperation = createOperationDescriptor(
      campaignDetailPageQuery,
      { campaignId },
    );
    environment.commitPayload(pageOperation, {
      node: {
        __typename: "AccessReviewCampaign",
        id: campaignId,
        name: "Q4 review",
        status: "DRAFT",
        pendingEntries: { totalCount: 0 },
        canDelete: true,
        canRetrySourceFetch: true,
        sources: [],
      },
    });

    const mutationOperation = createOperationDescriptor(
      addCampaignSourceMutation,
      {
        input: {
          accessReviewCampaignId: campaignId,
          accessReviewSourceId: "source-1",
        },
      },
    );
    environment.commitPayload(mutationOperation, {
      addAccessReviewCampaignSource: {
        accessReviewCampaign: {
          id: campaignId,
          sources: [
            {
              id: "campaign-source-1",
              name: "Slack",
              source: {
                id: "source-1",
                name: "Slack",
                connector: { id: "connector-1", provider: "SLACK" },
                connectorAccount: { id: "account-1", name: "Acme" },
              },
              fetchAttempts: { edges: [] },
              entries: {
                totalCount: 0,
                edges: [],
                pageInfo: { endCursor: null, hasNextPage: false },
              },
            },
          ],
        },
      },
    });

    const page = environment.lookup(pageOperation.fragment);
    expect(page.isMissingData).toBe(false);

    const { node } = page.data as CampaignDetailPageQuery$data;
    if (node.__typename !== "AccessReviewCampaign") {
      throw new Error("expected an AccessReviewCampaign node");
    }
    const [source] = node.sources;
    const sourceSection = environment.lookup(
      getSingularSelector(accessEntrySourceSectionFragment, source),
    );
    expect(sourceSection.isMissingData).toBe(false);

    const section = environment.lookup(
      getSingularSelector(accessEntrySectionFragment, sourceSection.data),
    );
    expect(section.isMissingData).toBe(false);
  });
});
