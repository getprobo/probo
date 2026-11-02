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

import { useEffect, useRef, useState } from "react";
import { fetchQuery, graphql, useRelayEnvironment } from "react-relay";

import type { useSlackChannelPageQuery } from "#/__generated__/core/useSlackChannelPageQuery.graphql";

const slackChannelPollIntervalMs = 5000;

export interface SlackChannel {
  id: string;
  name: string;
}

interface ChannelRequest {
  cursor: string | null;
  mode: "replace" | "append";
  reset: boolean;
  id: number;
}

const slackChannelPageQuery = graphql`
  query useSlackChannelPageQuery($organizationId: ID!, $cursor: String) {
    node(id: $organizationId) {
      __typename
      ... on Organization {
        slackbotChannels(cursor: $cursor) {
          channels {
            id
            name
          }
          nextCursor
        }
      }
    }
  }
`;

export function useSlackChannelPage(organizationId: string, enabled: boolean) {
  const environment = useRelayEnvironment();
  const scopeKey = `${organizationId}:${String(enabled)}`;
  const [scope, setScope] = useState(scopeKey);
  const [channels, setChannels] = useState<SlackChannel[]>([]);
  const [nextCursor, setNextCursor] = useState<string | null>(null);
  const [request, setRequest] = useState<ChannelRequest>({
    cursor: null,
    mode: "replace",
    reset: true,
    id: 0,
  });
  const [settledId, setSettledId] = useState<number | null>(null);
  const [failedId, setFailedId] = useState<number | null>(null);
  const pagedBeyondFirstRef = useRef(false);
  const nextCursorRef = useRef<string | null>(null);
  const loadingRef = useRef(false);
  const appendingRef = useRef(false);
  const loading = enabled && settledId !== request.id;
  const loadingMore = loading && request.mode === "append";
  const failed = failedId === request.id && request.mode !== "append";
  const failedMore = failedId === request.id && request.mode === "append";

  if (scope !== scopeKey) {
    setScope(scopeKey);
    setChannels([]);
    setNextCursor(null);
    setSettledId(null);
    setFailedId(null);
    setRequest(current => ({
      cursor: null,
      mode: "replace",
      reset: true,
      id: current.id + 1,
    }));
  }

  useEffect(() => {
    loadingRef.current = loading;
  }, [loading]);

  useEffect(() => {
    if (!enabled) {
      return;
    }

    if (request.mode !== "append") {
      appendingRef.current = false;
    }

    if (request.reset) {
      pagedBeyondFirstRef.current = false;
      nextCursorRef.current = null;
    }

    let active = true;
    const subscription = fetchQuery<useSlackChannelPageQuery>(
      environment,
      slackChannelPageQuery,
      { organizationId, cursor: request.cursor },
      { fetchPolicy: "network-only" },
    ).subscribe({
      next: (data) => {
        if (!active) {
          return;
        }

        const page = readPage(data);
        const pagedBeyondFirst = pagedBeyondFirstRef.current;

        setChannels((current) => {
          if (request.mode === "append") {
            return dedupe([...current, ...page.channels]);
          }
          // Later pages were loaded from the previous cursor. A channel
          // that left page 1 is not in this response or in those pages.
          if (pagedBeyondFirst) {
            return dedupe([...page.channels, ...current]);
          }
          return page.channels;
        });

        if (request.mode === "append" || !pagedBeyondFirst) {
          nextCursorRef.current = page.nextCursor;
          setNextCursor(page.nextCursor);
        }

        if (request.mode === "append") {
          appendingRef.current = false;
        }

        setFailedId(null);
        setSettledId(request.id);
      },
      error: () => {
        if (!active) {
          return;
        }

        if (request.mode === "append") {
          appendingRef.current = false;
        }

        setFailedId(request.id);
        setSettledId(request.id);
      },
    });

    return () => {
      active = false;
      subscription.unsubscribe();
    };
  }, [enabled, environment, organizationId, request]);

  useEffect(() => {
    if (!enabled) {
      return;
    }

    const refresh = () => {
      if (document.hidden || loadingRef.current) {
        return;
      }

      setRequest(current => ({
        cursor: null,
        mode: "replace",
        reset: false,
        id: current.id + 1,
      }));
    };
    const interval = window.setInterval(refresh, slackChannelPollIntervalMs);
    document.addEventListener("visibilitychange", refresh);

    return () => {
      window.clearInterval(interval);
      document.removeEventListener("visibilitychange", refresh);
    };
  }, [enabled]);

  function onLoadMore() {
    const cursor = nextCursorRef.current;
    if (cursor == null || appendingRef.current) {
      return;
    }

    appendingRef.current = true;
    pagedBeyondFirstRef.current = true;
    setRequest(current => ({
      cursor,
      mode: "append",
      reset: false,
      id: current.id + 1,
    }));
  }

  return {
    channels,
    loading,
    loadingMore,
    failed,
    failedMore,
    hasNext: nextCursor != null,
    onLoadMore,
  };
}

function readPage(data: useSlackChannelPageQuery["response"]): {
  channels: SlackChannel[];
  nextCursor: string | null;
} {
  const organization = data.node?.__typename === "Organization" ? data.node : null;
  const page = organization?.slackbotChannels;
  return {
    channels: page?.channels.map(channel => ({
      id: channel.id,
      name: channel.name,
    })) ?? [],
    nextCursor: page?.nextCursor ?? null,
  };
}

function dedupe(channels: SlackChannel[]): SlackChannel[] {
  const seen = new Set<string>();
  const unique: SlackChannel[] = [];
  for (const channel of channels) {
    if (seen.has(channel.id)) {
      continue;
    }
    seen.add(channel.id);
    unique.push(channel);
  }
  return unique;
}
