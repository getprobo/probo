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

import { formatError, type GraphQLError } from "@probo/helpers";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Dialog } from "@probo/ui/src/v2/Dialog/Dialog";
import { DialogBody } from "@probo/ui/src/v2/Dialog/DialogBody";
import { DialogClose } from "@probo/ui/src/v2/Dialog/DialogClose";
import { DialogFooter } from "@probo/ui/src/v2/Dialog/DialogFooter";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import { type FormEvent, useState } from "react";
import { useTranslation } from "react-i18next";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { CreateDeviceDialogMutation } from "#/__generated__/core/CreateDeviceDialogMutation.graphql";

import { DeviceOwnerSelect } from "../_components/DeviceOwnerSelect";
import { EnrollmentInstructions } from "../_components/EnrollmentInstructions";

const createDeviceMutation = graphql`
  mutation CreateDeviceDialogMutation($input: CreateDeviceInput!) {
    createDevice(input: $input) {
      enrollmentToken
      device {
        id
      }
    }
  }
`;

interface CreateDeviceDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  organizationId: string;
  onCreated: () => void;
}

export function CreateDeviceDialog({
  open,
  onOpenChange,
  organizationId,
  onCreated,
}: CreateDeviceDialogProps) {
  const { t } = useTranslation("organizations/devices");
  const toast = useToast();
  const [ownerId, setOwnerId] = useState<string | null>(null);
  const [enrollmentToken, setEnrollmentToken] = useState<string | null>(null);
  const [prevOpen, setPrevOpen] = useState(open);
  const [createDevice, isCreating] = useMutation<CreateDeviceDialogMutation>(createDeviceMutation);

  // Reset on the next open. Clearing the token while the dialog closes
  // replaces Close with the form's submit button and creates another device.
  if (open && !prevOpen) {
    setPrevOpen(true);
    setOwnerId(null);
    setEnrollmentToken(null);
  } else if (!open && prevOpen) {
    setPrevOpen(false);
  }

  function handleOpenChange(next: boolean) {
    if (!next && isCreating) {
      return;
    }
    if (!next && enrollmentToken !== null) {
      onCreated();
    }
    onOpenChange(next);
  }

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (enrollmentToken !== null || isCreating) {
      return;
    }
    createDevice({
      variables: {
        input: {
          organizationId,
          ownerId: ownerId ?? undefined,
        },
      },
      onCompleted(response, errors) {
        if (errors?.length) {
          toast.add({
            title: t("common.error"),
            description: errors[0].message,
            type: "error",
          });
          return;
        }
        setEnrollmentToken(response.createDevice.enrollmentToken);
        toast.add({
          title: t("common.success"),
          description: t("deviceEnrollment.messages.created"),
          type: "success",
        });
      },
      onError(error) {
        toast.add({
          title: t("common.error"),
          description: formatError(t("devices.errors.create"), error as GraphQLError),
          type: "error",
        });
      },
    });
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogPopup>
        <DialogHeader>
          <DialogTitle>{t("devices.actions.new")}</DialogTitle>
        </DialogHeader>
        {enrollmentToken
          ? (
              <>
                <DialogBody>
                  <EnrollmentInstructions enrollmentToken={enrollmentToken} />
                </DialogBody>
                <DialogFooter>
                  <DialogClose
                    render={<Button>{t("common.actions.close")}</Button>}
                  />
                </DialogFooter>
              </>
            )
          : (
              <form className="flex flex-col gap-4" onSubmit={handleSubmit}>
                <DialogBody>
                  <DeviceOwnerSelect
                    organizationId={organizationId}
                    profileKey={null}
                    value={ownerId}
                    onValueChange={setOwnerId}
                  />
                </DialogBody>
                <DialogFooter>
                  <Button type="submit" loading={isCreating}>
                    {t("common.create")}
                  </Button>
                </DialogFooter>
              </form>
            )}
      </DialogPopup>
    </Dialog>
  );
}
