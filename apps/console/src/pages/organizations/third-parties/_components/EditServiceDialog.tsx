// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { cleanFormData, formatError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
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
import { useEffect } from "react";
import { graphql, useFragment, useMutation } from "react-relay";
import { z } from "zod";

import type { EditServiceDialog_service$key } from "#/__generated__/core/EditServiceDialog_service.graphql";
import type { EditServiceDialogUpdateMutation } from "#/__generated__/core/EditServiceDialogUpdateMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

type Props = {
  serviceKey: EditServiceDialog_service$key;
  onClose: () => void;
};

const editServiceDialogFragment = graphql`
  fragment EditServiceDialog_service on ThirdPartyService {
    id
    name
    description
  }
`;

const updateServiceMutation = graphql`
  mutation EditServiceDialogUpdateMutation($input: UpdateThirdPartyServiceInput!) {
    updateThirdPartyService(input: $input) {
      thirdPartyService {
        ...ThirdPartyServiceRow_service
        ...EditServiceDialog_service
      }
    }
  }
`;

export function EditServiceDialog({ serviceKey, onClose }: Props) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const service = useFragment(editServiceDialogFragment, serviceKey);

  const schema = z.object({
    name: z.string().min(1, __("Service name is required")),
    description: z.string().optional(),
  });

  const { register, handleSubmit, formState } = useFormWithSchema(
    schema,
    {
      defaultValues: {
        name: service.name || "",
        description: service.description || "",
      },
    },
  );

  const [updateService, isUpdating]
    = useMutation<EditServiceDialogUpdateMutation>(updateServiceMutation);

  const onSubmit = (data: z.infer<typeof schema>) => {
    const cleanData = cleanFormData(data);

    updateService({
      variables: {
        input: {
          id: service.id,
          ...cleanData,
        },
      },
      onCompleted(_response, errors) {
        if (errors) {
          toast({
            title: __("Error"),
            description: formatError(__("Failed to update service"), errors),
            variant: "error",
          });
          return;
        }
        toast({
          title: __("Success"),
          description: __("Service updated successfully."),
          variant: "success",
        });
        onClose();
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(__("Failed to update service"), error),
          variant: "error",
        });
      },
    });
  };

  const dialogRef = useDialogRef();

  useEffect(() => {
    dialogRef.current?.open();
  }, [dialogRef]);

  return (
    <Dialog
      className="max-w-lg"
      ref={dialogRef}
      onClose={onClose}
      title={
        <Breadcrumb items={[__("Services"), __("Edit Service")]} />
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
          <Button type="submit" disabled={isUpdating}>
            {__("Save")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
