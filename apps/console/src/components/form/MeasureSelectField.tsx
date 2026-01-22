import { useTranslate } from "@probo/i18n";
import { Combobox, ComboboxItem, Field } from "@probo/ui";
import { type ComponentProps, Suspense, useMemo, useState } from "react";
import { type Control, Controller, type FieldPath, type FieldValues } from "react-hook-form";

import { usePaginatedMeasures } from "#/hooks/graph/usePaginatedMeasures";

type Props<
  TFieldValues extends FieldValues = FieldValues,
  TName extends FieldPath<TFieldValues> = FieldPath<TFieldValues>,
> = {
  organizationId: string;
  control: Control<TFieldValues>;
  name: TName;
  label?: string;
  error?: string;
  disabled?: boolean;
  optional?: boolean;
} & ComponentProps<typeof Field>;

export function MeasureSelectField<TFieldValues extends FieldValues = FieldValues>({
  organizationId,
  control,
  disabled,
  optional,
  ...props
}: Props<TFieldValues>) {
  return (
    <Field {...props}>
      <Suspense
        fallback={<Combobox onSearch={() => {}} placeholder="Loading..." disabled><div /></Combobox>}
      >
        <MeasureSelectWithQuery<TFieldValues>
          organizationId={organizationId}
          control={control}
          name={props.name}
          disabled={disabled}
          optional={optional}
        />
      </Suspense>
    </Field>
  );
}

function MeasureSelectWithQuery<TFieldValues extends FieldValues = FieldValues>(
  props: Pick<Props<TFieldValues>, "organizationId" | "control" | "name" | "disabled" | "optional">,
) {
  const { __ } = useTranslate();
  const { name, organizationId, control, disabled, optional } = props;
  const { data } = usePaginatedMeasures(organizationId);
  const [search, setSearch] = useState("");
  const measures = useMemo(() => {
    return (
      data?.measures.edges
        ?.filter(
          edge =>
            edge.node.name.toLowerCase().includes(search.toLowerCase())
            || edge.node.description?.toLowerCase().includes(search.toLowerCase()),
        )
        .map(edge => edge.node) ?? []
    );
  }, [data?.measures.edges, search]);

  const allMeasures = useMemo(() => {
    return data?.measures.edges?.map(edge => edge.node) ?? [];
  }, [data?.measures.edges]);

  return (
    <div>
      <Controller
        control={control}
        name={name}
        render={({ field }) => {
          const selectedMeasure = field.value ? allMeasures?.find(m => m.id === field.value) : null;

          return (
            <Combobox
              id={name}
              placeholder={__("Select a measure")}
              value={selectedMeasure ? selectedMeasure.name : search}
              onSearch={setSearch}
              disabled={disabled}
            >
              {optional && (
                <ComboboxItem
                  onClick={() => {
                    field.onChange(null);
                    setSearch("");
                  }}
                >
                  {__("None")}
                </ComboboxItem>
              )}
              {measures?.map(m => (
                <ComboboxItem
                  key={m.id}
                  onClick={() => {
                    field.onChange(m.id);
                    setSearch(m.name);
                  }}
                >
                  <div className="space-y-1 text-start min-w-0">
                    <div className="max-w-75 ellipsis overflow-hidden whitespace-pre-wrap">
                      {m.name}
                    </div>
                    <div className="text-sm text-txt-secondary">{m.category}</div>
                  </div>
                </ComboboxItem>
              ))}
            </Combobox>
          );
        }}
      />
    </div>
  );
}
