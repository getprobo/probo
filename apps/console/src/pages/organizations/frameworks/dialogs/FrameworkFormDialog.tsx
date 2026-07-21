// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
import { useTranslation } from "react-i18next";
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
  const { t } = useTranslation();
  const newRef = useDialogRef();
  const dialogRef = ref ?? newRef;
  const { register, handleSubmit, reset } = useFormWithSchema(schema, {
    defaultValues: {
      name: framework?.name ?? "",
      description: framework?.description ?? "",
    },
  });
  const [create, isCreating] = useMutationWithToasts(createFrameworkMutation, {
    successMessage: t("frameworkFormDialog.messages.created"),
    errorMessage: t("frameworkFormDialog.errors.create"),
  });
  const [update, isUpdating] = useMutationWithToasts(updateFrameworkMutation, {
    successMessage: t("frameworkFormDialog.messages.updated"),
    errorMessage: t("frameworkFormDialog.errors.update"),
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
      title={<Breadcrumb items={[t("frameworkFormDialog.breadcrumb.framework"), t("frameworkFormDialog.breadcrumb.newFramework")]} />}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <Input
            {...register("name")}
            variant="title"
            required
            placeholder={t("frameworkFormDialog.fields.title")}
          />
          <Textarea
            {...register("description")}
            variant="ghost"
            autogrow
            placeholder={t("frameworkFormDialog.fields.description")}
          />
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isCreating || isUpdating}>
            {framework ? t("frameworkFormDialog.actions.update") : t("frameworkFormDialog.actions.create")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
