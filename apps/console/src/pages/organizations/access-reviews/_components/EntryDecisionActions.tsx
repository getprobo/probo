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

import { formatError } from "@probo/helpers";
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
import { useTranslation } from "react-i18next";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { AccessReviewEntryDecision, EntryDecisionActionsMutation } from "#/__generated__/core/EntryDecisionActionsMutation.graphql";

import { decisionBadgeVariant } from "./accessReviewHelpers";

const mutation = graphql`
  mutation EntryDecisionActionsMutation(
    $input: RecordAccessReviewEntryDecisionInput!
  ) {
    recordAccessReviewEntryDecision(input: $input) {
      accessReviewEntry {
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
  const { t } = useTranslation();
  const { toast } = useToast();
  const ref = useDialogRef();
  const [editing, setEditing] = useState(false);
  const [pendingDecision, setPendingDecision] = useState<AccessReviewEntryDecision | null>(null);
  const [note, setNote] = useState("");
  const [recordDecision, isRecording]
    = useMutation<EntryDecisionActionsMutation>(mutation);

  const submitDecision = (decisionValue: AccessReviewEntryDecision, decisionNote?: string) => {
    recordDecision({
      variables: {
        input: {
          accessReviewEntryId: entryId,
          decision: decisionValue,
          decisionNote: decisionNote || null,
        },
      },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast({
            title: t("entryDecisionActions.messages.error"),
            description: formatError(
              t("entryDecisionActions.errors.record"),
              errors,
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
          title: t("entryDecisionActions.messages.error"),
          description: formatError(
            t("entryDecisionActions.errors.record"),
            error,
          ),
          variant: "error",
        });
      },
    });
  };

  const openNoteDialog = (decisionValue: AccessReviewEntryDecision) => {
    setPendingDecision(decisionValue);
    setNote("");
    ref.current?.open();
  };

  const handleDecision = (value: string) => {
    const decision = value as AccessReviewEntryDecision;
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
          {t(`entryDecisionActions.decisions.${decision.toLowerCase()}`)}
        </Badge>
        <button
          type="button"
          className="text-txt-tertiary hover:text-txt-primary cursor-pointer"
          onClick={() => setEditing(true)}
          title={t("entryDecisionActions.actions.change")}
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
        placeholder={t("entryDecisionActions.placeholder")}
        onValueChange={handleDecision}
        disabled={isRecording}
      >
        <Option value="APPROVED">
          {t("entryDecisionActions.actions.approve")}
        </Option>
        <Option value="REVOKE">
          {t("entryDecisionActions.actions.revoke")}
        </Option>
        <Option value="DEFER">
          {t("entryDecisionActions.actions.modify")}
        </Option>
        <Option value="ESCALATE">
          {t("entryDecisionActions.actions.escalate")}
        </Option>
      </Select>

      <Dialog ref={ref} title={t("entryDecisionActions.note.title")}>
        <DialogContent padded className="space-y-4">
          <p className="text-sm text-txt-secondary">
            {t("entryDecisionActions.note.description")}
          </p>
          <Field
            label={t("entryDecisionActions.note.label")}
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
            {t("entryDecisionActions.actions.confirm")}
          </Button>
        </DialogFooter>
      </Dialog>
    </>
  );
}
