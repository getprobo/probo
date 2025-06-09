import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Input,
  Textarea,
  useDialogRef,
} from "@probo/ui";
import type { ReactNode } from "react";
import { useTranslate } from "@probo/i18n";
import { graphql } from "relay-runtime";
import { useFragment } from "react-relay";
import type { FrameworkControlDialogFragment$key } from "./__generated__/FrameworkControlDialogFragment.graphql";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { z } from "zod";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";

type Props = {
  children: ReactNode;
  control?: FrameworkControlDialogFragment$key;
  frameworkId: string;
  connectionId?: string;
};

const controlFragment = graphql`
  fragment FrameworkControlDialogFragment on Control {
    id
    name
    description
    sectionTitle
  }
`;

const createMutation = graphql`
  mutation FrameworkControlDialogCreateMutation(
    $input: CreateControlInput!
    $connections: [ID!]!
  ) {
    createControl(input: $input) {
      controlEdge @prependEdge(connections: $connections) {
        node {
          ...FrameworkControlDialogFragment
        }
      }
    }
  }
`;

const updateMutation = graphql`
  mutation FrameworkControlDialogUpdateMutation($input: UpdateControlInput!) {
    updateControl(input: $input) {
      control {
        ...FrameworkControlDialogFragment
      }
    }
  }
`;

const schema = z.object({
  name: z.string(),
  description: z.string(),
  sectionTitle: z.string(),
});

export function FrameworkControlDialog(props: Props) {
  const { __ } = useTranslate();
  const frameworkControl = useFragment(controlFragment, props.control);
  const dialogRef = useDialogRef();
  const [mutate, isMutating] = props.control
    ? useMutationWithToasts(updateMutation, {
        successMessage: __("Control updated successfully."),
        errorMessage: __("Failed to update control. Please try again."),
      })
    : useMutationWithToasts(createMutation, {
        successMessage: __("Control created successfully."),
        errorMessage: __("Failed to create control. Please try again."),
      });
  const { control, handleSubmit, register } = useFormWithSchema(schema, {
    defaultValues: {
      name: frameworkControl?.name ?? "",
      description: frameworkControl?.description ?? "",
      sectionTitle: frameworkControl?.sectionTitle ?? "",
    },
  });

  const onSubmit = handleSubmit(async (data) => {
    if (frameworkControl) {
      // Update the control
      await mutate({
        variables: {
          input: {
            id: frameworkControl.id,
            name: data.name,
            description: data.description,
            sectionTitle: data.sectionTitle,
          },
        },
      });
    } else {
      // Create a new control
      await mutate({
        variables: {
          input: {
            frameworkId: props.frameworkId,
            name: data.name,
            description: data.description,
            sectionTitle: data.sectionTitle,
          },
          connections: [props.connectionId!],
        },
      });
    }
    dialogRef.current?.close();
  });

  return (
    <Dialog
      trigger={props.children}
      ref={dialogRef}
      title={
        <Breadcrumb
          items={[
            __("Controls"),
            control ? __("Edit Control") : __("New Control"),
          ]}
        />
      }
    >
      <form onSubmit={onSubmit}>
        <DialogContent padded className="space-y-2">
          <Input
            id="sectionTitle"
            required
            variant="ghost"
            placeholder={__("Section title")}
            {...register("sectionTitle")}
          />
          <Input
            id="title"
            required
            variant="title"
            placeholder={__("Document title")}
            {...register("name")}
          />
          <Textarea
            id="content"
            variant="ghost"
            required
            autogrow
            placeholder={__("Add description")}
            {...register("description")}
          />
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isMutating}>
            {props.control ? __("Update control") : __("Create control")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
