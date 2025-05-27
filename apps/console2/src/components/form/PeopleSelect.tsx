import { Avatar, Option, Select } from "@probo/ui";
import { Suspense } from "react";
import { useTranslate } from "@probo/i18n";
import { type Control, Controller } from "react-hook-form";
import { usePeople } from "/hooks/graph/PeopleGraph.ts";

export function PeopleSelect({
  organization,
  name,
  control,
}: {
  organization: string;
  name: string;
  control: Control<any>;
}) {
  return (
    <Suspense
      fallback={<Select variant="editor" disabled placeholder="Loading..." />}
    >
      <PeopleSelectWithQuery
        organization={organization}
        name={name}
        control={control}
      />
    </Suspense>
  );
}

function PeopleSelectWithQuery({
  organization,
  name,
  control,
}: {
  organization: string;
  name: string;
  control: Control;
}) {
  const { __ } = useTranslate();
  const people = usePeople(organization);

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
