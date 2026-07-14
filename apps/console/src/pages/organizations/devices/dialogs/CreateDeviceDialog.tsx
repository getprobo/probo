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

import { formatError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { type ReactNode, useState } from "react";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { CreateDeviceDialogMutation } from "#/__generated__/core/CreateDeviceDialogMutation.graphql";
import { PeopleSelectField } from "#/components/form/PeopleSelectField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

import { EnrollmentInstructions } from "../../employee/_components/EnrollmentInstructions";

const createDeviceMutation = graphql`
  mutation CreateDeviceDialogMutation($input: CreateDeviceInput!) {
    createDevice(input: $input) {
      enrollmentToken
      serverUrl
      device {
        id
      }
    }
  }
`;

const schema = z.object({
  ownerId: z.string().nullable().optional(),
});

type Props = {
  children: ReactNode;
  organizationId: string;
  onCreated: () => void;
};

export function CreateDeviceDialog({
  children,
  organizationId,
  onCreated,
}: Props) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const dialogRef = useDialogRef();
  const [enrollment, setEnrollment] = useState<{
    enrollmentToken: string;
    serverUrl: string;
  } | null>(null);

  const { control, handleSubmit, formState, reset } = useFormWithSchema(schema, {
    defaultValues: {
      ownerId: "",
    },
  });

  const [createDevice, isCreating] = useMutation<CreateDeviceDialogMutation>(
    createDeviceMutation,
  );

  const handleClose = () => {
    const createdDevice = enrollment !== null;
    setEnrollment(null);
    reset();
    if (createdDevice) {
      onCreated();
    }
  };

  const closeDialog = () => {
    handleClose();
    dialogRef.current?.close();
  };

  const onSubmit = handleSubmit((formData) => {
    const ownerId
      = formData.ownerId === null
        ? null
        : formData.ownerId || undefined;

    createDevice({
      variables: {
        input: {
          organizationId,
          ownerId,
        },
      },
      onCompleted(response, errors) {
        if (errors?.length) {
          toast({
            title: __("Error"),
            description: errors[0].message,
            variant: "error",
          });
          return;
        }

        setEnrollment({
          enrollmentToken: response.createDevice.enrollmentToken,
          serverUrl: response.createDevice.serverUrl,
        });
        toast({
          title: __("Success"),
          description: __(
            "Device created. Copy the enrollment token now — it will not be shown again.",
          ),
          variant: "success",
        });
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(
            __("Failed to create device"),
            error,
          ),
          variant: "error",
        });
      },
    });
  });

  return (
    <Dialog
      ref={dialogRef}
      trigger={children}
      onClose={handleClose}
      closable={!isCreating}
      title={<Breadcrumb items={[__("Devices"), __("New device")]} />}
    >
      <form onSubmit={e => void onSubmit(e)} className="space-y-4">
        <DialogContent padded className="space-y-4">
          {enrollment
            ? (
                <EnrollmentInstructions
                  enrollmentToken={enrollment.enrollmentToken}
                  serverUrl={enrollment.serverUrl}
                />
              )
            : (
                <PeopleSelectField
                  organizationId={organizationId}
                  control={control}
                  name="ownerId"
                  label={__("Owner")}
                  optional
                />
              )}
        </DialogContent>
        {enrollment
          ? (
              <footer className="flex justify-end items-center p-3 border-t border-t-border-low gap-2">
                <Button
                  type="button"
                  onClick={closeDialog}
                >
                  {__("Close")}
                </Button>
              </footer>
            )
          : (
              <DialogFooter>
                <Button disabled={formState.isSubmitting || isCreating} type="submit">
                  {__("Create")}
                </Button>
              </DialogFooter>
            )}
      </form>
    </Dialog>
  );
}
