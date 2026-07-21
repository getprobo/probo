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

import { cleanFormData, formatError } from "@probo/helpers";
import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { type ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { CreateServiceDialogMutation } from "#/__generated__/core/CreateServiceDialogMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

type Props = {
  children: ReactNode;
  connectionId: string;
  thirdPartyId: string;
};

const createServiceMutation = graphql`
  mutation CreateServiceDialogMutation(
    $input: CreateThirdPartyServiceInput!
    $connections: [ID!]!
  ) {
    createThirdPartyService(input: $input) {
      thirdPartyServiceEdge @prependEdge(connections: $connections) {
        node {
          ...ThirdPartyServiceRow_service
        }
      }
    }
  }
`;

export function CreateServiceDialog({
  children,
  connectionId,
  thirdPartyId,
}: Props) {
  const { t } = useTranslation();

  const schema = z.object({
    name: z.string().min(1, t("createThirdPartyServiceDialog.validation.nameRequired")),
    description: z.string().optional(),
  });

  const { register, handleSubmit, formState, reset } = useFormWithSchema(
    schema,
    {
      defaultValues: {
        name: "",
        description: "",
      },
    },
  );
  const { toast } = useToast();
  const [createService, isCreating] = useMutation<CreateServiceDialogMutation>(
    createServiceMutation,
  );

  const onSubmit = (data: z.infer<typeof schema>) => {
    const cleanData = cleanFormData(data);

    createService({
      variables: {
        input: {
          thirdPartyId,
          ...cleanData,
        },
        connections: [connectionId],
      },
      onCompleted(_response, errors) {
        if (errors) {
          toast({
            title: t("createThirdPartyServiceDialog.messages.error"),
            description: formatError(t("createThirdPartyServiceDialog.errors.create"), errors),
            variant: "error",
          });
          return;
        }
        toast({
          title: t("createThirdPartyServiceDialog.messages.success"),
          description: t("createThirdPartyServiceDialog.messages.created"),
          variant: "success",
        });
        dialogRef.current?.close();
        reset();
      },
      onError(error) {
        toast({
          title: t("createThirdPartyServiceDialog.messages.error"),
          description: formatError(t("createThirdPartyServiceDialog.errors.create"), error),
          variant: "error",
        });
      },
    });
  };

  const dialogRef = useDialogRef();

  return (
    <Dialog
      className="max-w-lg"
      ref={dialogRef}
      trigger={children}
      title={
        <Breadcrumb items={[t("createThirdPartyServiceDialog.breadcrumb.services"), t("createThirdPartyServiceDialog.breadcrumb.newService")]} />
      }
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <Field
            label={t("createThirdPartyServiceDialog.fields.name")}
            {...register("name")}
            type="text"
            error={formState.errors.name?.message}
            placeholder={t("createThirdPartyServiceDialog.placeholders.name")}
            required
          />
          <Field
            label={t("createThirdPartyServiceDialog.fields.description")}
            {...register("description")}
            type="textarea"
            error={formState.errors.description?.message}
            placeholder={t("createThirdPartyServiceDialog.placeholders.description")}
          />
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isCreating}>
            {t("createThirdPartyServiceDialog.actions.create")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
