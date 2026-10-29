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

import { Button } from "@probo/ui/src/v2/Button/Button";
import { Checkbox } from "@probo/ui/src/v2/Checkbox/Checkbox";
import { Field } from "@probo/ui/src/v2/form/Field";
import { Select } from "@probo/ui/src/v2/Select/Select";
import { SelectItem } from "@probo/ui/src/v2/Select/SelectItem";
import { SelectPopup } from "@probo/ui/src/v2/Select/SelectPopup";
import { SelectTrigger } from "@probo/ui/src/v2/Select/SelectTrigger";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql } from "react-relay";

import type { AsanaMappingFormSaveMutation } from "#/__generated__/core/AsanaMappingFormSaveMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { asanaMappingEditor } from "../../variants";

const NONE = "__none__";
const STATES = ["BACKLOG", "TODO", "IN_PROGRESS", "DONE", "CANCELED", "DUPLICATE"] as const;
const PRIORITIES = ["URGENT", "HIGH", "MEDIUM", "LOW"] as const;

const STATE_NAMES: Record<(typeof STATES)[number], readonly string[]> = {
  BACKLOG: ["Backlog"],
  TODO: ["To do", "À faire", "Te doen"],
  IN_PROGRESS: ["In progress", "En cours", "In uitvoering"],
  DONE: ["Done", "Terminé", "Voltooid"],
  CANCELED: ["Canceled", "Annulé", "Geannuleerd"],
  DUPLICATE: ["Duplicate", "Doublon", "Duplicaat"],
};

const PRIORITY_NAMES: Record<(typeof PRIORITIES)[number], readonly string[]> = {
  URGENT: ["Urgent", "Urgente"],
  HIGH: ["High", "Haute", "Hoog"],
  MEDIUM: ["Medium", "Moyenne", "Gemiddeld"],
  LOW: ["Low", "Basse", "Laag"],
};

const saveMutation = graphql`
  mutation AsanaMappingFormSaveMutation($input: UpdateAsanaProjectMappingInput!) {
    updateAsanaProjectMapping(input: $input) {
      setup {
        project {
          id
        }
      }
    }
  }
`;

interface StateRow {
  state: (typeof STATES)[number];
  sectionId: string;
  inbound: boolean;
}

interface PriorityRow {
  priority: (typeof PRIORITIES)[number];
  optionId: string;
  inbound: boolean;
}

export interface AsanaSectionOption {
  id: string;
  name: string;
}

export interface AsanaEnumFieldOption {
  id: string;
  name: string;
  options: { id: string; name: string }[];
}

export interface AsanaMappingFormProps {
  organizationId: string;
  projectId: string;
  disabled?: boolean;
  sections: AsanaSectionOption[];
  fields: AsanaEnumFieldOption[];
  savedStates: { state: string; sectionId: string; inbound: boolean }[] | null;
  savedFieldId: string | null;
  savedPriorities: { priority: string; optionId: string; inbound: boolean }[] | null;
}

