import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Input,
  Label,
  PropertyRow,
  Textarea,
  useDialogRef,
} from "@probo/ui";
import type { ReactNode } from "react";
import { useTranslate } from "@probo/i18n";
import { Breadcrumb } from "@probo/ui";
import { graphql } from "relay-runtime";
import { z } from "zod";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { ControlledField } from "/components/form/ControlledField";
import { SnapshotTypeOptions } from "/components/form/SnapshotTypeOptions";

const snapshotCreateMutation = graphql`
  mutation SnapshotFormDialogCreateMutation(
    $input: CreateSnapshotInput!
    $connections: [ID!]!
  ) {
    createSnapshot(input: $input) {
      snapshotEdge @prependEdge(connections: $connections) {
        node {
          id
          name
          description
          type
          createdAt
        }
      }
    }
  }
`;

const snapshotSchema = z.object({
  name: z.string().min(2, { message: "Name is required" }),
  description: z.string().optional(),
  type: z.enum(["RISKS", "VENDORS", "ASSETS", "DATA", "NON_CONFORMITY_REGISTRIES", "COMPLIANCE_REGISTRIES"]),
});

type Props = {
  children?: ReactNode;
  connection?: string;
};

export default function SnapshotFormDialog(props: Props) {
  const { __ } = useTranslate();
  const dialogRef = useDialogRef();
  const organizationId = useOrganizationId();
  const [mutate] = useMutationWithToasts(snapshotCreateMutation, {
    successMessage: __("Snapshot created successfully."),
    errorMessage: __("Failed to create snapshot. Please try again."),
  });

  const { handleSubmit, register, reset, control, formState: { errors } } =
    useFormWithSchema(snapshotSchema, {
      defaultValues: {
        name: "",
        description: "",
        type: "RISKS",
      },
    });

  const onSubmit = handleSubmit(async (data) => {
    await mutate({
      variables: {
        input: {
          organizationId,
          name: data.name,
          description: data.description || undefined,
          type: data.type,
        },
        connections: props.connection ? [props.connection] : [],
      },
    });
    reset();
    dialogRef.current?.close();
  });

  return (
    <Dialog
      ref={dialogRef}
      trigger={props.children}
      title={
        <Breadcrumb
          items={[
            __("Snapshots"),
            __("New Snapshot"),
          ]}
        />
      }
    >
      <form onSubmit={onSubmit}>
        <DialogContent className="grid grid-cols-[1fr_420px]">
          <div className="py-8 px-10 space-y-4">
            <Input
              id="name"
              required
              variant="title"
              placeholder={__("Snapshot name")}
              {...register("name")}
            />
            <Textarea
              id="description"
              variant="ghost"
              autogrow
              placeholder={__("Add description (optional)")}
              {...register("description")}
            />
          </div>
          {/* Properties form */}
          <div className="py-5 px-6 bg-subtle">
            <Label>{__("Properties")}</Label>

            <PropertyRow
              id="type"
              label={__("Type")}
              error={errors.type?.message}
            >
              <ControlledField
                control={control}
                name="type"
                type="select"
              >
                <SnapshotTypeOptions />
              </ControlledField>
            </PropertyRow>
          </div>
        </DialogContent>
        <DialogFooter>
          <Button type="submit">
            {__("Create snapshot")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
