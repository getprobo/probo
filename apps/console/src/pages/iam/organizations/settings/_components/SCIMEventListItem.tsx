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

import { formatDate } from "@probo/helpers";
import { Badge, IconChevronDown, IconChevronRight, Td, Tr } from "@probo/ui";
import { useState } from "react";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { SCIMEventListItemFragment$key } from "#/__generated__/iam/SCIMEventListItemFragment.graphql";

const SCIMEventListItemFragment = graphql`
  fragment SCIMEventListItemFragment on SCIMEvent {
    method
    path
    statusCode
    errorMessage
    ipAddress
    createdAt
    userName
  }
`;

const getResultBadge = (statusCode: number) => {
  if (statusCode >= 200 && statusCode < 300) {
    return <Badge variant="success">Info</Badge>;
  }
  return <Badge variant="danger">Error</Badge>;
};

const getMethodBadge = (method: string) => {
  const variants: Record<string, "neutral" | "info" | "danger" | "highlight">
    = {
      GET: "info",
      POST: "highlight",
      PUT: "highlight",
      PATCH: "highlight",
      DELETE: "danger",
    };
  return <Badge variant={variants[method] || "neutral"}>{method}</Badge>;
};

const decodePath = (path: string): string => {
  try {
    return decodeURIComponent(path);
  } catch {
    return path;
  }
};

export function SCIMEventListItem(props: {
  fKey: SCIMEventListItemFragment$key;
}) {
  const { fKey } = props;
  const [isExpanded, setIsExpanded] = useState(false);

  const event = useFragment<SCIMEventListItemFragment$key>(
    SCIMEventListItemFragment,
    fKey,
  );

  const hasError = !!event.errorMessage;

  return (
    <>
      <Tr
        className="cursor-pointer hover:bg-bg-hover"
        onClick={() => setIsExpanded(!isExpanded)}
      >
        <Td className="whitespace-nowrap">
          <div className="flex items-center gap-2">
            {isExpanded
              ? (
                  <IconChevronDown size={16} className="text-txt-secondary" />
                )
              : (
                  <IconChevronRight size={16} className="text-txt-secondary" />
                )}
            {formatDate(event.createdAt)}
          </div>
        </Td>
        <Td>{getMethodBadge(event.method)}</Td>
        <Td className="font-mono text-xs">{decodePath(event.path)}</Td>
        <Td>{getResultBadge(event.statusCode)}</Td>
      </Tr>
      {isExpanded && (
        <Tr>
          <Td colSpan={4} className="bg-bg-subtle">
            <div className="py-2 pl-6 space-y-2">
              <div className="flex gap-8 text-sm">
                <div>
                  <span className="text-txt-secondary">User: </span>
                  <span>{event.userName || "-"}</span>
                </div>
                <div>
                  <span className="text-txt-secondary">IP Address: </span>
                  <span className="font-mono text-xs">{event.ipAddress}</span>
                </div>
                <div>
                  <span className="text-txt-secondary">Status Code: </span>
                  <span className="font-mono text-xs">{event.statusCode}</span>
                </div>
              </div>
              {hasError && (
                <div>
                  <div className="text-sm text-txt-secondary">Error</div>
                  <pre className="mt-1 whitespace-pre-wrap break-all font-mono text-xs text-txt-danger">
                    {event.errorMessage}
                  </pre>
                </div>
              )}
            </div>
          </Td>
        </Tr>
      )}
    </>
  );
}
