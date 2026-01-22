import { formatDate } from "@probo/helpers";
import { Badge, IconChevronDown, IconChevronRight, Td, Tr } from "@probo/ui";
import { useState } from "react";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { SCIMEventListItemFragment$key } from "#/__generated__/iam/SCIMEventListItemFragment.graphql";

const SCIMEventListItemFragment = graphql`
  fragment SCIMEventListItemFragment on SCIMEvent {
    id
    method
    path
    statusCode
    errorMessage
    ipAddress
    createdAt
    membership {
      id
      profile {
        fullName
      }
    }
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
                  <span>{event.membership?.profile?.fullName || "-"}</span>
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
