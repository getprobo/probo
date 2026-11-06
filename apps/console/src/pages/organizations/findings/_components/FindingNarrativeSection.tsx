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

import { Field } from "@probo/ui/src/v2/form/Field";
import { Textarea } from "@probo/ui/src/v2/form/Textarea";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { type Dispatch, type SetStateAction, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { FindingNarrativeSection_finding$key } from "#/__generated__/core/FindingNarrativeSection_finding.graphql";

import { useSaveQueue } from "../_lib/useSaveQueue";
import { useUpdateFinding } from "../_lib/useUpdateFinding";

import { findingsPage } from "./variants";

const findingNarrativeFragment = graphql`
  fragment FindingNarrativeSection_finding on Finding {
    id
    source
    rootCause
    correctiveAction
    effectivenessCheck
    description
    canUpdate: permission(action: "core:finding:update")
  }
`;

type NarrativeField = "source" | "rootCause" | "correctiveAction" | "effectivenessCheck" | "description";

interface FindingNarrativeSectionProps {
  findingKey: FindingNarrativeSection_finding$key;
}

export function FindingNarrativeSection({ findingKey }: FindingNarrativeSectionProps) {
  const finding = useFragment(findingNarrativeFragment, findingKey);
  const { t } = useTranslation("organizations/findings");
  const [updateFinding] = useUpdateFinding();
  const canUpdate = finding.canUpdate;
  const [source, setSource] = useState(finding.source ?? "");
  const [rootCause, setRootCause] = useState(finding.rootCause ?? "");
  const [correctiveAction, setCorrectiveAction] = useState(finding.correctiveAction ?? "");
  const [effectivenessCheck, setEffectivenessCheck] = useState(finding.effectivenessCheck ?? "");
  const [description, setDescription] = useState(finding.description ?? "");
  const queued = useRef<Record<NarrativeField, string | null>>({
    source: finding.source ?? null,
    rootCause: finding.rootCause ?? null,
    correctiveAction: finding.correctiveAction ?? null,
    effectivenessCheck: finding.effectivenessCheck ?? null,
    description: finding.description ?? null,
  });
  const enqueue = useSaveQueue();
  const { narrative, narrativeWide } = findingsPage();
  const setters: Record<NarrativeField, Dispatch<SetStateAction<string>>> = {
    source: setSource,
    rootCause: setRootCause,
    correctiveAction: setCorrectiveAction,
    effectivenessCheck: setEffectivenessCheck,
    description: setDescription,
  };

  function saveText(field: NarrativeField, value: string) {
    const next = value.trim() === "" ? null : value;
    const setValue = setters[field];
    if (next === queued.current[field]) {
      if (next === null) {
        setValue(current => (current === value ? "" : current));
      }
      return;
    }
    const previous = queued.current[field];
    queued.current[field] = next;
    void enqueue(() => updateFinding({
      variables: {
        input: {
          id: finding.id,
          [field]: next,
        },
      },
    })).then(() => {
      if (next === null) {
        setValue(current => (current === value ? "" : current));
      }
    }).catch(() => {
      queued.current[field] = previous;
      setValue(current => (current === value ? previous ?? "" : current));
    });
  }

  return (
    <div className={narrative()}>
      <Field className={narrativeWide()} label={t("details.fields.source")}>
        <TextField
          value={source}
          disabled={!canUpdate}
          onValueChange={setSource}
          onBlur={() => saveText("source", source)}
        />
      </Field>
      <Field label={t("details.fields.rootCause")}>
        <Textarea
          value={rootCause}
          rows={4}
          disabled={!canUpdate}
          onChange={event => setRootCause(event.currentTarget.value)}
          onBlur={() => saveText("rootCause", rootCause)}
        />
      </Field>
      <Field label={t("details.fields.correctiveAction")}>
        <Textarea
          value={correctiveAction}
          rows={4}
          disabled={!canUpdate}
          onChange={event => setCorrectiveAction(event.currentTarget.value)}
          onBlur={() => saveText("correctiveAction", correctiveAction)}
        />
      </Field>
      <Field className={narrativeWide()} label={t("details.fields.effectivenessCheck")}>
        <Textarea
          value={effectivenessCheck}
          rows={4}
          disabled={!canUpdate}
          onChange={event => setEffectivenessCheck(event.currentTarget.value)}
          onBlur={() => saveText("effectivenessCheck", effectivenessCheck)}
        />
      </Field>
      <Field className={narrativeWide()} label={t("details.fields.description")}>
        <Textarea
          value={description}
          rows={4}
          disabled={!canUpdate}
          placeholder={t("details.placeholders.description")}
          onChange={event => setDescription(event.currentTarget.value)}
          onBlur={() => saveText("description", description)}
        />
      </Field>
    </div>
  );
}
