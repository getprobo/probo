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
import {
  Button,
  Dialog,
  DialogContent,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { useState } from "react";
import { useTranslation } from "react-i18next";
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
  const { t } = useTranslation();
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
            title: t("common.error"),
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
          title: t("common.success"),
          description: t("deviceEnrollment.messages.created"),
          variant: "success",
        });
      },
      onError(error) {
        toast({
          title: t("common.error"),
          description: formatError(
            t("devices.errors.create"),
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
        {t("deviceEnrollment.manual.cannotEnroll")}
        {" "}
        <button
          type="button"
          onClick={handleManualEnroll}
          disabled={isCreating}
          className="text-txt-primary underline hover:no-underline disabled:opacity-60"
        >
          {t("deviceEnrollment.manual.tryCreating")}
        </button>
      </p>

      <Dialog
        ref={dialogRef}
        onClose={handleClose}
        closable={!(isCreating && !enrollment)}
        title={t("deviceEnrollment.manual.title")}
      >
        <DialogContent padded className="space-y-4">
          {isCreating && !enrollment
            ? <p>{t("deviceEnrollment.manual.creating")}</p>
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
                  {t("common.actions.close")}
                </Button>
              </footer>
            )
          : null}
      </Dialog>
    </>
  );
}
