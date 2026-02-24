import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  DropdownItem,
  IconCheckmark1,
  IconCrossLargeX,
  IconPencil,
} from "@probo/ui";
import { graphql } from "react-relay";

import type { RecordDecisionDropdownMutation } from "#/__generated__/core/RecordDecisionDropdownMutation.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

const recordDecisionMutation = graphql`
  mutation RecordDecisionDropdownMutation(
    $input: RecordAccessEntryDecisionInput!
  ) {
    recordAccessEntryDecision(input: $input) {
      accessEntry {
        id
        decision
        decidedAt
        decisionNote
        decidedBy {
          id
          fullName
        }
      }
    }
  }
`;

type Props = {
  accessEntryId: string;
};

export function RecordDecisionDropdown({ accessEntryId }: Props) {
  const { __ } = useTranslate();

  const [recordDecision, isRecording] = useMutationWithToasts<RecordDecisionDropdownMutation>(
    recordDecisionMutation,
    {
      successMessage: __("Decision recorded successfully."),
      errorMessage: __("Failed to record decision"),
    },
  );

  const handleDecision = (decision: "APPROVED" | "REVOKE" | "DEFER" | "ESCALATE") => {
    const needsNote = decision !== "APPROVED";
    let decisionNote: string | null = null;
    if (needsNote) {
      decisionNote = window.prompt(__("Add a justification note"));
      if (!decisionNote || decisionNote.trim().length === 0) {
        return;
      }
    }

    recordDecision({
      variables: {
        input: {
          accessEntryId,
          decision,
          decisionNote: decisionNote?.trim(),
        },
      },
    });
  };

  return (
    <ActionDropdown>
      <DropdownItem
        icon={IconCheckmark1}
        onSelect={(e) => {
          e.preventDefault();
          e.stopPropagation();
          handleDecision("APPROVED");
        }}
        disabled={isRecording}
      >
        {__("Approve")}
      </DropdownItem>
      <DropdownItem
        icon={IconCrossLargeX}
        variant="danger"
        onSelect={(e) => {
          e.preventDefault();
          e.stopPropagation();
          handleDecision("REVOKE");
        }}
        disabled={isRecording}
      >
        {__("Revoke")}
      </DropdownItem>
      <DropdownItem
        icon={IconPencil}
        onSelect={(e) => {
          e.preventDefault();
          e.stopPropagation();
          handleDecision("DEFER");
        }}
        disabled={isRecording}
      >
        {__("Defer")}
      </DropdownItem>
      <DropdownItem
        icon={IconPencil}
        onSelect={(e) => {
          e.preventDefault();
          e.stopPropagation();
          handleDecision("ESCALATE");
        }}
        disabled={isRecording}
      >
        {__("Escalate")}
      </DropdownItem>
    </ActionDropdown>
  );
}