export function AsanaMappingForm({
  organizationId,
  projectId,
  disabled,
  sections,
  fields,
  savedStates,
  savedFieldId,
  savedPriorities,
}: AsanaMappingFormProps) {
  const { t } = useTranslation("organizations/tasks");
  const { root, row, control, incoming } = asanaMappingEditor();
  const [states, setStates] = useState<StateRow[]>(() => initialStates(sections, savedStates));
  const [fieldId, setFieldId] = useState(savedFieldId ?? NONE);
  const field = fields.find(item => item.id === fieldId) ?? null;
  const [priorities, setPriorities] = useState<PriorityRow[]>(() => initialPriorities(field, savedPriorities));
  const [formError, setFormError] = useState<string | null>(null);
  const [save, isSaving] = useMutation<AsanaMappingFormSaveMutation>(saveMutation);

  function chooseField(next: string | null) {
    const id = next ?? NONE;
    setFieldId(id);
    const chosen = fields.find(item => item.id === id) ?? null;
    setPriorities(initialPriorities(chosen, null));
    setFormError(null);
  }

  async function submit() {
    const statesShare = sharedSelection(states.map(item => ({
      key: item.sectionId,
      inbound: item.inbound,
    })));
    const prioritiesShare = fieldId !== NONE && sharedSelection(priorities.map(item => ({
      key: item.optionId,
      inbound: item.inbound,
    })));
    if (statesShare || prioritiesShare) {
      setFormError(t("settingsDialog.asana.errors.shared"));
      return;
    }

    setFormError(null);

    try {
      await save({
        variables: {
          input: {
            organizationId,
            projectId,
            states: states.map(item => ({
              state: item.state,
              sectionId: item.sectionId === NONE ? "" : item.sectionId,
              inbound: item.sectionId !== NONE && item.inbound,
            })),
            priorityFieldId: fieldId === NONE ? null : fieldId,
            priorities: fieldId === NONE
              ? []
              : priorities.map(item => ({
                  priority: item.priority,
                  optionId: item.optionId === NONE ? "" : item.optionId,
                  inbound: item.optionId !== NONE && item.inbound,
                })),
          },
        },
      }, {
        successMessage: t("settingsDialog.asana.messages.mappingSaved"),
        errorToast: t("settingsDialog.asana.errors.mapping"),
      });
    } catch {
      // The mapping toast already explains the failure.
    }
  }

  return (
    <div className={root()}>
      <Text size={2} color="faint">{t("settingsDialog.asana.mappingDescription")}</Text>
      {states.map(item => (
        <div key={item.state} className={row()}>
          <div className={control()}>
            <Field label={stateLabel(t, item.state)}>
              <Select
                value={item.sectionId}
                disabled={disabled || isSaving}
                onValueChange={sectionId => setStates(current => current.map(rowItem => (
                  rowItem.state === item.state
                    ? {
                        ...rowItem,
                        sectionId: sectionId ?? NONE,
                        inbound: sectionId == null || sectionId === NONE ? false : rowItem.inbound,
                      }
                    : rowItem
                )))}
              >
                <SelectTrigger size={2} aria-label={stateLabel(t, item.state)}>
                  {(value: string | null) => sectionName(sections, value, t("settingsDialog.asana.noSection"))}
                </SelectTrigger>
                <SelectPopup>
                  <SelectItem value={NONE}>{t("settingsDialog.asana.noSection")}</SelectItem>
                  {sections.map(section => (
                    <SelectItem key={section.id} value={section.id}>{section.name}</SelectItem>
                  ))}
                </SelectPopup>
              </Select>
            </Field>
          </div>
          <label className={incoming()}>
            <Checkbox
              checked={item.inbound}
              disabled={disabled || isSaving || item.sectionId === NONE}
              onCheckedChange={checked => setStates(current => current.map(rowItem => (
                rowItem.state === item.state ? { ...rowItem, inbound: checked === true } : rowItem
              )))}
            />
            <Text size={2}>{t("settingsDialog.asana.incoming")}</Text>
          </label>
        </div>
      ))}
      <div className={control()}>
        <Field label={t("settingsDialog.asana.priorityField")}>
          <Select value={fieldId} disabled={disabled || isSaving} onValueChange={chooseField}>
            <SelectTrigger size={2} aria-label={t("settingsDialog.asana.priorityField")}>
              {(value: string | null) => fieldName(fields, value, t("settingsDialog.asana.noPriority"))}
            </SelectTrigger>
            <SelectPopup>
              <SelectItem value={NONE}>{t("settingsDialog.asana.noPriority")}</SelectItem>
              {fields.map(item => (
                <SelectItem key={item.id} value={item.id}>{item.name}</SelectItem>
              ))}
            </SelectPopup>
          </Select>
        </Field>
      </div>
      {field == null
        ? null
        : priorities.map(item => (
            <div key={item.priority} className={row()}>
              <div className={control()}>
                <Field label={priorityLabel(t, item.priority)}>
                  <Select
                    value={item.optionId}
                    disabled={disabled || isSaving}
                    onValueChange={optionId => setPriorities(current => current.map(rowItem => (
                      rowItem.priority === item.priority
                        ? {
                            ...rowItem,
                            optionId: optionId ?? NONE,
                            inbound: optionId == null || optionId === NONE ? false : rowItem.inbound,
                          }
                        : rowItem
                    )))}
                  >
                    <SelectTrigger size={2} aria-label={priorityLabel(t, item.priority)}>
                      {(value: string | null) => optionName(field.options, value, t("settingsDialog.asana.noOption"))}
                    </SelectTrigger>
                    <SelectPopup>
                      <SelectItem value={NONE}>{t("settingsDialog.asana.noOption")}</SelectItem>
                      {field.options.map(option => (
                        <SelectItem key={option.id} value={option.id}>{option.name}</SelectItem>
                      ))}
                    </SelectPopup>
                  </Select>
                </Field>
              </div>
              <label className={incoming()}>
                <Checkbox
                  checked={item.inbound}
                  disabled={disabled || isSaving || item.optionId === NONE}
                  onCheckedChange={checked => setPriorities(current => current.map(rowItem => (
                    rowItem.priority === item.priority ? { ...rowItem, inbound: checked === true } : rowItem
                  )))}
                />
                <Text size={2}>{t("settingsDialog.asana.incoming")}</Text>
              </label>
            </div>
          ))}
      {formError == null ? null : <Text size={2} color="red">{formError}</Text>}
      <Button
        size={1}
        variant="solid"
        color="neutral"
        highContrast
        disabled={disabled}
        loading={isSaving}
        onClick={() => void submit()}
      >
        {t("settingsDialog.asana.saveMapping")}
      </Button>
    </div>
  );
}

