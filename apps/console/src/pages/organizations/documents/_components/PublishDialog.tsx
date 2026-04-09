// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

import { formatError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
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
import { useFragment, useMutation } from "react-relay";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { PublishDialog_documentFragment$key } from "#/__generated__/core/PublishDialog_documentFragment.graphql";
import type { PublishDialog_publishMajorMutation } from "#/__generated__/core/PublishDialog_publishMajorMutation.graphql";
import type { PublishDialog_publishMinorMutation } from "#/__generated__/core/PublishDialog_publishMinorMutation.graphql";
import type { PublishDialog_requestApprovalMutation } from "#/__generated__/core/PublishDialog_requestApprovalMutation.graphql";
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

const publishMajorMutation = graphql`
  mutation PublishDialog_publishMajorMutation($input: PublishMajorDocumentVersionInput!) {
    publishMajorDocumentVersion(input: $input) {
      document {
        id
        status
      }
      documentVersion {
        id
        status
      }
    }
  }
`;

const publishMinorMutation = graphql`
  mutation PublishDialog_publishMinorMutation($input: PublishMinorDocumentVersionInput!) {
    publishMinorDocumentVersion(input: $input) {
      document {
        id
        status
      }
      documentVersion {
        id
        status
      }
    }
  }
`;

const requestApprovalMutation = graphql`
  mutation PublishDialog_requestApprovalMutation(
    $input: RequestDocumentVersionApprovalInput!
  ) {
    requestDocumentVersionApproval(input: $input) {
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
  const { __ } = useTranslate();
  const { toast } = useToast();
  const organizationId = useOrganizationId();
  const dialogRef = useDialogRef();

  const publishSchema = useMemo(() => z.object({
    changelog: z.string().min(1, __("Changelog is required")),
    approverIds: z.array(z.string()),
  }), [__]);

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

  const [publishMajor, isPublishingMajor]
    = useMutation<PublishDialog_publishMajorMutation>(publishMajorMutation);
  const [publishMinor, isPublishingMinor]
    = useMutation<PublishDialog_publishMinorMutation>(publishMinorMutation);
  const [requestApproval, isRequesting]
    = useMutation<PublishDialog_requestApprovalMutation>(requestApprovalMutation);

  const isBusy = isPublishingMajor || isPublishingMinor || isRequesting;
  const approverIds = watch("approverIds");
  const hasApprovers = approverIds.length > 0;
  const actionRef = useRef<"publish" | "publish-minor" | "request-approval">("publish");

  const onPublishCompleted = (_: unknown, errors: ReadonlyArray<{ message: string }> | null) => {
    if (errors?.length) {
      toast({
        title: __("Error"),
        description: formatError(__("Failed to publish document"), [...errors]),
        variant: "error",
      });
    } else {
      toast({
        title: __("Success"),
        description: __("Document published successfully."),
        variant: "success",
      });
      dialogRef.current?.close();
      onSuccess();
    }
  };

  const onPublishError = (error: Error) => {
    toast({ title: __("Error"), description: error.message, variant: "error" });
  };

  const handlePublishMajor = (data: z.infer<typeof publishSchema>) => {
    publishMajor({
      variables: { input: { documentId, changelog: data.changelog } },
      onCompleted: onPublishCompleted,
      onError: onPublishError,
    });
  };

  const handlePublishMinor = (data: z.infer<typeof publishSchema>) => {
    publishMinor({
      variables: { input: { documentId, changelog: data.changelog } },
      onCompleted: onPublishCompleted,
      onError: onPublishError,
    });
  };

  const onRequestApproval = (data: z.infer<typeof publishSchema>) => {
    requestApproval({
      variables: {
        input: {
          documentId,
          approverIds: data.approverIds,
          changelog: data.changelog,
        },
      },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast({
            title: __("Error"),
            description: formatError(__("Failed to request approval"), errors),
            variant: "error",
          });
        } else {
          toast({
            title: __("Success"),
            description: __("Approval requested successfully."),
            variant: "success",
          });
          dialogRef.current?.close();
          reset();
          onSuccess();
        }
      },
      onError(error) {
        toast({ title: __("Error"), description: error.message, variant: "error" });
      },
    });
  };

  return (
    <Dialog className="max-w-xl" ref={dialogRef} title={__("Publish document")}>
      <form
        onSubmit={e => void handleSubmit((data) => {
          const action = actionRef.current;
          actionRef.current = "publish";
          if (action === "publish-minor") {
            handlePublishMinor(data);
          } else if (action === "request-approval") {
            onRequestApproval(data);
          } else if (data.approverIds.length > 0) {
            onRequestApproval(data);
          } else {
            handlePublishMajor(data);
          }
        })(e)}
      >
        <DialogContent padded>
          <div className="space-y-4">
            <div>
              <label htmlFor="changelog" className="text-sm font-medium text-txt-primary mb-1 block">
                {__("Changelog")}
              </label>
              <Textarea
                id="changelog"
                aria-label={__("Changelog")}
                required
                autogrow
                placeholder={__("Describe what changed in this version...")}
                {...register("changelog")}
              />
              {errors.changelog?.message && (
                <p className="text-xs text-txt-danger mt-1">{errors.changelog.message}</p>
              )}
            </div>
            <div>
              <p className="text-xs text-txt-secondary mb-3">
                {__("Approvers will receive an email and the document will be published as a major version once all have approved. Remove all approvers to publish directly as major.")}
              </p>
              <PeopleMultiSelectField
                name="approverIds"
                label={__("Approvers")}
                control={control}
                organizationId={organizationId}
                placeholder={__("Add approvers...")}
              />
            </div>
          </div>
        </DialogContent>
        <DialogFooter>
          {hasApprovers
            ? (
                <>
                  <Button
                    type="submit"
                    variant="secondary"
                    icon={IconUpload}
                    onClick={() => { actionRef.current = "publish-minor"; }}
                    disabled={isBusy}
                  >
                    {__("Publish as minor")}
                  </Button>
                  <Button
                    type="submit"
                    icon={IconSend}
                    onClick={() => { actionRef.current = "request-approval"; }}
                    disabled={isBusy}
                  >
                    {__("Request approval")}
                  </Button>
                </>
              )
            : (
                <>
                  <Button
                    type="submit"
                    variant="secondary"
                    icon={IconUpload}
                    onClick={() => { actionRef.current = "publish-minor"; }}
                    disabled={isBusy}
                  >
                    {__("Publish as minor")}
                  </Button>
                  <Button
                    type="submit"
                    icon={IconUpload}
                    onClick={() => { actionRef.current = "publish"; }}
                    disabled={isBusy}
                  >
                    {__("Publish as major")}
                  </Button>
                </>
              )}
        </DialogFooter>
      </form>
    </Dialog>
  );
}
