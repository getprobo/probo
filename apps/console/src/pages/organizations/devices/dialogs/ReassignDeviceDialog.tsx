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
import { useEffect, useMemo, useState } from "react";
import { useFragment, useMutation } from "react-relay";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { ReassignDeviceDialog_device$key } from "#/__generated__/core/ReassignDeviceDialog_device.graphql";
import type { ReassignDeviceDialogMutation } from "#/__generated__/core/ReassignDeviceDialogMutation.graphql";
import { PeopleSelectField } from "#/components/form/PeopleSelectField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

const reassignDeviceDialogFragment = graphql`
  fragment ReassignDeviceDialog_device on Device {
    id
    owner {
      id
    }
  }
`;

const reassignDeviceMutation = graphql`
  mutation ReassignDeviceDialogMutation($input: SetDeviceOwnerInput!) {
    setDeviceOwner(input: $input) {
      device {
        id
        owner {
          id
          fullName
        }
      }
    }
  }
`;

const schema = z.object({
  ownerId: z.string().nullable().optional(),
});

interface ReassignDeviceDialogProps {
  deviceKey: ReassignDeviceDialog_device$key;
  organizationId: string;
  ref?: ReturnType<typeof useDialogRef>;
}

export function ReassignDeviceDialog({
  deviceKey,
  organizationId,
  ref: refProps,
}: ReassignDeviceDialogProps) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const dialogRef = useDialogRef();
  const ref = refProps ?? dialogRef;

  const device = useFragment(reassignDeviceDialogFragment, deviceKey);
  const [open, setOpen] = useState(false);

  const defaultValues = useMemo(
    () => ({
      ownerId: device.owner?.id ?? null,
    }),
    [device.owner?.id],
  );

  const { control, handleSubmit, formState, reset } = useFormWithSchema(schema, {
    defaultValues,
  });

  useEffect(() => {
    reset(defaultValues);
  }, [defaultValues, reset]);

  const handleClose = () => {
    reset(defaultValues);
  };

  const [setDeviceOwner, isInFlight] = useMutation<ReassignDeviceDialogMutation>(
    reassignDeviceMutation,
  );

  const onSubmit = (formData: z.infer<typeof schema>) => {
    const ownerId = formData.ownerId ?? undefined;

    setDeviceOwner({
      variables: {
        input: {
          deviceId: device.id,
          ownerId,
        },
      },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast({
            title: __("Error"),
            description: errors[0].message,
            variant: "error",
          });
          return;
        }

        toast({
          title: __("Success"),
          description: __("Device owner updated"),
          variant: "success",
        });
        handleClose();
        ref.current?.close();
      },
      onError(error) {
        toast({
          title: __("Error"),
          description: formatError(
            __("Failed to re-assign device"),
            error,
          ),
          variant: "error",
        });
      },
    });
  };

  return (
    <Dialog
      ref={ref}
      onClose={handleClose}
      onOpenChange={setOpen}
      title={<Breadcrumb items={[__("Devices"), __("Re-assign")]} />}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)} className="space-y-4">
        <DialogContent padded className="space-y-4">
          {open && (
            <PeopleSelectField
              organizationId={organizationId}
              control={control}
              name="ownerId"
              label={__("Owner")}
              optional
            />
          )}
        </DialogContent>
        <DialogFooter>
          <Button disabled={formState.isSubmitting || isInFlight} type="submit">
            {__("Re-assign")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
