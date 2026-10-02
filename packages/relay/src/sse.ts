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

import { GraphQLError } from "graphql";

// Relay treats a payload as @defer only when its label contains this marker,
// and finishes the operation when extensions.is_final is set.
const deferLabelMarker = "$defer$";

export type GraphQLSSEPayload = {
  data?: Record<string, unknown> | null;
  errors?: GraphQLError[];
  extensions?: Record<string, unknown>;
  label?: string;
  path?: Array<string | number>;
  hasNext?: boolean;
};

export type SSEPush = {
  buffer: string;
  payloads: GraphQLSSEPayload[];
  completed: boolean;
};

// pushSSE consumes the next chunk of a gqlgen text/event-stream body.
// Events are `event: next` with one GraphQL response, then `event: complete`.
export function pushSSE(buffer: string, chunk: string): SSEPush {
  const combined = (buffer + chunk).replaceAll("\r\n", "\n");
  const payloads: GraphQLSSEPayload[] = [];
  let rest = combined;
  let completed = false;

  while (!completed) {
    const boundary = rest.indexOf("\n\n");
    if (boundary === -1) {
      break;
    }

    const raw = rest.slice(0, boundary);
    rest = rest.slice(boundary + 2);
    const event = parseSSEEvent(raw);
    if (event == null) {
      continue;
    }
    if (event.event === "complete") {
      completed = true;
      break;
    }
    if (event.event !== "next" || event.data === "") {
      continue;
    }

    payloads.push(markRelayFinal(JSON.parse(event.data) as GraphQLSSEPayload));
  }

  return { buffer: rest, payloads, completed };
}

function parseSSEEvent(raw: string): { event: string; data: string } | null {
  let event = "message";
  const data: string[] = [];
  let meaningful = false;

  for (const line of raw.split("\n")) {
    if (line === "" || line.startsWith(":")) {
      continue;
    }

    meaningful = true;
    if (line.startsWith("event:")) {
      event = line.slice("event:".length).trim();
      continue;
    }
    if (line.startsWith("data:")) {
      data.push(line.slice("data:".length).trimStart());
    }
  }

  if (!meaningful) {
    return null;
  }

  return { event, data: data.join("\n") };
}

function markRelayFinal(payload: GraphQLSSEPayload): GraphQLSSEPayload {
  const incremental = payload.label != null
    && payload.label.includes(deferLabelMarker)
    && payload.path != null;
  if (!incremental || payload.hasNext !== false) {
    return payload;
  }

  return {
    ...payload,
    extensions: {
      ...payload.extensions,
      is_final: true,
    },
  };
}
