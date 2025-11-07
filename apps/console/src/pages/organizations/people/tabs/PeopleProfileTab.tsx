import z from "zod";
import { peopleRoles, formatDatetime } from "@probo/helpers";
import { useOutletContext } from "react-router";
import type { PeopleGraphNodeQuery$data } from "/hooks/graph/__generated__/PeopleGraphNodeQuery.graphql";
import { useTranslate } from "@probo/i18n";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import type { PeopleGraphUpdateMutation } from "/hooks/graph/__generated__/PeopleGraphUpdateMutation.graphql";
import { updatePeopleMutation } from "/hooks/graph/PeopleGraph";
import { Button, Card, Field, Input } from "@probo/ui";
import { EmailsField } from "/components/form/EmailsField";
import { IfAuthorized } from "/permissions/IfAuthorized";

const schema = z.object({
  fullName: z.string().min(1),
  primaryEmailAddress: z.string().email(),
  position: z.string().min(1).nullable(),
  additionalEmailAddresses: z.preprocess(
    // Empty additional emails are skipped
    (v) => (v as string[]).filter((v) => !!v),
    z.array(z.string().email())
  ),
  kind: z.enum(peopleRoles),
  contractStartDate: z.string().optional().nullable(),
  contractEndDate: z.string().optional().nullable(),
});

export default function PeopleProfileTab() {
  const { people } = useOutletContext<{
    people: PeopleGraphNodeQuery$data["node"];
  }>();
  const { __ } = useTranslate();
  const { control, formState, handleSubmit, register, reset } =
    useFormWithSchema(schema, {
      defaultValues: {
        kind: people.kind,
        fullName: people.fullName,
        primaryEmailAddress: people.primaryEmailAddress,
        position: people.position,
        additionalEmailAddresses: [...(people.additionalEmailAddresses ?? [])],
        contractStartDate: people.contractStartDate?.split('T')[0] || "",
        contractEndDate: people.contractEndDate?.split('T')[0] || "",
      },
    });
  const [mutate, isMutating] = useMutationWithToasts<PeopleGraphUpdateMutation>(
    updatePeopleMutation,
    {
      successMessage: __("Member updated successfully."),
      errorMessage: __("Failed to update member"),
    }
  );

  const onSubmit = handleSubmit((data) => {
    const input = {
      id: people.id!,
      fullName: data.fullName,
      primaryEmailAddress: data.primaryEmailAddress,
      position: data.position,
      additionalEmailAddresses: data.additionalEmailAddresses,
      kind: people.kind,
      contractStartDate: formatDatetime(data.contractStartDate) ?? null,
      contractEndDate: formatDatetime(data.contractEndDate) ?? null,
    };

    mutate({
      variables: { input },
      onCompleted: () => {
        reset(data);
      },
    });
  });

  return (
    <form onSubmit={onSubmit} className="space-y-4">
      <Card padded className="space-y-4">
        <Field label={__("Full name")} {...register("fullName")} type="text" />
        <Field
          label={__("Position")}
          {...register("position")}
          type="text"
          placeholder={__("e.g. CEO, CFO, etc.")}
        />
        <Field
          label={__("Primary email")}
          {...register("primaryEmailAddress")}
          type="email"
        />
        <EmailsField control={control} register={register} />
        <Field label={__("Contract start date")}>
          <Input {...register("contractStartDate")} type="date" />
        </Field>
        <Field label={__("Contract end date")}>
          <Input {...register("contractEndDate")} type="date" />
        </Field>
      </Card>
      <div className="flex justify-end">
        {formState.isDirty && (
          <IfAuthorized entity="People" action="update">
            <Button type="submit" disabled={isMutating}>
              {__("Update")}
            </Button>
          </IfAuthorized>
        )}
      </div>
    </form>
  );
}
