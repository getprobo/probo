import { formatDatetime, getRoles, peopleRoles } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Button, Card, Field, Input, Option } from "@probo/ui";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { UserFormFragment$key } from "#/__generated__/iam/UserFormFragment.graphql";
import type { UserFormMutation } from "#/__generated__/iam/UserFormMutation.graphql";
import { ControlledField } from "#/components/form/ControlledField";
import { EmailsField } from "#/components/form/EmailsField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

const fragment = graphql`
  fragment UserFormFragment on MembershipProfile {
    id
    fullName
    kind
    position
    additionalEmailAddresses
    contractStartDate
    contractEndDate
    canUpdate: permission(action: "iam:membership-profile:update")
  }
`;

const updateUserMutation = graphql`
  mutation UserFormMutation($input: UpdateProfileInput!) {
    updateProfile(input: $input) {
      profile {
        id
      }
    }
  }
`;

const schema = z.object({
  fullName: z.string().min(1),
  position: z.string().min(1).nullable(),
  additionalEmailAddresses: z.preprocess(
    // Empty additional emails are skipped
    v => (v as string[]).filter(v => !!v),
    z.array(z.string().email()),
  ),
  kind: z.enum(peopleRoles),
  contractStartDate: z.string().optional().nullable(),
  contractEndDate: z.string().optional().nullable(),
});

export function UserForm(props: { fragmentRef: UserFormFragment$key }) {
  const { fragmentRef } = props;

  const { __ } = useTranslate();

  const user = useFragment<UserFormFragment$key>(fragment, fragmentRef);

  const { control, formState, handleSubmit, register, reset }
    = useFormWithSchema(schema, {
      defaultValues: {
        kind: user.kind,
        fullName: user.fullName,
        position: user.position,
        additionalEmailAddresses: [...user.additionalEmailAddresses],
        contractStartDate: user.contractStartDate?.split("T")[0] || "",
        contractEndDate: user.contractEndDate?.split("T")[0] || "",
      },
    });
  const [mutate, isMutating] = useMutationWithToasts<UserFormMutation>(
    updateUserMutation,
    {
      successMessage: __("Member updated successfully."),
      errorMessage: __("Failed to update member"),
    },
  );
  const onSubmit = handleSubmit(async (data: z.infer<typeof schema>) => {
    const input = {
      id: user.id,
      fullName: data.fullName,
      additionalEmailAddresses: data.additionalEmailAddresses,
      kind: data.kind,
      position: data.position,
      contractStartDate: formatDatetime(data.contractStartDate) ?? null,
      contractEndDate: formatDatetime(data.contractEndDate) ?? null,
    };

    await mutate({
      variables: { input },
      onCompleted: () => {
        reset(data);
      },
    });
  });

  return (
    <form onSubmit={e => void onSubmit(e)} className="space-y-4">
      <Card padded className="space-y-4">
        <Field label={__("Full name")} {...register("fullName")} type="text" />
        <ControlledField
          control={control}
          name="kind"
          type="select"
          label={__("Type")}
          disabled={!user.canUpdate}
        >
          {getRoles(__).map(role => (
            <Option key={role.value} value={role.value}>
              {role.label}
            </Option>
          ))}
        </ControlledField>
        <Field
          label={__("Position")}
          {...register("position")}
          type="text"
          placeholder={__("e.g. CEO, CFO, etc.")}
          disabled={!user.canUpdate}
        />
        <EmailsField control={control} register={register} />
        <Field label={__("Contract start date")}>
          <Input
            {...register("contractStartDate")}
            type="date"
            disabled={!user.canUpdate}
          />
        </Field>
        <Field label={__("Contract end date")}>
          <Input
            {...register("contractEndDate")}
            type="date"
            disabled={!user.canUpdate}
          />
        </Field>
      </Card>
      <div className="flex justify-end">
        {formState.isDirty && user.canUpdate && (
          <Button type="submit" disabled={isMutating}>
            {__("Update")}
          </Button>
        )}
      </div>
    </form>
  );
}
