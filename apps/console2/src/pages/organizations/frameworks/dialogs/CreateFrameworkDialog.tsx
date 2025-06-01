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
  type DialogRef,
} from "@probo/ui";
import { z } from "zod";
import { useFormWithSchema } from "/hooks/useFormWithSchema";

const createFrameworkMutation = graphql`
  mutation CreateFrameworkDialogMutation(
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

type Props = {
  connectionId: string;
  organizationId: string;
  ref: DialogRef;
};

const schema = z.object({
  name: z.string().min(1).max(255),
  description: z.string().max(255).optional(),
});

export function CreateFrameworkDialog(props: Props) {
  const { __ } = useTranslate();
  const { register, handleSubmit, reset } = useFormWithSchema(schema, {});
  const [commitCreate, isCreating] = useMutationWithToasts(
    createFrameworkMutation,
    {
      successMessage: __("Framework created successfully"),
      errorMessage: __("Failed to create framework"),
    }
  );
  const onSubmit = handleSubmit(async (data) => {
    await commitCreate({
      variables: {
        input: {
          ...data,
          organizationId: props.organizationId,
        },
        connections: [props.connectionId],
      },
    });
    reset();
    props.ref.current?.close();
  });

  return (
    <Dialog
      defaultOpen
      ref={props.ref}
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
          <Button type="submit" disabled={isCreating}>
            {__("Create framework")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
