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
  Button,
  Dialog,
  DialogContent,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { useState } from "react";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { CreateDeviceFormMutation } from "#/__generated__/core/CreateDeviceFormMutation.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { EnrollmentInstructions } from "./EnrollmentInstructions";

const enrollDeviceMutation = graphql`
  mutation CreateDeviceFormMutation($input: EnrollDeviceInput!) {
    enrollDevice(input: $input) {
      enrollmentToken
      serverUrl
      device {
        id
      }
    }
  }
`;

interface CreateDeviceFormProps {
  onDeviceCreated?: () => void;
}

export function CreateDeviceForm({ onDeviceCreated }: CreateDeviceFormProps) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const dialogRef = useDialogRef();

  const [enrollment, setEnrollment] = useState<{
    enrollmentToken: string;
    serverUrl: string;
  } | null>(null);

  const organizationId = useOrganizationId();
  const [enrollDevice, isCreating] = useMutation<CreateDeviceFormMutation>(
    enrollDeviceMutation,
  );

  const handleClose = () => {
    const createdDevice = enrollment !== null;
    setEnrollment(null);
    if (createdDevice) {
      onDeviceCreated?.();
    }
  };

  const closeDialog = () => {
    handleClose();
    dialogRef.current?.close();
  };

  const handleCreate = () => {
    enrollDevice({
      variables: {
        input: {
          organizationId,
        },
      },
      onCompleted(response, errors) {
        if (errors?.length) {
          toast({
            title: __("Error"),
            description: errors[0].message,
            variant: "error",
          });
          dialogRef.current?.close();
          return;
        }

        setEnrollment({
          enrollmentToken: response.enrollDevice.enrollmentToken,
          serverUrl: response.enrollDevice.serverUrl,
        });
        dialogRef.current?.open();
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
        dialogRef.current?.close();
      },
    });
  };

  const handleManualEnroll = () => {
    dialogRef.current?.open();
    if (!isCreating && !enrollment) {
      handleCreate();
    }
  };

  return (
    <>
      <p className="text-center text-xs text-txt-secondary">
        {__("Can't enroll new device?")}
        {" "}
        <button
          type="button"
          onClick={handleManualEnroll}
          disabled={isCreating}
          className="text-txt-primary underline hover:no-underline disabled:opacity-60"
        >
          {__("Try creating it manually")}
        </button>
      </p>

      <Dialog
        ref={dialogRef}
        onClose={handleClose}
        closable={!(isCreating && !enrollment)}
        title={__("Manual enrollment")}
      >
        <DialogContent padded className="space-y-4">
          {isCreating && !enrollment
            ? <p>{__("Creating device…")}</p>
            : null}
          {enrollment
            ? (
                <EnrollmentInstructions
                  enrollmentToken={enrollment.enrollmentToken}
                  serverUrl={enrollment.serverUrl}
                />
              )
            : null}
        </DialogContent>
        {enrollment
          ? (
              <footer className="flex items-center justify-end gap-2 border-t border-t-border-low p-3">
                <Button type="button" onClick={closeDialog}>
                  {__("Close")}
                </Button>
              </footer>
            )
          : null}
      </Dialog>
    </>
  );
}
