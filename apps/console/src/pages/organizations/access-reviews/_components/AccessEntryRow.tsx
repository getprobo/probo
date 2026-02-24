import {
  Badge,
  Td,
  Tr,
} from "@probo/ui";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { AccessEntryRowFragment$key } from "#/__generated__/core/AccessEntryRowFragment.graphql";

import { RecordDecisionDropdown } from "../dialogs/RecordDecisionDropdown";

const fragment = graphql`
  fragment AccessEntryRowFragment on AccessEntry {
    id
    email
    fullName
    role
    flag
    decision
    decisionNote
    incrementalTag
    mfaStatus
    authMethod
    canDecide: permission(action: "core:access-entry:decide")
  }
`;

function flagBadgeVariant(flag: string) {
  switch (flag) {
    case "ORPHANED":
      return "danger" as const;
    case "INACTIVE":
      return "warning" as const;
    case "EXCESSIVE":
      return "warning" as const;
    case "ROLE_MISMATCH":
      return "warning" as const;
    case "NEW":
      return "info" as const;
    default:
      return "neutral" as const;
  }
}

function decisionBadgeVariant(decision: string) {
  switch (decision) {
    case "APPROVED":
      return "success" as const;
    case "REVOKE":
      return "danger" as const;
    case "DEFER":
      return "warning" as const;
    case "ESCALATE":
      return "warning" as const;
    default:
      return "neutral" as const;
  }
}

type Props = {
  fKey: AccessEntryRowFragment$key;
};

export function AccessEntryRow({ fKey }: Props) {
  const entry = useFragment(fragment, fKey);

  return (
    <Tr>
      <Td>{entry.fullName}</Td>
      <Td>{entry.email}</Td>
      <Td>{entry.role || "-"}</Td>
      <Td>{entry.incrementalTag}</Td>
      <Td>
        {entry.flag && entry.flag !== "NONE"
          ? (
              <Badge variant={flagBadgeVariant(entry.flag)} size="sm">
                {entry.flag}
              </Badge>
            )
          : "-"}
      </Td>
      <Td>
        <Badge variant={decisionBadgeVariant(entry.decision)} size="sm">
          {entry.decision}
        </Badge>
      </Td>
      <Td>{entry.decisionNote || "-"}</Td>
      <Td noLink width={50} className="text-end">
        {entry.canDecide && entry.decision === "PENDING" && (
          <RecordDecisionDropdown accessEntryId={entry.id} />
        )}
      </Td>
    </Tr>
  );
}
