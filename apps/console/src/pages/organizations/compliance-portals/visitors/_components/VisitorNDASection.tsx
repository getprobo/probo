// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { DownloadSimpleIcon, FileTextIcon, ShieldSlashIcon, TrashIcon } from "@phosphor-icons/react";
import { safeOpenUrl } from "@probo/helpers";
import { dateFormat } from "@probo/i18n";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Callout } from "@probo/ui/src/v2/Callout/Callout";
import { Dialog } from "@probo/ui/src/v2/Dialog/Dialog";
import { DialogClose } from "@probo/ui/src/v2/Dialog/DialogClose";
import { DialogDescription } from "@probo/ui/src/v2/Dialog/DialogDescription";
import { DialogFooter } from "@probo/ui/src/v2/Dialog/DialogFooter";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import { DialogTrigger } from "@probo/ui/src/v2/Dialog/DialogTrigger";
import { DateField } from "@probo/ui/src/v2/form/DateField";
import { Field } from "@probo/ui/src/v2/form/Field";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type ReactElement, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { VisitorNDASection_access$key } from "#/__generated__/core/VisitorNDASection_access.graphql";
import type { VisitorNDASectionDeleteMutation } from "#/__generated__/core/VisitorNDASectionDeleteMutation.graphql";
import type { VisitorNDASectionSetBypassMutation } from "#/__generated__/core/VisitorNDASectionSetBypassMutation.graphql";
import type { VisitorNDASectionUploadMutation } from "#/__generated__/core/VisitorNDASectionUploadMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { visitorNda } from "../variants";

const RECORDED_NDA_MAX_BYTES = 10 * 1024 * 1024;

const fragment = graphql`
  fragment VisitorNDASection_access on CompliancePortalAccess {
    id
    ndaBypassed
    recordedNdaSignedAt
    canUpdate: permission(action: "compliance-portal:portal-access:update")
    recordedNda {
      fileName
      downloadUrl
    }
  }
`;

const setBypassMutation = graphql`
  mutation VisitorNDASectionSetBypassMutation($input: SetCompliancePortalAccessNDABypassInput!) {
    setCompliancePortalAccessNDABypass(input: $input) {
      compliancePortalAccess {
        id
        ndaBypassed
      }
    }
  }
`;

const uploadMutation = graphql`
  mutation VisitorNDASectionUploadMutation($input: UploadCompliancePortalAccessRecordedNDAInput!) {
    uploadCompliancePortalAccessRecordedNDA(input: $input) {
      compliancePortalAccess {
        id
        recordedNdaSignedAt
        recordedNda {
          fileName
          downloadUrl
        }
      }
    }
  }
`;

const deleteMutation = graphql`
  mutation VisitorNDASectionDeleteMutation($input: DeleteCompliancePortalAccessRecordedNDAInput!) {
    deleteCompliancePortalAccessRecordedNDA(input: $input) {
      compliancePortalAccess {
        id
        recordedNdaSignedAt
        recordedNda {
          fileName
          downloadUrl
        }
      }
    }
  }
`;

type RecordedFileError = "invalidFileType" | "fileTooLarge" | "signedAt";

function todayISO(): string {
  const now = new Date();
  const month = String(now.getMonth() + 1).padStart(2, "0");
  const day = String(now.getDate()).padStart(2, "0");
  return `${now.getFullYear()}-${month}-${day}`;
}

interface VisitorNDASectionProps {
  accessKey: VisitorNDASection_access$key;
}

