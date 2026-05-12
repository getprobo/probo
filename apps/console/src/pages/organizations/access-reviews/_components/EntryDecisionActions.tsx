// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { formatError, type GraphQLError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Badge,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  IconPencil,
  Option,
  Select,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { useState } from "react";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { AccessEntryDecision, EntryDecisionActionsMutation } from "#/__generated__/core/EntryDecisionActionsMutation.graphql";

import { decisionBadgeVariant, decisionLabel } from "./accessReviewHelpers";

const mutation = graphql`
  mutation EntryDecisionActionsMutation(
    $input: RecordAccessEntryDecisionInput!
  ) {
    recordAccessEntryDecision(input: $input) {
      accessEntry {
        id
        decision
        decisionNote
      }
    }
  }
`;

type Props = {
  entryId: string;
  decision: string;
};

export function EntryDecisionActions({ entryId, decision }: Props) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const ref = useDialogRef();
  const [editing, setEditing] = useState(false);
  const [pendingDecision, setPendingDecision] = useState<AccessEntryDecision | null>(null);
  const [note, setNote] = useState("");
  const [recordDecision, isRecording]
    = useMutation<EntryDecisionActionsMutation>(mutation);

  const submitDecision = (decisionValue: AccessEntryDecision, decisionNote?: string) => {
    recordDecision({
      variables: {
        input: {
          accessEntryId: entryId,
          decision: decisionValue,
          decisionNote: decisionNote || null,
        },
      },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast({
            title: __("Error"),
            description: formatError(
              __("Failed to record decision"),
              errors as GraphQLError[],
            ),
            variant: "error",
          });
          return;
        }
        setPendingDecision(null);
        setNote("");
        setEditing(false);
        ref.current?.close();
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(
            __("Failed to record decision"),
            error as GraphQLError,
          ),
          variant: "error",
        });
      },
    });
  };

  const openNoteDialog = (decisionValue: AccessEntryDecision) => {
    setPendingDecision(decisionValue);
    setNote("");
    ref.current?.open();
  };

  const handleDecision = (value: string) => {
    const decision = value as AccessEntryDecision;
    if (decision === "APPROVED") {
      submitDecision(decision);
    } else {
      openNoteDialog(decision);
    }
  };

  // Already decided -- show badge with edit button
  if (decision !== "PENDING" && !editing) {
    return (
      <div className="flex items-center gap-1">
        <Badge variant={decisionBadgeVariant(decision)}>
          {decisionLabel(__, decision)}
        </Badge>
        <button
          type="button"
          className="text-txt-tertiary hover:text-txt-primary cursor-pointer"
          onClick={() => setEditing(true)}
          title={__("Change decision")}
        >
          <IconPencil size={14} />
        </button>
      </div>
    );
  }

  return (
    <>
      <Select
        variant="editor"
        placeholder={__("Decide...")}
        onValueChange={handleDecision}
        disabled={isRecording}
      >
        <Option value="APPROVED">{__("Approve")}</Option>
        <Option value="REVOKE">{__("Revoke")}</Option>
        <Option value="DEFER">{__("Modify")}</Option>
        <Option value="ESCALATE">{__("Escalate")}</Option>
      </Select>

      <Dialog ref={ref} title={__("Decision note")}>
        <DialogContent padded className="space-y-4">
          <p className="text-sm text-txt-secondary">
            {__("Please provide a reason for this decision.")}
          </p>
          <Field
            label={__("Note")}
            type="textarea"
            value={note}
            onValueChange={setNote}
          />
        </DialogContent>
        <DialogFooter>
          <Button
            disabled={isRecording || !note.trim()}
            onClick={() => {
              if (pendingDecision) {
                submitDecision(pendingDecision, note);
              }
            }}
          >
            {__("Confirm")}
          </Button>
        </DialogFooter>
      </Dialog>
    </>
  );
}
