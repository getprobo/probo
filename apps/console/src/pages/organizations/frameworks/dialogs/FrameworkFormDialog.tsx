import { useTranslate } from "@probo/i18n";
import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  type DialogRef,
  Input,
  Textarea,
  useDialogRef,
} from "@probo/ui";
import { graphql } from "relay-runtime";
import { z } from "zod";

import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

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
    description?: string | null;
  };
  ref?: DialogRef;
  children?: React.ReactNode;
};

const schema = z.object({
  name: z.string().min(1).max(255),
  description: z.string().max(255).optional().nullable(),
});

/**
 * Form to update or create a new framework
 */
export function FrameworkFormDialog(props: Props) {
  const { children, connectionId, ref, framework, organizationId } = props;
  const { __ } = useTranslate();
  const newRef = useDialogRef();
  const dialogRef = ref ?? newRef;
  const { register, handleSubmit, reset } = useFormWithSchema(schema, {
    defaultValues: {
      name: framework?.name ?? "",
      description: framework?.description ?? "",
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
  const onSubmit = async (data: z.infer<typeof schema>) => {
    if (framework) {
      await update({
        variables: {
          input: {
            id: framework.id,
            ...data,
            description: data.description || null,
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
          description: data.description || null,
          organizationId: organizationId,
        },
        connections: [connectionId],
      },
    });
    reset();
    dialogRef.current?.close();
  };

  return (
    <Dialog
      trigger={children}
      ref={dialogRef}
      title={<Breadcrumb items={[__("Framework"), __("New Framework")]} />}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
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
            {framework ? __("Update framework") : __("Create framework")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
