import { graphql } from "relay-runtime";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { useTranslate } from "@probo/i18n";
import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Input,
  Textarea,
  useDialogRef,
  type DialogRef,
} from "@probo/ui";
import { z } from "zod";
import { useFormWithSchema } from "/hooks/useFormWithSchema";

const createFrameworkMutation = graphql`
  mutation FrameworkFormDialogMutation(
    $input: CreateFrameworkInput!
    $connections: [ID!]!
  ) {
    createFramework(input: $input) {
      frameworkEdge @prependEdge(connections: $connections) {
        node {
          id
          ...FrameworksPageCardFragment
        }
      }
    }
  }
`;

const updateFrameworkMutation = graphql`
  mutation FrameworkFormDialogUpdateMutation($input: UpdateFrameworkInput!) {
    updateFramework(input: $input) {
      framework {
        id
        name
        description
      }
    }
  }
`;

type Props = {
  connectionId?: string;
  organizationId: string;
  framework?: {
    id: string;
    name: string;
    description: string;
  };
  ref?: DialogRef;
  children?: React.ReactNode;
};

const schema = z.object({
  name: z.string().min(1).max(255),
  description: z.string().max(255).optional(),
});

/**
 * Form to update or create a new framework
 */
export function FrameworkFormDialog(props: Props) {
  const { __ } = useTranslate();
  const dialogRef = props.ref ?? useDialogRef();
  const { register, handleSubmit, reset } = useFormWithSchema(schema, {
    defaultValues: {
      name: props.framework?.name ?? "",
      description: props.framework?.description ?? "",
    },
  });
  const [create, isCreating] = useMutationWithToasts(createFrameworkMutation, {
    successMessage: __("Framework created successfully"),
    errorMessage: __("Failed to create framework"),
  });
  const [update, isUpdating] = useMutationWithToasts(updateFrameworkMutation, {
    successMessage: __("Framework updated successfully"),
    errorMessage: __("Failed to update framework"),
  });
  const onSubmit = handleSubmit(async (data) => {
    if (props.framework) {
      await update({
        variables: {
          input: {
            id: props.framework.id,
            ...data,
          },
        },
      });
      reset(data);
      dialogRef.current?.close();
      return;
    }
    await create({
      variables: {
        input: {
          ...data,
          organizationId: props.organizationId,
        },
        connections: [props.connectionId],
      },
    });
    reset();
    dialogRef.current?.close();
  });

  return (
    <Dialog
      trigger={props.children}
      ref={dialogRef}
      title={<Breadcrumb items={[__("Framework"), __("New Framework")]} />}
    >
      <form onSubmit={onSubmit}>
        <DialogContent padded className="space-y-4">
          <Input
            {...register("name")}
            variant="title"
            required
            placeholder={__("Framework title")}
          />
          <Textarea
            {...register("description")}
            variant="ghost"
            autogrow
            placeholder={__("Add description")}
          />
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isCreating || isUpdating}>
            {props.framework ? __("Update framework") : __("Create framework")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