export function VisitorNDASection({ accessKey }: VisitorNDASectionProps) {
  const { i18n, t } = useTranslation("organizations/compliance-portals");
  const access = useFragment(fragment, accessKey);
  const { root, intro, actions, record, recordRow, copy, date } = visitorNda();
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [signedOn, setSignedOn] = useState("");
  const [fileError, setFileError] = useState<RecordedFileError | null>(null);
  const [setBypass, isSettingBypass] = useMutation<VisitorNDASectionSetBypassMutation>(
    setBypassMutation,
  );
  const [uploadRecorded, isUploading] = useMutation<VisitorNDASectionUploadMutation>(
    uploadMutation,
    {
      successMessage: t("visitorNda.messages.uploaded"),
      errorToast: t("visitorNda.errors.upload"),
    },
  );

  if (!access.canUpdate && !access.ndaBypassed && access.recordedNda == null) {
    return null;
  }

  function handleFile(file: File) {
    setFileError(null);

    const isPdf = file.type === "application/pdf" || file.name.toLowerCase().endsWith(".pdf");
    if (!isPdf) {
      setFileError("invalidFileType");
      return;
    }

    if (file.size > RECORDED_NDA_MAX_BYTES) {
      setFileError("fileTooLarge");
      return;
    }

    if (signedOn === "") {
      setFileError("signedAt");
      return;
    }

    void uploadRecorded({
      variables: {
        input: {
          compliancePortalAccessId: access.id,
          fileName: file.name,
          file: null,
          signedAt: `${signedOn}T12:00:00.000Z`,
        },
      },
      uploadables: {
        "input.file": file,
      },
    }).catch(() => {
      // Error toast is already shown by useMutation.
    });
  }

  return (
    <section className={root()}>
      <div className={intro()}>
        <Heading level={2} size={4} weight="medium" highContrast>
          {t("visitorNda.title")}
        </Heading>
        <Text size={2} color="faint">
          {t("visitorNda.description")}
        </Text>
      </div>

      {access.ndaBypassed && (
        <Callout color="sky" icon={<ShieldSlashIcon weight="fill" />}>
          {t("visitorNda.bypassed")}
        </Callout>
      )}

      {access.canUpdate && (
        <div className={actions()}>
          {access.ndaBypassed
            ? (
                <ConfirmNDADialog
                  title={t("visitorNda.requireConfirm.title")}
                  description={t("visitorNda.requireConfirm.description")}
                  confirmLabel={t("visitorNda.requireConfirm.confirm")}
                  cancelLabel={t("visitorNda.requireConfirm.cancel")}
                  loading={isSettingBypass}
                  onConfirm={() => setBypass({
                    variables: { input: { id: access.id, bypassed: false } },
                  }, {
                    successMessage: t("visitorNda.messages.required"),
                    errorToast: t("visitorNda.errors.bypass"),
                  })}
                >
                  <Button variant="soft" color="neutral" loading={isSettingBypass}>
                    {t("visitorNda.actions.require")}
                  </Button>
                </ConfirmNDADialog>
              )
            : (
                <ConfirmNDADialog
                  title={t("visitorNda.bypassConfirm.title")}
                  description={t("visitorNda.bypassConfirm.description")}
                  confirmLabel={t("visitorNda.bypassConfirm.confirm")}
                  cancelLabel={t("visitorNda.bypassConfirm.cancel")}
                  confirmColor="amber"
                  loading={isSettingBypass}
                  onConfirm={() => setBypass({
                    variables: { input: { id: access.id, bypassed: true } },
                  }, {
                    successMessage: t("visitorNda.messages.bypassed"),
                    errorToast: t("visitorNda.errors.bypass"),
                  })}
                >
                  <Button variant="soft" color="amber" iconStart={<ShieldSlashIcon />} loading={isSettingBypass}>
                    {t("visitorNda.actions.bypass")}
                  </Button>
                </ConfirmNDADialog>
              )}
        </div>
      )}

      <div className={record()}>
        {access.recordedNda != null
          ? (
              <div className={recordRow()}>
                <div className={copy()}>
                  <Text size={2} weight="medium" color="neutral" highContrast>
                    {access.recordedNda.fileName}
                  </Text>
                  {access.recordedNdaSignedAt != null && (
                    <Text size={1} color="faint">
                      {t("visitorNda.signedAt", {
                        date: dateFormat(i18n.language, access.recordedNdaSignedAt),
                      })}
                    </Text>
                  )}
                </div>
                <div className={actions()}>
                  <IconButton
                    variant="ghost"
                    color="neutral"
                    aria-label={t("visitorNda.actions.download")}
                    onClick={() => {
                      const url = access.recordedNda?.downloadUrl;
                      if (url != null) {
                        safeOpenUrl(url);
                      }
                    }}
                  >
                    <DownloadSimpleIcon />
                  </IconButton>
                  {access.canUpdate && (
                    <DeleteRecordedNDADialog accessId={access.id}>
                      <IconButton
                        variant="ghost"
                        color="red"
                        aria-label={t("visitorNda.actions.delete")}
                      >
                        <TrashIcon />
                      </IconButton>
                    </DeleteRecordedNDADialog>
                  )}
                </div>
              </div>
            )
          : (
              <div className={copy()}>
                <Text size={2} weight="medium" color="neutral" highContrast>
                  {t("visitorNda.empty")}
                </Text>
                <Text size={1} color="faint">
                  {t("visitorNda.emptyDescription")}
                </Text>
              </div>
            )}

        {access.canUpdate && (
          <>
            <Field
              className={date()}
              label={t("visitorNda.fields.signedAt")}
              error={fileError === "signedAt" ? t("visitorNda.errors.signedAt") : undefined}
            >
              <DateField
                name="recordedNdaSignedAt"
                size={2}
                value={signedOn}
                locale={i18n.language}
                max={todayISO()}
                required
                onValueChange={(next) => {
                  setSignedOn(next);
                  setFileError(null);
                }}
              />
            </Field>
            {fileError != null && fileError !== "signedAt" && (
              <Text size={1} color="red">
                {t(`visitorNda.errors.${fileError}`)}
              </Text>
            )}
            <div className={actions()}>
              <input
                ref={fileInputRef}
                type="file"
                accept="application/pdf,.pdf"
                hidden
                onChange={(event) => {
                  const file = event.target.files?.[0];
                  event.target.value = "";
                  if (file != null) {
                    handleFile(file);
                  }
                }}
              />
              <Button
                variant="soft"
                color="neutral"
                iconStart={<FileTextIcon />}
                loading={isUploading}
                onClick={() => fileInputRef.current?.click()}
              >
                {access.recordedNda != null
                  ? t("visitorNda.actions.replace")
                  : t("visitorNda.actions.upload")}
              </Button>
            </div>
          </>
        )}
      </div>
    </section>
  );
}