function initialStates(
  sections: { id: string; name: string }[],
  saved: { state: string; sectionId: string; inbound: boolean }[] | null,
): StateRow[] {
  return STATES.map((state) => {
    const existing = saved?.find(row => row.state === state);
    if (existing != null) {
      return {
        state,
        sectionId: existing.sectionId === "" ? NONE : existing.sectionId,
        inbound: existing.inbound,
      };
    }

    const match = sections.find(section => sameName(section.name, STATE_NAMES[state]));
    return { state, sectionId: match?.id ?? NONE, inbound: false };
  });
}

function initialPriorities(
  field: { options: { id: string; name: string }[] } | null,
  saved: { priority: string; optionId: string; inbound: boolean }[] | null,
): PriorityRow[] {
  return PRIORITIES.map((priority) => {
    const existing = saved?.find(row => row.priority === priority);
    if (existing != null && field != null) {
      return {
        priority,
        optionId: existing.optionId === "" ? NONE : existing.optionId,
        inbound: existing.inbound,
      };
    }

    const match = field?.options.find(option => sameName(option.name, PRIORITY_NAMES[priority]));
    return { priority, optionId: match?.id ?? NONE, inbound: false };
  });
}

function sameName(name: string, aliases: readonly string[]) {
  return aliases.some(alias => name.localeCompare(alias, undefined, { sensitivity: "accent" }) === 0);
}

function sharedSelection(rows: { key: string; inbound: boolean }[]) {
  const groups = new Map<string, { count: number; inbound: number }>();
  for (const row of rows) {
    if (row.key === NONE || row.key === "") {
      continue;
    }

    const group = groups.get(row.key) ?? { count: 0, inbound: 0 };
    group.count += 1;
    if (row.inbound) {
      group.inbound += 1;
    }
    groups.set(row.key, group);
  }

  for (const group of groups.values()) {
    if (group.count > 1 && group.inbound !== 1) {
      return true;
    }
  }

  return false;
}

function stateLabel(t: (key: string) => string, state: string) {
  switch (state) {
    case "BACKLOG":
      return t("detailsPage.states.backlog");
    case "TODO":
      return t("detailsPage.states.todo");
    case "IN_PROGRESS":
      return t("detailsPage.states.inProgress");
    case "DONE":
      return t("detailsPage.states.done");
    case "CANCELED":
      return t("detailsPage.states.canceled");
    default:
      return t("detailsPage.states.duplicate");
  }
}

function priorityLabel(t: (key: string) => string, priority: string) {
  switch (priority) {
    case "URGENT":
      return t("detailsPage.priorities.urgent");
    case "HIGH":
      return t("detailsPage.priorities.high");
    case "LOW":
      return t("detailsPage.priorities.low");
    default:
      return t("detailsPage.priorities.medium");
  }
}

function sectionName(sections: { id: string; name: string }[], id: string | null, empty: string) {
  if (id == null || id === "" || id === NONE) {
    return empty;
  }

  return sections.find(section => section.id === id)?.name ?? empty;
}

function fieldName(fields: { id: string; name: string }[], id: string | null, empty: string) {
  if (id == null || id === "" || id === NONE) {
    return empty;
  }

  return fields.find(field => field.id === id)?.name ?? empty;
}

function optionName(options: { id: string; name: string }[], id: string | null, empty: string) {
  if (id == null || id === "" || id === NONE) {
    return empty;
  }

  return options.find(option => option.id === id)?.name ?? empty;
}
