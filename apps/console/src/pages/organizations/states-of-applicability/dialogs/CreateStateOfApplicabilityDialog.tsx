import { useTranslate } from "@probo/i18n";
import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  useDialogRef,
} from "@probo/ui";
import type { ReactNode } from "react";
import { Suspense } from "react";
import { graphql } from "react-relay";
import { useNavigate } from "react-router";
import { z } from "zod";

import type { CreateStateOfApplicabilityDialogMutation } from "#/__generated__/core/CreateStateOfApplicabilityDialogMutation.graphql";
import { PeopleSelectField } from "#/components/form/PeopleSelectField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const createMutation = graphql`
    mutation CreateStateOfApplicabilityDialogMutation(
        $input: CreateStateOfApplicabilityInput!
        $connections: [ID!]!
    ) {
        createStateOfApplicability(input: $input) {
            stateOfApplicabilityEdge @prependEdge(connections: $connections) {
                node {
                    id
                    name
                    sourceId
                    snapshotId
                    createdAt
                    updatedAt
                    canDelete: permission(action: "core:state-of-applicability:delete")
                    ...StateOfApplicabilityRowFragment
                }
            }
        }
    }
`;

type Props = {
  children: ReactNode;
  connectionId: string;
};

const schema = z.object({
  name: z.string().min(1),
  ownerId: z.string().min(1),
});

export function CreateStateOfApplicabilityDialog({
  children,
  connectionId,
}: Props) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();
  const { control, register, handleSubmit, reset } = useFormWithSchema(
    schema,
    {
      defaultValues: {
        name: "",
        ownerId: "",
      },
    },
  );
  const ref = useDialogRef();

  const [createStateOfApplicability, isCreating]
    = useMutationWithToasts<CreateStateOfApplicabilityDialogMutation>(
      createMutation,
      {
        successMessage: __(
          "State of applicability created successfully.",
        ),
        errorMessage: __("Failed to create state of applicability"),
      },
    );

  const onSubmit = async (data: z.infer<typeof schema>) => {
    await createStateOfApplicability({
      variables: {
        input: {
          name: data.name,
          organizationId,
          ownerId: data.ownerId,
        },
        connections: [connectionId],
      },
      onCompleted: (response) => {
        reset();
        ref.current?.close();
        const stateOfApplicabilityId
          = response.createStateOfApplicability.stateOfApplicabilityEdge
            .node.id;
        void navigate(
          `/organizations/${organizationId}/states-of-applicability/${stateOfApplicabilityId}`,
        );
      },
    });
  };

  return (
    <Dialog
      ref={ref}
      trigger={children}
      title={(
        <Breadcrumb
          items={[
            __("States of Applicability"),
            __("New State of Applicability"),
          ]}
        />
      )}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <Field
            label={__("Name")}
            {...register("name")}
            type="text"
            required
          />
          <Field label={__("Owner")}>
            <Suspense fallback={<div>{__("Loading...")}</div>}>
              <PeopleSelectField
                organizationId={organizationId}
                control={control}
                name="ownerId"
              />
            </Suspense>
          </Field>
        </DialogContent>
        <DialogFooter>
          <Button disabled={isCreating} type="submit">
            {__("Create")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
