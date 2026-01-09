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
import { useNavigate } from "react-router";
import { useOrganizationId } from "/hooks/useOrganizationId";
import z from "zod";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { createStateOfApplicabilityMutation } from "/hooks/graph/StateOfApplicabilityGraph";
import type { StateOfApplicabilityGraphCreateMutation } from "/hooks/graph/__generated__/StateOfApplicabilityGraphCreateMutation.graphql";

type Props = {
  children: ReactNode;
  connectionId: string;
};

const schema = z.object({
  name: z.string().min(1),
  description: z.string().optional(),
});

export function CreateStateOfApplicabilityDialog({
  children,
  connectionId,
}: Props) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();
  const { register, handleSubmit, reset } = useFormWithSchema(schema, {
    defaultValues: {
      name: "",
      description: "",
    },
  });
  const ref = useDialogRef();

  const [mutate, isMutating] = useMutationWithToasts<StateOfApplicabilityGraphCreateMutation>(
    createStateOfApplicabilityMutation,
    {
      successMessage: __("State of applicability created successfully."),
      errorMessage: __("Failed to create state of applicability"),
    },
  );

  const onSubmit = handleSubmit((data) => {
    mutate({
      variables: {
        input: {
          name: data.name,
          description: data.description || null,
          organizationId,
        },
        connections: [connectionId],
      },
      onCompleted: (response) => {
        reset();
        ref.current?.close();
        const stateOfApplicabilityId = response.createStateOfApplicability.stateOfApplicabilityEdge.node.id;
        navigate(`/organizations/${organizationId}/states-of-applicability/${stateOfApplicabilityId}`);
      },
    });
  });

  return (
    <Dialog
      ref={ref}
      trigger={children}
      title={
        <Breadcrumb
          items={[__("States of Applicability"), __("New State of Applicability")]}
        />
      }
    >
      <form onSubmit={onSubmit}>
        <DialogContent padded className="space-y-4">
          <Field
            label={__("Name")}
            {...register("name")}
            type="text"
            required
          />
          <Field
            label={__("Description")}
            {...register("description")}
            type="textarea"
          />
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
