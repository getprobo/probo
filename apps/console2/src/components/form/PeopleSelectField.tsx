import { Avatar, Field, Option, Select } from "@probo/ui";
import { Suspense } from "react";
import { useTranslate } from "@probo/i18n";
import { type Control, Controller } from "react-hook-form";
import { usePeople } from "/hooks/graph/PeopleGraph.ts";

type Props = {
  organizationId: string;
  control: Control<any>;
  name: string;
  label?: string;
  error?: string;
};

export function PeopleSelectField({
  organizationId,
  control,
  ...props
}: Props) {
  return (
    <Field {...props}>
      <Suspense
        fallback={<Select variant="editor" disabled placeholder="Loading..." />}
      >
        <PeopleSelectWithQuery
          organizationId={organizationId}
          control={control}
          name={props.name}
        />
      </Suspense>
    </Field>
  );
}

function PeopleSelectWithQuery(
  props: Pick<Props, "organizationId" | "control" | "name">
) {
  const { __ } = useTranslate();
  const { name, organizationId, control } = props;
  const people = usePeople(organizationId);

  return (
    <>
      <Controller
        control={control}
        name={name}
        render={({ field }) => (
          <Select
            id={name}
            variant="editor"
            placeholder={__("Select an owner")}
            onValueChange={field.onChange}
            key={people?.length.toString() ?? "0"}
            {...field}
            className="w-full"
            value={field.value ?? ""}
          >
            {people?.map((p) => (
              <Option key={p.id} value={p.id} className="flex gap-2">
                <Avatar name={p.fullName} />
                {p.fullName}
              </Option>
            ))}
          </Select>
        )}
      />
    </>
  );
}
