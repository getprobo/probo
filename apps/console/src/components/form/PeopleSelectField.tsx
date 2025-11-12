import { Avatar, Field, Option, Select } from "@probo/ui";
import { Suspense, type ComponentProps } from "react";
import { useTranslate } from "@probo/i18n";
import { type Control, Controller } from "react-hook-form";
import { usePeople } from "/hooks/graph/PeopleGraph.ts";

type Props = {
  organizationId: string;
  control: Control<any>;
  name: string;
  label?: string;
  error?: string;
  optional?: boolean;
} & ComponentProps<typeof Field>;

export function PeopleSelectField({
  organizationId,
  control,
  ...props
}: Props) {
  return (
    <Field {...props}>
      <Suspense
        fallback={<Select variant="editor" loading placeholder="Loading..." />}
      >
        <PeopleSelectWithQuery
          organizationId={organizationId}
          control={control}
          name={props.name}
          disabled={props.disabled}
          optional={props.optional}
        />
      </Suspense>
    </Field>
  );
}

function PeopleSelectWithQuery(
  props: Pick<
    Props,
    "organizationId" | "control" | "name" | "disabled" | "optional"
  >,
) {
  const { __ } = useTranslate();
  const { name, organizationId, control } = props;
  const people = usePeople(organizationId, { excludeContractEnded: true });

  return (
    <>
      <Controller
        control={control}
        name={name}
        render={({ field }) => (
          <Select
            disabled={props.disabled}
            id={name}
            variant="editor"
            placeholder={__("Select an owner")}
            onValueChange={(value) =>
              field.onChange(value === "__NONE__" ? null : value)
            }
            key={people?.length.toString() ?? "0"}
            {...field}
            className="w-full"
            value={field.value ?? (props.optional ? "__NONE__" : "")}
          >
            {props.optional && <Option value="__NONE__">{__("None")}</Option>}
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

type OptionsProps = {
  organizationId: string;
  optional?: boolean;
} & ComponentProps<typeof Field>;

export function PeopleSelectOptions({
  organizationId,
  ...props
}: OptionsProps) {
  return (
    <Suspense
      fallback={<Select variant="editor" loading placeholder="Loading..." />}
    >
      <PeopleSelectOptionsWithQuery
        organizationId={organizationId}
        optional={props.optional}
      />
    </Suspense>
  );
}

function PeopleSelectOptionsWithQuery(
  props: Pick<Props, "organizationId" | "disabled" | "optional">,
) {
  const { __ } = useTranslate();
  const { organizationId } = props;
  const people = usePeople(organizationId, { excludeContractEnded: true });

  return (
    <>
      {props.optional && <Option value="__NONE__">{__("None")}</Option>}
      {people?.map((p) => (
        <Option key={p.id} value={p.id} className="flex gap-2">
          <Avatar name={p.fullName} />
          {p.fullName}
        </Option>
      ))}
    </>
  );
}
