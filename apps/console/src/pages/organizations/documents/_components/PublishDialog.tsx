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

import { formatError } from "@probo/helpers";
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  IconSend,
  IconUpload,
  Textarea,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { type Ref, useImperativeHandle, useMemo, useRef } from "react";
import { useTranslation } from "react-i18next";
import { useFragment, useMutation } from "react-relay";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { PublishDialog_documentFragment$key } from "#/__generated__/core/PublishDialog_documentFragment.graphql";
import type { PublishDialog_publishMutation } from "#/__generated__/core/PublishDialog_publishMutation.graphql";
import { PeopleMultiSelectField } from "#/components/form/PeopleMultiSelectField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useOrganizationId } from "#/hooks/useOrganizationId";

export type PublishDialogRef = {
  open: () => void;
};

type PublishDialogProps = {
  ref: Ref<PublishDialogRef>;
  documentId: string;
  documentFragmentRef: PublishDialog_documentFragment$key;
  onSuccess: () => void;
};

const documentFragment = graphql`
  fragment PublishDialog_documentFragment on Document {
    defaultApprovers {
      id
    }
  }
`;

const publishMutation = graphql`
  mutation PublishDialog_publishMutation($input: PublishDocumentInput!) {
    publishDocument(input: $input) {
      document {
        id
        status
      }
      documentVersion {
        id
        status
      }
      approvalQuorum {
        id
        status
        decisions(first: 0) {
          totalCount
        }
        approvedDecisions: decisions(first: 0 filter: { states: [APPROVED] }) {
          totalCount
        }
        documentVersion {
          id
          approvalQuorums(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
            edges {
              node {
                id
                status
                decisions(first: 0) {
                  totalCount
                }
                approvedDecisions: decisions(first: 0 filter: { states: [APPROVED] }) {
                  totalCount
                }
              }
            }
          }
        }
      }
    }
  }
`;

export function PublishDialog({
  ref,
  documentId,
  documentFragmentRef,
  onSuccess,
}: PublishDialogProps) {
  const document = useFragment(documentFragment, documentFragmentRef);
  const { t } = useTranslation();
  const { toast } = useToast();
  const organizationId = useOrganizationId();
  const dialogRef = useDialogRef();

  const publishSchema = useMemo(() => z.object({
    changelog: z.string().min(1, t("publishDialog.validation.changelogRequired")),
    approverIds: z.array(z.string()),
  }), [t]);

  const defaultApproverIds = document.defaultApprovers.map(a => a.id);

  const {
    control,
    handleSubmit,
    register,
    reset,
    watch,
    formState: { errors },
  } = useFormWithSchema(publishSchema, {
    defaultValues: {
      changelog: "",
      approverIds: [],
    },
  });

  useImperativeHandle(ref, () => ({
    open: () => {
      reset({
        changelog: "",
        approverIds: defaultApproverIds,
      });
      dialogRef.current?.open();
    },
  }));

  const [publish, isPublishing]
    = useMutation<PublishDialog_publishMutation>(publishMutation);

  const approverIds = watch("approverIds");
  const hasApprovers = approverIds.length > 0;
  const minorRef = useRef(false);

  const submit = (data: z.infer<typeof publishSchema>, minor: boolean) => {
    publish({
      variables: {
        input: {
          documentId,
          minor,
          approverIds: minor ? null : data.approverIds,
          changelog: data.changelog,
        },
      },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast({
            title: t("publishDialog.errors.title"),
            description: formatError(t("publishDialog.errors.publish"), errors),
            variant: "error",
          });
          return;
        }
        toast({
          title: t("publishDialog.messages.successTitle"),
          description: !minor && data.approverIds.length > 0
            ? t("publishDialog.messages.approvalRequested")
            : t("publishDialog.messages.published"),
          variant: "success",
        });
        dialogRef.current?.close();
        reset();
        onSuccess();
      },
      onError(error) {
        toast({ title: t("publishDialog.errors.title"), description: error.message, variant: "error" });
      },
    });
  };

  return (
    <Dialog className="max-w-xl" ref={dialogRef} title={t("publishDialog.title")}>
      <form
        onSubmit={e => void handleSubmit((data) => {
          const minor = minorRef.current;
          minorRef.current = false;
          submit(data, minor);
        })(e)}
      >
        <DialogContent padded>
          <div className="space-y-4">
            <div>
              <label htmlFor="changelog" className="text-sm font-medium text-txt-primary mb-1 block">
                {t("publishDialog.fields.changelog")}
              </label>
              <Textarea
                id="changelog"
                aria-label={t("publishDialog.fields.changelog")}
                required
                autogrow
                placeholder={t("publishDialog.fields.changelogPlaceholder")}
                {...register("changelog")}
              />
              {errors.changelog?.message && (
                <p className="text-xs text-txt-danger mt-1">{errors.changelog.message}</p>
              )}
            </div>
            <div>
              <p className="text-xs text-txt-secondary mb-3">
                {t("publishDialog.approversDescription")}
              </p>
              <PeopleMultiSelectField
                name="approverIds"
                label={t("publishDialog.fields.approvers")}
                control={control}
                organizationId={organizationId}
                placeholder={t("publishDialog.fields.approversPlaceholder")}
              />
            </div>
          </div>
        </DialogContent>
        <DialogFooter>
          <Button
            type="submit"
            variant="secondary"
            icon={IconUpload}
            onClick={() => { minorRef.current = true; }}
            disabled={isPublishing}
          >
            {t("publishDialog.actions.publishMinor")}
          </Button>
          <Button
            type="submit"
            icon={hasApprovers ? IconSend : IconUpload}
            onClick={() => { minorRef.current = false; }}
            disabled={isPublishing}
          >
            {hasApprovers
              ? t("publishDialog.actions.requestApproval")
              : t("publishDialog.actions.publishMajor")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
