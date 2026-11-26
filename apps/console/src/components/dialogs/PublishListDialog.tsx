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

import { PaperPlaneTiltIcon, UploadSimpleIcon } from "@phosphor-icons/react";
import { formatError } from "@probo/helpers";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Dialog } from "@probo/ui/src/v2/Dialog/Dialog";
import { DialogBody } from "@probo/ui/src/v2/Dialog/DialogBody";
import { DialogClose } from "@probo/ui/src/v2/Dialog/DialogClose";
import { DialogDescription } from "@probo/ui/src/v2/Dialog/DialogDescription";
import { DialogFooter } from "@probo/ui/src/v2/Dialog/DialogFooter";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import { DialogTrigger } from "@probo/ui/src/v2/Dialog/DialogTrigger";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import { type FormEvent, type ReactElement, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { PublishListDialog_approver$key } from "#/__generated__/core/PublishListDialog_approver.graphql";
import { type Approver, ApproversField } from "#/components/people/ApproversField";

import { publishListDialog } from "./variants";

const approverFragment = graphql`
  fragment PublishListDialog_approver on Profile @relay(plural: true) {
    id
    fullName
    emailAddress
    avatar {
      downloadUrl
    }
  }
`;

export type PublishListDialogApprovers = PublishListDialog_approver$key;

export type PublishListDialogInput = {
  organizationId: string;
  minor: boolean;
  approverIds?: string[];
};

type Props = {
  children: ReactElement;
  organizationId: string;
  defaultApprovers: PublishListDialogApprovers;
  title: string;
  description?: string;
  publishedMessage: string;
  publishError: string;
  isPublishing: boolean;
  onPublish: (input: PublishListDialogInput) => Promise<string | null | undefined>;
  onPublished?: (documentId: string) => void;
};

export function PublishListDialog({ children, ...formProps }: Props) {
  const [open, setOpen] = useState(false);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger render={children} />
      <DialogPopup>
        <PublishListForm
          {...formProps}
          onDone={() => setOpen(false)}
        />
      </DialogPopup>
    </Dialog>
  );
}

type FormProps = Omit<Props, "children"> & {
  onDone: () => void;
};

function PublishListForm({
  organizationId,
  defaultApprovers,
  title,
  description,
  publishedMessage,
  publishError,
  isPublishing,
  onPublish,
  onPublished,
  onDone,
}: FormProps) {
  const { t } = useTranslation();
  const toast = useToast();
  const defaults = useFragment(approverFragment, defaultApprovers);
  const [approvers, setApprovers] = useState<Approver[]>(
    () => defaults.map(({ id, fullName, emailAddress, avatar }) => ({
      id,
      fullName,
      emailAddress,
      avatarUrl: avatar?.downloadUrl,
    })),
  );
  const { form, body } = publishListDialog();
  const requestsApproval = approvers.length > 0;

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const submitter = (event.nativeEvent as SubmitEvent).submitter;
    const minor = submitter instanceof HTMLButtonElement && submitter.value === "minor";
    const requestedApproval = !minor && requestsApproval;

    try {
      const documentId = await onPublish({
        organizationId,
        minor,
        approverIds: requestedApproval ? approvers.map(approver => approver.id) : undefined,
      });
      if (!documentId) {
        return;
      }
      toast.add({
        title: t("publishListDialog.messages.success"),
        description: requestedApproval
          ? t("publishListDialog.messages.approvalRequested")
          : publishedMessage,
        type: "success",
      });
      onDone();
      onPublished?.(documentId);
    } catch (error) {
      toast.add({
        title: t("publishListDialog.messages.error"),
        description: formatError(publishError, error as Parameters<typeof formatError>[1]),
        type: "error",
      });
    }
  }

  return (
    <form className={form()} onSubmit={e => void handleSubmit(e)}>
      <DialogHeader>
        <DialogTitle>{title}</DialogTitle>
        <DialogDescription>{description ?? t("publishListDialog.description")}</DialogDescription>
      </DialogHeader>
      <DialogBody>
        <div className={body()}>
          <ApproversField
            approvers={approvers}
            onChange={setApprovers}
            label={t("publishListDialog.fields.approvers")}
            description={t("publishListDialog.fields.approversDescription")}
            disabled={isPublishing}
          />
        </div>
      </DialogBody>
      <DialogFooter>
        <DialogClose
          render={(
            <Button variant="soft" color="neutral" disabled={isPublishing}>
              {t("publishListDialog.actions.cancel")}
            </Button>
          )}
        />
        <Button
          type="submit"
          value="minor"
          variant="soft"
          color="neutral"
          disabled={isPublishing}
          iconStart={<UploadSimpleIcon />}
        >
          {t("publishListDialog.actions.publishMinor")}
        </Button>
        <Button
          type="submit"
          value="major"
          variant="solid"
          color="neutral"
          highContrast
          loading={isPublishing}
          iconStart={requestsApproval ? <PaperPlaneTiltIcon /> : <UploadSimpleIcon />}
        >
          {requestsApproval
            ? t("publishListDialog.actions.requestApproval")
            : t("publishListDialog.actions.publish")}
        </Button>
      </DialogFooter>
    </form>
  );
}
