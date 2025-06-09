import { Field, Option, Select } from "@probo/ui";
import { Suspense, useMemo, useState, type ComponentProps } from "react";
import { useTranslate } from "@probo/i18n";
import { type Control, Controller } from "react-hook-form";
import { usePaginatedMeasures } from "/hooks/graph/usePaginatedMeasures";

type Props = {
  organizationId: string;
  control: Control<any>;
  name: string;
  label?: string;
  error?: string;
  disabled?: boolean;
} & ComponentProps<typeof Field>;

export function MeasureSelectField({
  organizationId,
  control,
  disabled,
  ...props
}: Props) {
  return (
    <Field {...props}>
      <Suspense
        fallback={<Select variant="editor" disabled placeholder="Loading..." />}
      >
        <MeasureSelectWithQuery
          organizationId={organizationId}
          control={control}
          name={props.name}
          disabled={disabled}
        />
      </Suspense>
    </Field>
  );
}

function MeasureSelectWithQuery(
  props: Pick<Props, "organizationId" | "control" | "name" | "disabled">
) {
  const { __ } = useTranslate();
  const { name, organizationId, control, disabled } = props;
  const { data } = usePaginatedMeasures(organizationId);
  const [search, setSearch] = useState("");
  const measures = useMemo(() => {
    return (
      data?.measures.edges
        ?.filter(
          (edge) =>
            edge.node.name.toLowerCase().includes(search.toLowerCase()) ||
            edge.node.description?.toLowerCase().includes(search.toLowerCase())
        )
        .map((edge) => edge.node) ?? []
    );
  }, [data?.measures.edges, search]);

  return (
    <div>
      <Controller
        control={control}
        name={name}
        render={({ field }) => (
          <Select
            id={name}
            variant="editor"
            placeholder={__("Select an owner")}
            onValueChange={field.onChange}
            {...field}
            className="w-full"
            value={field.value ?? ""}
            onSearch={setSearch}
            searchValue={search}
            disabled={disabled}
          >
            {measures?.map((m) => (
              <Option key={m.id} value={m.id}>
                <div className="space-y-1 text-start min-w-0">
                  <div className="max-w-75 ellipsis overflow-hidden whitespace-pre-wrap">
                    {m.name}
                  </div>
                  <div className="text-sm text-txt-secondary">{m.category}</div>
                </div>
              </Option>
            ))}
          </Select>
        )}
      />
    </div>
  );
}
