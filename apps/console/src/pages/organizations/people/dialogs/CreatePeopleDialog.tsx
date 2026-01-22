import { getRoles, peopleRoles } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  Option,
  useDialogRef,
} from "@probo/ui";
import type { ReactNode } from "react";
import { graphql } from "relay-runtime";
import { z } from "zod";

import { ControlledField } from "#/components/form/ControlledField";
import { EmailsField } from "#/components/form/EmailsField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

type Props = {
  children: ReactNode;
  connectionId: string;
};

const schema = z.object({
  fullName: z.string().min(1),
  position: z.string().min(1),
  primaryEmailAddress: z.string().email(),
  additionalEmailAddresses: z.preprocess(
    // Empty additional emails are skipped
    v => (v as string[]).filter(v => !!v),
    z.array(z.string().email()),
  ),
  kind: z.enum(peopleRoles),
});

export const createPeopleMutation = graphql`
  mutation CreatePeopleDialogMutation(
    $input: CreatePeopleInput!
    $connections: [ID!]!
  ) {
    createPeople(input: $input) {
      peopleEdge @prependEdge(connections: $connections) {
        node {
          id
          fullName
          primaryEmailAddress
          position
          kind
          additionalEmailAddresses
          canDelete: permission(action: "core:people:delete")
          canUpdate: permission(action: "core:people:update")
        }
      }
    }
  }
`;

export function CreatePeopleDialog({ children, connectionId }: Props) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const { control, handleSubmit, register, reset } = useFormWithSchema(schema, {
    defaultValues: {
      additionalEmailAddresses: [],
    },
  });
  const ref = useDialogRef();

  const [mutate, isMutating] = useMutationWithToasts(createPeopleMutation, {
    successMessage: __("Person created successfully."),
    errorMessage: __("Failed to create person"),
  });

  const onSubmit = async (data: z.infer<typeof schema>) => {
    await mutate({
      variables: {
        input: {
          ...data,
          organizationId,
          additionalEmailAddresses: data.additionalEmailAddresses ?? [],
        },
        connections: [connectionId],
      },
      onSuccess: () => {
        reset();
        ref.current?.close();
      },
    });
  };

  return (
    <Dialog
      ref={ref}
      trigger={children}
      title={<Breadcrumb items={[__("People"), __("New Person")]} />}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <Field
            label={__("Full name")}
            {...register("fullName")}
            type="text"
          />
          <Field
            label={__("Primary email")}
            {...register("primaryEmailAddress")}
            type="email"
          />
          <Field
            label={__("Position")}
            {...register("position")}
            type="text"
            placeholder={__("e.g. CEO, CFO, etc.")}
          />
          <ControlledField
            control={control}
            name="kind"
            type="select"
            label={__("Role")}
          >
            {getRoles(__).map(role => (
              <Option key={role.value} value={role.value}>
                {role.label}
              </Option>
            ))}
          </ControlledField>
          <EmailsField control={control} register={register} />
        </DialogContent>
        <DialogFooter>
          <Button disabled={isMutating} type="submit">
            {__("Create")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
