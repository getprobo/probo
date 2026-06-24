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
