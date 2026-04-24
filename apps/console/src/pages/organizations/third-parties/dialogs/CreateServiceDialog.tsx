// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { cleanFormData } from "@probo/helpers";
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
import { type ReactNode } from "react";
import { graphql } from "relay-runtime";
import { z } from "zod";

import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

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
          ...ThirdPartyServicesTabFragment_service
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
  const { __ } = useTranslate();

  const schema = z.object({
    name: z.string().min(1, __("Service name is required")),
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
  const [createService, isLoading] = useMutationWithToasts(
    createServiceMutation,
    {
      successMessage: __("Service created successfully."),
      errorMessage: __("Failed to create service"),
    },
  );

  const onSubmit = async (data: z.infer<typeof schema>) => {
    const cleanData = cleanFormData(data);

    await createService({
      variables: {
        input: {
          thirdPartyId,
          ...cleanData,
        },
        connections: [connectionId],
      },
      onSuccess: () => {
        dialogRef.current?.close();
        reset();
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
        <Breadcrumb items={[__("Services"), __("New Service")]} />
      }
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <Field
            label={__("Name")}
            {...register("name")}
            type="text"
            error={formState.errors.name?.message}
            placeholder={__("Service name")}
            required
          />
          <Field
            label={__("Description")}
            {...register("description")}
            type="textarea"
            error={formState.errors.description?.message}
            placeholder={__("Brief description of the service")}
          />
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isLoading}>
            {__("Create")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
