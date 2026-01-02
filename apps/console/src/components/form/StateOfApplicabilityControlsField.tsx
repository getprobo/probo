import { useTranslate } from "@probo/i18n";
import {
  Field,
  Select,
  Option,
  Checkbox,
  Textarea,
  Spinner,
  Button,
  IconChevronDown,
  IconChevronUp,
  IconTrashCan,
  IconPlusLarge,
} from "@probo/ui";
import { Suspense, useState, useMemo, useEffect } from "react";
import { Controller, type Control, type UseFormSetValue, type FieldValues, type Path, type PathValue } from "react-hook-form";
import { useLazyLoadQuery } from "react-relay";
import { graphql } from "relay-runtime";
import { useOrganizationId } from "/hooks/useOrganizationId";
import type { StateOfApplicabilityControlsFieldFrameworksQuery } from "./__generated__/StateOfApplicabilityControlsFieldFrameworksQuery.graphql";
import type { StateOfApplicabilityControlsFieldFrameworkControlsQuery } from "./__generated__/StateOfApplicabilityControlsFieldFrameworkControlsQuery.graphql";

const frameworksQuery = graphql`
  query StateOfApplicabilityControlsFieldFrameworksQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        frameworks(first: 100) {
          edges {
            node {
              id
              name
            }
          }
        }
      }
    }
  }
`;

const frameworkControlsQuery = graphql`
  query StateOfApplicabilityControlsFieldFrameworkControlsQuery(
    $frameworkId: ID!
  ) {
    framework: node(id: $frameworkId) {
      ... on Framework {
        id
        controls(first: 500, orderBy: { field: SECTION_TITLE, direction: ASC }) {
          edges {
            node {
              id
              sectionTitle
              name
            }
          }
        }
      }
    }
  }
`;

type ControlSelection = {
  controlId: string;
  state: "EXCLUDED" | "IMPLEMENTED" | "NOT_IMPLEMENTED";
  exclusionJustification?: string;
};

type FrameworkData = {
  id: string;
  name: string;
  controls: Array<{
    id: string;
    sectionTitle: string;
    name: string;
  }>;
};

type Props<T extends FieldValues = FieldValues> = {
  control: Control<T>;
  setValue: UseFormSetValue<T>;
  name: string;
  initialControls?: ControlSelection[];
  initialFrameworkIds?: Set<string>;
};

