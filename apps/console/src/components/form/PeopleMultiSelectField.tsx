import { useTranslate } from "@probo/i18n";
import { Avatar, Badge, Button, Field, IconCrossLargeX, Option, Select } from "@probo/ui";
import { type ComponentProps, Suspense, useState } from "react";
import { type Control, Controller, type FieldValues, type Path } from "react-hook-form";

import { usePeople } from "#/hooks/graph/PeopleGraph";

type Person = {
  id: string;
  fullName: string;
  emailAddress?: string | null;
};

type Props<T extends FieldValues = FieldValues> = {
  organizationId: string;
  control: Control<T>;
  name: string;
  label?: string;
  error?: string;
  selectedPeople?: Person[];
  placeholder?: string;
} & ComponentProps<typeof Field>;

export function PeopleMultiSelectField<T extends FieldValues = FieldValues>({
  organizationId,
  control,
  selectedPeople = [],
  placeholder,
  ...props
}: Props<T>) {
  return (
    <Field {...props}>
      <Suspense
        fallback={<Select variant="editor" disabled placeholder="Loading..." />}
      >
        <PeopleMultiSelectWithQuery
          organizationId={organizationId}
          control={control}
          name={props.name}
          disabled={props.disabled}
          selectedPeople={selectedPeople}
          placeholder={placeholder}
        />
      </Suspense>
    </Field>
  );
}

function PeopleMultiSelectWithQuery<T extends FieldValues = FieldValues>(
  props: Pick<Props<T>, "organizationId" | "control" | "name" | "disabled" | "selectedPeople" | "placeholder">,
) {
  const { __ } = useTranslate();
  const { name, organizationId, control, selectedPeople = [], placeholder } = props;
  const people = usePeople(organizationId, { excludeContractEnded: true });
  const [isOpen, setIsOpen] = useState(false);

  const allPeople = [...people];
  selectedPeople.forEach((selectedPerson) => {
    if (!allPeople.find(p => p.id === selectedPerson.id)) {
      allPeople.push({
        id: selectedPerson.id,
        fullName: selectedPerson.fullName,
        emailAddress: selectedPerson.emailAddress ?? "",
      });
    }
  });

  return (
    <>
      <Controller
        control={control}
        name={name as Path<T>}
        render={({ field }) => {
          const selectedPeopleIds = (Array.isArray(field.value) ? field.value : []) as string[];

          const selectedPeople = allPeople.filter(p => selectedPeopleIds.includes(p.id));
          const availablePeople = allPeople.filter(p => !selectedPeopleIds.includes(p.id));

          const handleAddPerson = (personId: string) => {
            const newValue = [...selectedPeopleIds, personId];
            field.onChange(newValue);
            setIsOpen(false);
          };

          const handleRemovePerson = (personId: string) => {
            const newValue = selectedPeopleIds.filter((id: string) => id !== personId);
            field.onChange(newValue);
          };

          return (
            <div className="space-y-2">
              {availablePeople.length > 0 && !props.disabled && (
                <Select
                  disabled={props.disabled}
                  id={name}
                  variant="editor"
                  placeholder={placeholder ?? __("Add people...")}
                  onValueChange={handleAddPerson}
                  key={`${selectedPeopleIds.length}-${people.length}`}
                  className="w-full"
                  value=""
                  open={isOpen}
                  onOpenChange={setIsOpen}
                >
                  {availablePeople.map(person => (
                    <Option key={person.id} value={person.id} className="flex gap-2">
                      <Avatar
                        name={person.fullName}
                        size="s"
                      />
                      <div className="flex flex-col">
                        <span>{person.fullName}</span>
                        {person.emailAddress && (
                          <span className="text-xs text-txt-secondary">
                            {person.emailAddress}
                          </span>
                        )}
                      </div>
                    </Option>
                  ))}
                </Select>
              )}

              {selectedPeople.length > 0 && (
                <div className="flex flex-wrap gap-2">
                  {selectedPeople.map(person => (
                    <Badge key={person.id} variant="neutral" className="flex items-center gap-2">
                      <Avatar
                        name={person.fullName}
                        size="s"
                      />
                      <span>{person.fullName}</span>
                      {!props.disabled && (
                        <Button
                          type="button"
                          variant="tertiary"
                          icon={IconCrossLargeX}
                          onClick={() => handleRemovePerson(person.id)}
                          className="h-4 w-4 p-0 hover:bg-transparent"
                        />
                      )}
                    </Badge>
                  ))}
                </div>
              )}

              {selectedPeople.length === 0 && availablePeople.length === 0 && (
                <div className="text-sm text-txt-secondary py-2">
                  {__("No people available")}
                </div>
              )}
            </div>
          );
        }}
      />
    </>
  );
}
