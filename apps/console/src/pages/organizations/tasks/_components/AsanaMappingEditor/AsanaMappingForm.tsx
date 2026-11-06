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

import { ArrowRightIcon, TrashIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Callout } from "@probo/ui/src/v2/Callout/Callout";
import { Dialog } from "@probo/ui/src/v2/Dialog/Dialog";
import { DialogClose } from "@probo/ui/src/v2/Dialog/DialogClose";
import { DialogDescription } from "@probo/ui/src/v2/Dialog/DialogDescription";
import { DialogFooter } from "@probo/ui/src/v2/Dialog/DialogFooter";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
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
}

interface PriorityRow {
  priority: (typeof PRIORITIES)[number];
  optionId: string;
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
  customFieldsAvailable: boolean;
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
  customFieldsAvailable,
  sections,
  fields,
  savedStates,
  savedFieldId,
  savedPriorities,
}: AsanaMappingFormProps) {
  const { t } = useTranslation("organizations/tasks");
  const { root, group, row, name, arrow, control, actions } = asanaMappingEditor();
  const [states, setStates] = useState<StateRow[]>(() => initialStates(sections, savedStates));
  const [fieldId, setFieldId] = useState(savedFieldId ?? NONE);
  const field = fields.find(item => item.id === fieldId) ?? null;
  const [priorities, setPriorities] = useState<PriorityRow[]>(() => initialPriorities(field, savedPriorities));
  const [hasMapping, setHasMapping] = useState(savedStates != null);
  const [confirmDelete, setConfirmDelete] = useState(false);
  const [pending, setPending] = useState<"save" | "delete" | null>(null);
  const [save] = useMutation<AsanaMappingFormSaveMutation>(saveMutation);

  function chooseField(next: string | null) {
    const id = next ?? NONE;
    setFieldId(id);
    const chosen = fields.find(item => item.id === id) ?? null;
    setPriorities(initialPriorities(chosen, id === savedFieldId ? savedPriorities : null));
  }

  async function submit() {
    setPending("save");
    try {
      await save({
        variables: {
          input: {
            organizationId,
            projectId,
            states: states.map(item => ({
              state: item.state,
              sectionId: item.sectionId === NONE ? "" : item.sectionId,
              inbound: false,
            })),
            priorityFieldId: !customFieldsAvailable || field == null ? null : field.id,
            priorities: !customFieldsAvailable || field == null
              ? []
              : priorities.map(item => ({
                  priority: item.priority,
                  optionId: item.optionId === NONE ? "" : item.optionId,
                  inbound: false,
                })),
          },
        },
      }, {
        successMessage: t("settingsDialog.asana.messages.mappingSaved"),
        errorToast: t("settingsDialog.asana.errors.mapping"),
      });
      setHasMapping(true);
    } catch {
      // The mapping toast already explains the failure.
    } finally {
      setPending(null);
    }
  }

  async function remove() {
    setPending("delete");
    try {
      await save({
        variables: {
          input: {
            organizationId,
            projectId,
            clear: true,
          },
        },
      }, {
        successMessage: t("settingsDialog.asana.messages.mappingDeleted"),
        errorToast: t("settingsDialog.asana.errors.deleteMapping"),
      });
      setHasMapping(false);
      setConfirmDelete(false);
      setStates(STATES.map(state => ({ state, sectionId: NONE })));
      setFieldId(NONE);
      setPriorities(PRIORITIES.map(priority => ({ priority, optionId: NONE })));
    } catch {
      // The mapping toast already explains the failure.
    } finally {
      setPending(null);
    }
  }

  return (
    <div className={root()}>
      <div className={group()}>
        {states.map(item => (
          <div key={item.state} className={row()}>
            <Text size={2} className={name()}>{stateLabel(t, item.state)}</Text>
            <ArrowRightIcon aria-hidden className={arrow()} />
            <div className={control()}>
              <Select
                value={item.sectionId}
                disabled={disabled || pending != null}
                onValueChange={sectionId => setStates(current => current.map(rowItem => (
                  rowItem.state === item.state
                    ? { ...rowItem, sectionId: sectionId ?? NONE }
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
            </div>
          </div>
        ))}
      </div>
      {customFieldsAvailable
        ? (
            <div className={group()}>
              <div className={row()}>
                <Text size={2} className={name()}>{t("settingsDialog.asana.priorityField")}</Text>
                <ArrowRightIcon aria-hidden className={arrow()} />
                <div className={control()}>
                  <Select value={fieldId} disabled={disabled || pending != null} onValueChange={chooseField}>
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
                </div>
              </div>
              {field == null
                ? null
                : priorities.map(item => (
                    <div key={item.priority} className={row()}>
                      <Text size={2} className={name()}>{priorityLabel(t, item.priority)}</Text>
                      <ArrowRightIcon aria-hidden className={arrow()} />
                      <div className={control()}>
                        <Select
                          value={item.optionId}
                          disabled={disabled || pending != null}
                          onValueChange={optionId => setPriorities(current => current.map(rowItem => (
                            rowItem.priority === item.priority
                              ? { ...rowItem, optionId: optionId ?? NONE }
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
                      </div>
                    </div>
                  ))}
            </div>
          )
        : (
            <Callout color="amber">
              {t("settingsDialog.asana.customFieldsUnavailable")}
            </Callout>
          )}
      <div className={actions()}>
        {hasMapping
          ? (
              <Button
                type="button"
                variant="surface"
                color="red"
                iconStart={<TrashIcon />}
                disabled={disabled || pending != null}
                onClick={() => setConfirmDelete(true)}
              >
                {t("settingsDialog.asana.deleteMapping")}
              </Button>
            )
          : null}
        <Button
          type="button"
          variant="solid"
          color="neutral"
          highContrast
          disabled={disabled || pending != null}
          loading={pending === "save"}
          onClick={() => void submit()}
        >
          {t("settingsDialog.asana.saveMapping")}
        </Button>
      </div>
      <Dialog open={confirmDelete} onOpenChange={setConfirmDelete}>
        <DialogPopup>
          <DialogHeader>
            <DialogTitle>{t("settingsDialog.asana.deleteMappingTitle")}</DialogTitle>
            <DialogDescription>
              {t("settingsDialog.asana.deleteMappingDescription")}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <DialogClose
              render={(
                <Button variant="soft" color="neutral" disabled={pending === "delete"}>
                  {t("detailsPage.actions.cancel")}
                </Button>
              )}
            />
            <Button
              type="button"
              variant="solid"
              color="red"
              iconStart={<TrashIcon />}
              loading={pending === "delete"}
              onClick={() => void remove()}
            >
              {t("detailsPage.delete.confirm")}
            </Button>
          </DialogFooter>
        </DialogPopup>
      </Dialog>
    </div>
  );
}

function initialStates(
  sections: { id: string; name: string }[],
  saved: { state: string; sectionId: string }[] | null,
): StateRow[] {
  return STATES.map((state) => {
    const existing = saved?.find(row => row.state === state);
    if (existing != null) {
      const known = existing.sectionId !== ""
        && sections.some(section => section.id === existing.sectionId);

      return {
        state,
        sectionId: known ? existing.sectionId : NONE,
      };
    }

    const match = sections.find(section => sameName(section.name, STATE_NAMES[state]));
    return { state, sectionId: match?.id ?? NONE };
  });
}

function initialPriorities(
  field: { options: { id: string; name: string }[] } | null,
  saved: { priority: string; optionId: string }[] | null,
): PriorityRow[] {
  return PRIORITIES.map((priority) => {
    const existing = saved?.find(row => row.priority === priority);
    if (existing != null && field != null) {
      const known = existing.optionId !== ""
        && field.options.some(option => option.id === existing.optionId);

      return {
        priority,
        optionId: known ? existing.optionId : NONE,
      };
    }

    const match = field?.options.find(option => sameName(option.name, PRIORITY_NAMES[priority]));
    return { priority, optionId: match?.id ?? NONE };
  });
}

function sameName(name: string, aliases: readonly string[]) {
  return aliases.some(alias => name.localeCompare(alias, undefined, { sensitivity: "accent" }) === 0);
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