export function StateOfApplicabilityControlsField<T extends FieldValues = FieldValues>({ control, setValue, name, initialControls, initialFrameworkIds }: Props<T>) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();

  // Initialize form value with initialControls
  useEffect(() => {
    if (initialControls && initialControls.length > 0) {
      setValue(name as Path<T>, initialControls as PathValue<T, Path<T>>);
    }
  }, [initialControls, setValue, name]);

  // Initialize framework selection from initialFrameworkIds
  const [selectedFrameworkIds, setSelectedFrameworkIds] = useState<Set<string>>(initialFrameworkIds || new Set());
  const [expandedFrameworks, setExpandedFrameworks] = useState<Set<string>>(initialFrameworkIds || new Set());
  const [frameworkDataMap, setFrameworkDataMap] = useState<Map<string, FrameworkData>>(new Map());
  const [newFrameworkId, setNewFrameworkId] = useState<string>("");

  // Update framework selection when initialFrameworkIds changes
  useEffect(() => {
    if (initialFrameworkIds) {
      setSelectedFrameworkIds(initialFrameworkIds);
      setExpandedFrameworks(initialFrameworkIds);
    }
  }, [initialFrameworkIds]);

  const addFramework = (frameworkId: string) => {
    if (!frameworkId || selectedFrameworkIds.has(frameworkId)) return;
    setSelectedFrameworkIds(new Set([...selectedFrameworkIds, frameworkId]));
    setExpandedFrameworks(new Set([...expandedFrameworks, frameworkId]));
    setNewFrameworkId("");
  };

  const removeFramework = (frameworkId: string) => {
    const newSet = new Set(selectedFrameworkIds);
    newSet.delete(frameworkId);
    setSelectedFrameworkIds(newSet);

    const newExpanded = new Set(expandedFrameworks);
    newExpanded.delete(frameworkId);
    setExpandedFrameworks(newExpanded);

    const newMap = new Map(frameworkDataMap);
    newMap.delete(frameworkId);
    setFrameworkDataMap(newMap);
  };

  const toggleFramework = (frameworkId: string) => {
    const newExpanded = new Set(expandedFrameworks);
    if (newExpanded.has(frameworkId)) {
      newExpanded.delete(frameworkId);
    } else {
      newExpanded.add(frameworkId);
    }
    setExpandedFrameworks(newExpanded);
  };

  return (
    <div className="space-y-4">
      <div>
        <h4 className="font-medium text-txt-primary mb-4">
          {__("Select Controls")}
        </h4>

        <div className="space-y-3">
          <Field label={__("Add Framework")}>
            <Suspense fallback={<Select variant="editor" disabled placeholder={__("Loading...")} />}>
              <FrameworkSelect
                organizationId={organizationId}
                selectedFrameworkIds={selectedFrameworkIds}
                value={newFrameworkId}
                onValueChange={setNewFrameworkId}
                onAdd={addFramework}
              />
            </Suspense>
          </Field>

          {Array.from(selectedFrameworkIds).map((frameworkId) => (
            <Suspense
              key={frameworkId}
              fallback={
                <div className="flex items-center justify-center py-8">
                  <Spinner />
                </div>
              }
            >
              <FrameworkSection
                frameworkId={frameworkId}
                isExpanded={expandedFrameworks.has(frameworkId)}
                onToggle={() => toggleFramework(frameworkId)}
                onRemove={() => removeFramework(frameworkId)}
                control={control}
                name={name}
                frameworkDataMap={frameworkDataMap}
                setFrameworkDataMap={setFrameworkDataMap}
              />
            </Suspense>
          ))}
        </div>
      </div>
    </div>
  );
}

function FrameworkSelect({
  organizationId,
  selectedFrameworkIds,
  value,
  onValueChange,
  onAdd,
}: {
  organizationId: string;
  selectedFrameworkIds: Set<string>;
  value: string;
  onValueChange: (value: string) => void;
  onAdd: (frameworkId: string) => void;
}) {
  const { __ } = useTranslate();
  const data = useLazyLoadQuery<StateOfApplicabilityControlsFieldFrameworksQuery>(
    frameworksQuery,
    { organizationId },
    { fetchPolicy: "network-only" }
  );
  const frameworks: Array<{ id: string; name: string }> =
    (data?.organization && "frameworks" in data.organization && data.organization.frameworks?.edges
      ?.map((edge) => edge.node)
      .filter((node): node is NonNullable<typeof node> => node !== null)) || [];

  const availableFrameworks = frameworks.filter(
    (framework: { id: string; name: string }) => !selectedFrameworkIds.has(framework.id)
  );

  return (
    <div className="flex gap-2">
      <Select
        variant="editor"
        placeholder={__("Select a framework")}
        onValueChange={onValueChange}
        value={value}
        className="flex-1"
      >
        {availableFrameworks.map((framework: { id: string; name: string }) => (
          <Option key={framework.id} value={framework.id}>
            {framework.name}
          </Option>
        ))}
      </Select>
      <Button
        type="button"
        variant="secondary"
        icon={IconPlusLarge}
        onClick={() => onAdd(value)}
        disabled={!value || selectedFrameworkIds.has(value)}
      >
        {__("Add")}
      </Button>
    </div>
  );
}

