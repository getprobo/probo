import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";
import type { SCIMEventListItemFragment$key } from "/__generated__/iam/SCIMEventListItemFragment.graphql";
import { Td, Tr, Badge } from "@probo/ui";
import { formatDate } from "@probo/helpers";

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

const getStatusBadge = (statusCode: number) => {
  if (statusCode >= 200 && statusCode < 300) {
    return <Badge variant="success">{statusCode}</Badge>;
  } else if (statusCode >= 400 && statusCode < 500) {
    return <Badge variant="warning">{statusCode}</Badge>;
  } else if (statusCode >= 500) {
    return <Badge variant="danger">{statusCode}</Badge>;
  }
  return <Badge variant="neutral">{statusCode}</Badge>;
};

const getMethodBadge = (method: string) => {
  const variants: Record<string, "neutral" | "info" | "danger" | "highlight"> =
    {
      GET: "info",
      POST: "highlight",
      PUT: "highlight",
      PATCH: "highlight",
      DELETE: "danger",
    };
  return <Badge variant={variants[method] || "neutral"}>{method}</Badge>;
};

export function SCIMEventListItem(props: {
  fKey: SCIMEventListItemFragment$key;
}) {
  const { fKey } = props;

  const event = useFragment<SCIMEventListItemFragment$key>(
    SCIMEventListItemFragment,
    fKey
  );

  return (
    <Tr>
      <Td className="whitespace-nowrap">{formatDate(event.createdAt)}</Td>
      <Td>{getMethodBadge(event.method)}</Td>
      <Td className="font-mono text-xs">{event.path}</Td>
      <Td>{getStatusBadge(event.statusCode)}</Td>
      <Td>{event.membership?.profile?.fullName || "-"}</Td>
      <Td className="font-mono text-xs">{event.ipAddress}</Td>
      <Td className="max-w-xs truncate text-txt-danger">
        {event.errorMessage || "-"}
      </Td>
    </Tr>
  );
}
