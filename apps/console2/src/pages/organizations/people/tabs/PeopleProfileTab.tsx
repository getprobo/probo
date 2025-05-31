import z from "zod";
import { peopleRoles } from "@probo/helpers";
import { useOutletContext } from "react-router";
import type { PeopleGraphNodeQuery$data } from "/hooks/graph/__generated__/PeopleGraphNodeQuery.graphql";
import { useTranslate } from "@probo/i18n";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import type { PeopleGraphUpdateMutation } from "/hooks/graph/__generated__/PeopleGraphUpdateMutation.graphql";
import { updatePeopleMutation } from "/hooks/graph/PeopleGraph";
import { Button, Card, Field } from "@probo/ui";
import { EmailsField } from "/components/form/EmailsField";

const schema = z.object({
  fullName: z.string().min(1),
  primaryEmailAddress: z.string().email(),
  additionalEmailAddresses: z.preprocess(
    // Empty additional emails are skipped
    (v) => (v as string[]).filter((v) => !!v),
    z.array(z.string().email())
  ),
  kind: z.enum(peopleRoles),
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
        additionalEmailAddresses: [...(people.additionalEmailAddresses ?? [])],
      },
    });
  const [mutate, isMutating] = useMutationWithToasts<PeopleGraphUpdateMutation>(
    updatePeopleMutation,
    {
      successMessage: __("Member updated successfully."),
      errorMessage: __("Failed to update member. Please try again."),
    }
  );

  const onSubmit = handleSubmit((data) => {
    mutate({
      variables: {
        input: {
          id: people.id!,
          fullName: data.fullName,
          primaryEmailAddress: data.primaryEmailAddress,
          additionalEmailAddresses: data.additionalEmailAddresses,
          // TODO : make these field optional in the query (server side)
          kind: people.kind,
        },
      },
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
          label={__("Primary email")}
          {...register("primaryEmailAddress")}
          type="email"
        />
        <EmailsField control={control} register={register} />
      </Card>
      <div className="flex justify-end">
        {formState.isDirty && (
          <Button type="submit" disabled={isMutating}>
            {__("Update")}
          </Button>
        )}
      </div>
    </form>
  );
}