function FrameworkSection<T extends FieldValues = FieldValues>({
  frameworkId,
  isExpanded,
  onToggle,
  onRemove,
  control,
  name,
  frameworkDataMap,
  setFrameworkDataMap,
}: {
  frameworkId: string;
  isExpanded: boolean;
  onToggle: () => void;
  onRemove: () => void;
  control: Control<T>;
  name: string;
  frameworkDataMap: Map<string, FrameworkData>;
  setFrameworkDataMap: React.Dispatch<React.SetStateAction<Map<string, FrameworkData>>>;
}) {
  const { __ } = useTranslate();
  const data = useLazyLoadQuery<StateOfApplicabilityControlsFieldFrameworkControlsQuery>(
    frameworkControlsQuery,
    { frameworkId },
    { fetchPolicy: "network-only" }
  );

  const framework = data?.framework && "controls" in data.framework ? data.framework : null;
  const frameworkName: string = framework && "name" in framework && typeof framework.name === "string" ? framework.name : "";
  const controls: Array<{ id: string; sectionTitle: string; name: string }> = useMemo(
    () =>
      (framework?.controls?.edges
        ?.map((edge) => edge.node)
        .filter((node): node is NonNullable<typeof node> => node !== null)) ?? [],
    [framework]
  );

  useEffect(() => {
    if (framework && !frameworkDataMap.has(frameworkId)) {
      setFrameworkDataMap((prev) => {
        const newMap = new Map(prev);
        newMap.set(frameworkId, {
          id: frameworkId,
          name: frameworkName,
          controls,
        });
        return newMap;
      });
    }
  }, [framework, frameworkId, frameworkName, controls, frameworkDataMap, setFrameworkDataMap]);

  const cachedData = frameworkDataMap.get(frameworkId);
  const displayName: string = cachedData?.name || frameworkName;
  const displayControls: Array<{ id: string; sectionTitle: string; name: string }> = cachedData?.controls || controls;

  return (
    <div className="border border-border-low rounded-lg">
      <div className="flex items-center justify-between p-4">
        <button
          type="button"
          onClick={onToggle}
          className="flex items-center gap-2 flex-1 text-left hover:bg-subtle -m-4 p-4 rounded-lg"
        >
          {isExpanded ? (
            <IconChevronUp size={16} className="text-txt-tertiary" />
          ) : (
            <IconChevronDown size={16} className="text-txt-tertiary" />
          )}
          <span className="font-medium text-txt-primary">{displayName}</span>
          <span className="text-sm text-txt-tertiary">
            ({displayControls.length} {__("controls")})
          </span>
        </button>
        <Button
          type="button"
          variant="secondary"
          icon={IconTrashCan}
          onClick={onRemove}
          className="ml-2"
        >
          {__("Remove")}
        </Button>
      </div>

      {isExpanded && (
        <div className="border-t border-border-low">
          <Controller
            control={control}
            name={name as Path<T>}
            render={({ field }) => {
              const selectedControls: ControlSelection[] = (Array.isArray(field.value) ? field.value : []) as ControlSelection[];
              const selectedControlIds = new Set(
                selectedControls.map((c: ControlSelection) => c.controlId)
              );

              const toggleControl = (controlId: string) => {
                const isSelected = selectedControlIds.has(controlId);
                if (isSelected) {
                  field.onChange(
                    selectedControls.filter((c: ControlSelection) => c.controlId !== controlId)
                  );
                } else {
                  field.onChange([
                    ...selectedControls,
                    {
                      controlId,
                      state: "IMPLEMENTED" as const,
                      exclusionJustification: undefined,
                    },
                  ]);
                }
              };

              const updateControlState = (
                controlId: string,
                state: "EXCLUDED" | "IMPLEMENTED" | "NOT_IMPLEMENTED"
              ) => {
                field.onChange(
                  selectedControls.map((c: ControlSelection) =>
                    c.controlId === controlId ? { ...c, state } : c
                  )
                );
              };

              const updateJustification = (
                controlId: string,
                exclusionJustification: string
              ) => {
                field.onChange(
                  selectedControls.map((c: ControlSelection) =>
                    c.controlId === controlId
                      ? { ...c, exclusionJustification }
                      : c
                  )
                );
              };

              const getControlState = (controlId: string) => {
                const selected = selectedControls.find((c: ControlSelection) => c.controlId === controlId);
                return selected?.state || "IMPLEMENTED";
              };

              const getExclusionJustification = (controlId: string) => {
                const selected = selectedControls.find((c: ControlSelection) => c.controlId === controlId);
                return selected?.exclusionJustification || "";
              };

              const selectAll = () => {
                const newSelectedControls: ControlSelection[] = [...selectedControls];
                displayControls.forEach((ctrl) => {
                  if (!selectedControlIds.has(ctrl.id)) {
                    newSelectedControls.push({
                      controlId: ctrl.id,
                      state: "IMPLEMENTED" as const,
                      exclusionJustification: undefined,
                    });
                  }
                });
                field.onChange(newSelectedControls);
              };

              const deselectAll = () => {
                const controlIdsToRemove = new Set(displayControls.map((ctrl) => ctrl.id));
                const newSelectedControls = selectedControls.filter(
                  (c: ControlSelection) => !controlIdsToRemove.has(c.controlId)
                );
                field.onChange(newSelectedControls);
              };

              const allSelected = displayControls.length > 0 && displayControls.every((ctrl) => selectedControlIds.has(ctrl.id));

              return (
                <div className="p-4 space-y-4">
                  {displayControls.length > 0 && (
                    <div className="flex justify-end">
                      <Button
                        type="button"
                        variant="quaternary"
                        onClick={allSelected ? deselectAll : selectAll}
                        className="text-xs h-7 min-h-7"
                      >
                        {allSelected ? __("Deselect All") : __("Select All")}
                      </Button>
                    </div>
                  )}
                  <div className="border border-border-low rounded-lg max-h-96 overflow-y-auto">
                    {displayControls.length === 0 ? (
                      <div className="p-4 text-center text-txt-tertiary">
                        {__("No controls found in this framework")}
                      </div>
                    ) : (
                      <div className="divide-y divide-border-low">
                        {displayControls.map((ctrl) => {
                          const isSelected = selectedControlIds.has(ctrl.id);
                          const state = getControlState(ctrl.id);
                          const exclusionJustification = getExclusionJustification(
                            ctrl.id
                          );

                          return (
                            <div key={ctrl.id} className="p-4 space-y-3">
                              <div className="flex items-start gap-3">
                                <Checkbox
                                  checked={isSelected}
                                  onChange={() => toggleControl(ctrl.id)}
                                />
                                <div className="flex-1 min-w-0">
                                  <div className="font-medium">
                                    {ctrl.sectionTitle}: {ctrl.name}
                                  </div>
                                </div>
                              </div>

                              {isSelected && (
                                <div className="ml-7 space-y-2">
                                  <Field label={__("State")}>
                                    <Select
                                      variant="editor"
                                      value={state}
                                      onValueChange={(value) =>
                                        updateControlState(
                                          ctrl.id,
                                          value as
                                            | "EXCLUDED"
                                            | "IMPLEMENTED"
                                            | "NOT_IMPLEMENTED"
                                        )
                                      }
                                      className="w-full"
                                    >
                                      <Option value="IMPLEMENTED">
                                        {__("Implemented")}
                                      </Option>
                                      <Option value="NOT_IMPLEMENTED">
                                        {__("Not Implemented")}
                                      </Option>
                                      <Option value="EXCLUDED">
                                        {__("Excluded")}
                                      </Option>
                                    </Select>
                                  </Field>

                                  {state === "EXCLUDED" || state === "NOT_IMPLEMENTED" && (
                                    <Field label={__("Justification")}>
                                      <Textarea
                                        value={exclusionJustification}
                                        onChange={(e) =>
                                          updateJustification(
                                            ctrl.id,
                                            e.target.value
                                          )
                                        }
                                        placeholder={__(
                                          "Reason for exclusion"
                                        )}
                                        autogrow
                                      />
                                    </Field>
                                  )}
                                </div>
                              )}
                            </div>
                          );
                        })}
                      </div>
                    )}
                  </div>
                </div>
              );
            }}
          />
        </div>
      )}
    </div>
  );
}