interface ConfirmNDADialogProps {
  title: string;
  description: string;
  confirmLabel: string;
  cancelLabel: string;
  confirmColor?: "amber" | "neutral";
  loading: boolean;
  onConfirm: () => Promise<unknown>;
  children: ReactElement;
}

function ConfirmNDADialog({
  title,
  description,
  confirmLabel,
  cancelLabel,
  confirmColor = "neutral",
  loading,
  onConfirm,
  children,
}: ConfirmNDADialogProps) {
  const [open, setOpen] = useState(false);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger render={children} />
      <DialogPopup>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <DialogClose
            render={(
              <Button variant="soft" color="neutral">
                {cancelLabel}
              </Button>
            )}
          />
          <Button
            type="button"
            variant="solid"
            color={confirmColor}
            loading={loading}
            onClick={() => {
              void onConfirm().then(
                () => {
                  setOpen(false);
                },
                () => {
                  // Error toast is already shown by useMutation.
                },
              );
            }}
          >
            {confirmLabel}
          </Button>
        </DialogFooter>
      </DialogPopup>
    </Dialog>
  );
}

function DeleteRecordedNDADialog({
  accessId,
  children,
}: {
  accessId: string;
  children: ReactElement;
}) {
  const { t } = useTranslation("organizations/compliance-portals");
  const [open, setOpen] = useState(false);
  const [deleteRecorded, isDeleting] = useMutation<VisitorNDASectionDeleteMutation>(
    deleteMutation,
    {
      successMessage: t("visitorNda.messages.deleted"),
      errorToast: t("visitorNda.errors.delete"),
    },
  );

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger render={children} />
      <DialogPopup>
        <DialogHeader>
          <DialogTitle>{t("visitorNda.deleteConfirm.title")}</DialogTitle>
          <DialogDescription>
            {t("visitorNda.deleteConfirm.description")}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <DialogClose
            render={(
              <Button variant="soft" color="neutral">
                {t("visitorNda.deleteConfirm.cancel")}
              </Button>
            )}
          />
          <Button
            type="button"
            variant="solid"
            color="red"
            iconStart={<TrashIcon />}
            loading={isDeleting}
            onClick={() => {
              void deleteRecorded({
                variables: { input: { compliancePortalAccessId: accessId } },
              }).then(
                () => {
                  setOpen(false);
                },
                () => {
                  // Error toast is already shown by useMutation.
                },
              );
            }}
          >
            {t("visitorNda.deleteConfirm.confirm")}
          </Button>
        </DialogFooter>
      </DialogPopup>
    </Dialog>
  );
}
