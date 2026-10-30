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
  getRequest,
  type GraphQLResponse,
  Network,
  RecordSource,
  Store,
} from "relay-runtime";
import { describe, expect, it, vi } from "vitest";

import type { usePendingCountRefreshQuery } from "#/__generated__/core/usePendingCountRefreshQuery.graphql";

import { createPendingCountRefresh, pendingCountQuery } from "./usePendingCountRefresh";

// Vitest has no Relay Babel transform, so the graphql tag returns the compiled artifact.
vi.mock("relay-runtime", async (importOriginal) => {
  const { default: query } = await import("#/__generated__/core/usePendingCountRefreshQuery.graphql");
  return { ...await importOriginal<typeof import("relay-runtime")>(), graphql: () => query };
});

const campaignId = "campaign-1";

const campaignResponse = (pendingCount: number): GraphQLResponse => ({
  data: {
    node: {
      __typename: "AccessReviewCampaign",
      id: campaignId,
      status: "PENDING_ACTIONS",
      pendingEntries: { totalCount: pendingCount },
    },
  },
});

const setup = () => {
  const responses: { resolve: (response: GraphQLResponse) => void; reject: (error: Error) => void }[] = [];
  const environment = new Environment({
    network: Network.create(() => new Promise((resolve, reject) => responses.push({ resolve, reject }))),
    store: new Store(new RecordSource()),
  });
  const pendingCount = () => {
    const operation = createOperationDescriptor(getRequest(pendingCountQuery), { campaignId });
    const { node } = environment.lookup(operation.fragment).data as usePendingCountRefreshQuery["response"];
    return node?.pendingEntries?.totalCount;
  };
  return { responses, environment, pendingCount };
};

const settle = () => new Promise(resolve => setTimeout(resolve, 0));

describe("createPendingCountRefresh", () => {
  it("refreshes a count left stale by an earlier visit", async () => {
    const { responses, environment, pendingCount } = setup();
    const previous = createPendingCountRefresh(environment, campaignId, true, () => {});
    responses[0].resolve(campaignResponse(1));
    await settle();
    previous.dispose();

    createPendingCountRefresh(environment, campaignId, true, () => {});
    await settle();
    expect(responses).toHaveLength(2);

    responses[1].resolve(campaignResponse(0));
    await settle();
    expect(pendingCount()).toBe(0);
  });

  it("waits for a call when not asked to refresh on creation", async () => {
    const { responses, environment } = setup();
    const { refresh } = createPendingCountRefresh(environment, campaignId, false, () => {});
    await settle();
    expect(responses).toHaveLength(0);

    refresh();
    await settle();
    expect(responses).toHaveLength(1);
  });

  it("queues a call made while a request is in flight", async () => {
    const { responses, environment, pendingCount } = setup();
    const { refresh } = createPendingCountRefresh(environment, campaignId, true, () => {});

    refresh();
    refresh();
    await settle();
    expect(responses).toHaveLength(1);

    responses[0].resolve(campaignResponse(1));
    await settle();
    expect(responses).toHaveLength(2);

    responses[1].resolve(campaignResponse(0));
    await settle();
    expect(pendingCount()).toBe(0);
  });

  it("drops the in-flight response and later calls once disposed", async () => {
    const { responses, environment, pendingCount } = setup();
    const previous = createPendingCountRefresh(environment, campaignId, true, () => {});
    previous.dispose();
    previous.refresh();
    createPendingCountRefresh(environment, campaignId, true, () => {});
    await settle();
    expect(responses).toHaveLength(2);

    responses[1].resolve(campaignResponse(0));
    responses[0].resolve(campaignResponse(1));
    await settle();
    expect(pendingCount()).toBe(0);
  });

  it("reports a failed refresh and accepts the next call", async () => {
    const { responses, environment } = setup();
    const onError = vi.fn();
    const { refresh } = createPendingCountRefresh(environment, campaignId, true, onError);

    responses[0].reject(new Error("boom"));
    await settle();
    expect(onError).toHaveBeenCalledTimes(1);

    refresh();
    await settle();
    expect(responses).toHaveLength(2);
  });

  it("leaves the campaign status alone", async () => {
    const { responses, environment } = setup();
    environment.commitUpdate((store) => {
      store.create(campaignId, "AccessReviewCampaign").setValue("COMPLETED", "status");
    });
    createPendingCountRefresh(environment, campaignId, true, () => {});

    await settle();
    responses[0].resolve(campaignResponse(0));
    await settle();

    expect(environment.getStore().getSource().get(campaignId)?.status).toBe("COMPLETED");
  });
});
