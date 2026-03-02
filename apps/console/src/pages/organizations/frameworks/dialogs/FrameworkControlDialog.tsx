import { useTranslate } from "@probo/i18n";
import {
  Breadcrumb,
  Button,
  Checkbox,
  Dialog,
  DialogContent,
  DialogFooter,
  Input,
  Textarea,
  useDialogRef,
} from "@probo/ui";
import type { ReactNode } from "react";
import { useEffect, useMemo } from "react";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { FrameworkControlDialogFragment$key } from "#/__generated__/core/FrameworkControlDialogFragment.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

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
        bestPractice
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
  description: z.string().optional().nullable(),
  sectionTitle: z.string(),
  bestPractice: z.boolean(),
});

export function FrameworkControlDialog(props: Props) {
  const { __ } = useTranslate();
  const frameworkControl = useFragment(controlFragment, props.control);
  const dialogRef = useDialogRef();
  const [mutate, isMutating] = useMutationWithToasts(
    props.control ? updateMutation : createMutation,
    {
      successMessage: __(
        `Control ${props.control ? "updated" : "created"} successfully.`,
      ),
      errorMessage: __(
        `Failed to ${props.control ? "update" : "create"} control`,
      ),
    },
  );

  const defaultValues = useMemo(
    () => ({
      name: frameworkControl?.name ?? "",
      description: frameworkControl?.description ?? "",
      sectionTitle: frameworkControl?.sectionTitle ?? "",
      bestPractice: frameworkControl?.bestPractice ?? true,
    }),
    [frameworkControl],
  );

  const { handleSubmit, register, reset, watch, setValue }
    = useFormWithSchema(schema, {
      defaultValues,
    });

  useEffect(() => {
    reset(defaultValues);
  }, [defaultValues, reset]);

  const bestPracticeValue = watch("bestPractice");

  const onSubmit = async (data: z.infer<typeof schema>) => {
    if (frameworkControl) {
      // Update the control
      await mutate({
        variables: {
          input: {
            id: frameworkControl.id,
            name: data.name,
            description: data.description || null,
            sectionTitle: data.sectionTitle,
            bestPractice: data.bestPractice,
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
            description: data.description || null,
            sectionTitle: data.sectionTitle,
            bestPractice: data.bestPractice ?? true,
          },
          connections: [props.connectionId!],
        },
      });
      reset();
    }
    dialogRef.current?.close();
  };

  return (
    <Dialog
      trigger={props.children}
      ref={dialogRef}
      title={(
        <Breadcrumb
          items={[
            __("Controls"),
            frameworkControl
              ? __("Edit Control")
              : __("New Control"),
          ]}
        />
      )}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
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
            autogrow
            placeholder={__("Add description")}
            {...register("description")}
          />
          <label className="flex items-center gap-2 cursor-pointer">
            <Checkbox
              checked={bestPracticeValue}
              onChange={checked =>
                setValue("bestPractice", checked)}
            />
            <span className="text-sm">{__("Best Practice")}</span>
          </label>
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isMutating}>
            {props.control
              ? __("Update control")
              : __("Create control")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
