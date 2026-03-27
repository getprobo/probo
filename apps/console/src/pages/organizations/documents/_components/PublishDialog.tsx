import { formatError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  IconSend,
  IconUpload,
  IconWarning,
  Textarea,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { type Ref, useImperativeHandle, useRef } from "react";
import { useFragment, useMutation } from "react-relay";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { PublishDialog_documentFragment$key } from "#/__generated__/core/PublishDialog_documentFragment.graphql";
import type { PublishDialog_publishMutation } from "#/__generated__/core/PublishDialog_publishMutation.graphql";
import type { PublishDialog_requestApprovalMutation } from "#/__generated__/core/PublishDialog_requestApprovalMutation.graphql";
import { PeopleMultiSelectField } from "#/components/form/PeopleMultiSelectField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useOrganizationId } from "#/hooks/useOrganizationId";

export type PublishDialogRef = {
  open: () => void;
};

type Props = {
  ref: Ref<PublishDialogRef>;
  documentId: string;
  documentFragmentRef: PublishDialog_documentFragment$key;
  hasPendingApproval: boolean;
  onSuccess: () => void;
};

const documentFragment = graphql`
  fragment PublishDialog_documentFragment on Document {
    lastPublishedVersion: versions(first: 1, orderBy: { field: CREATED_AT, direction: DESC }, filter: { statuses: [PUBLISHED] }) {
      edges {
        node {
          approvalQuorums(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
            edges {
              node {
                decisions(first: 100) {
                  edges {
                    node {
                      approver {
                        id
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
`;

const publishMutation = graphql`
  mutation PublishDialog_publishMutation($input: PublishDocumentVersionInput!) {
    publishDocumentVersion(input: $input) {
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
  hasPendingApproval,
  onSuccess,
}: Props) {
  const document = useFragment(documentFragment, documentFragmentRef);
  const { __ } = useTranslate();
  const { toast } = useToast();
  const organizationId = useOrganizationId();
  const dialogRef = useDialogRef();

  const previousApproverIds = document.lastPublishedVersion.edges[0]
    ?.node.approvalQuorums.edges[0]
    ?.node.decisions?.edges.map(e => e.node.approver.id)
    ?? [];

  const schema = z.object({
    changelog: z.string().min(1, __("Changelog is required")),
    approverIds: z.array(z.string()),
  });

  const {
    control,
    handleSubmit,
    register,
    reset,
    watch,
    formState: { errors },
  } = useFormWithSchema(schema, {
    defaultValues: {
      changelog: "",
      approverIds: [],
    },
  });

  useImperativeHandle(ref, () => ({
    open: () => {
      reset({
        changelog: "",
        approverIds: previousApproverIds,
      });
      dialogRef.current?.open();
    },
  }));

  const [publishVersion, isPublishing] = useMutation<PublishDialog_publishMutation>(publishMutation);
  const [requestApproval, isRequesting] = useMutation<PublishDialog_requestApprovalMutation>(requestApprovalMutation);

  const isBusy = isPublishing || isRequesting;
  const approverIds = watch("approverIds");
  const actionRef = useRef<"publish" | "request-approval">("publish");

  const handlePublish = (data: z.infer<typeof schema>) => {
    publishVersion({
      variables: { input: { documentId, changelog: data.changelog } },
      onCompleted(_, errors) {
        if (errors?.length) {
          toast({
            title: __("Error"),
            description: formatError(__("Failed to publish document"), errors),
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
      },
      onError(error) {
        toast({ title: __("Error"), description: error.message, variant: "error" });
      },
    });
  };

  const onRequestApproval = (data: z.infer<typeof schema>) => {
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
          if (actionRef.current === "publish") {
            handlePublish(data);
          } else {
            onRequestApproval(data);
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
            {hasPendingApproval
              ? (
                  <div className="flex items-start gap-2 rounded-lg bg-bg-warning/10 border border-border-warning p-3">
                    <IconWarning size={16} className="text-txt-warning shrink-0 mt-0.5" />
                    <p className="text-sm text-txt-warning">
                      {__("An approval review is currently in progress. Publishing now will bypass the pending approval.")}
                    </p>
                  </div>
                )
              : (
                  <div>
                    <div className="text-sm font-medium text-txt-primary mb-1">
                      {__("Request approval before publishing")}
                    </div>
                    <p className="text-xs text-txt-secondary mb-3">
                      {__("Select approvers to review this document. The document will be published once all approvers have approved it. You can also publish directly without requiring approval.")}
                    </p>
                    <PeopleMultiSelectField
                      name="approverIds"
                      label={__("Approvers")}
                      control={control}
                      organizationId={organizationId}
                      placeholder={__("Add approvers...")}
                    />
                  </div>
                )}
          </div>
        </DialogContent>
        <DialogFooter>
          {hasPendingApproval
            ? (
                <Button
                  type="submit"
                  icon={IconUpload}
                  onClick={() => { actionRef.current = "publish"; }}
                  disabled={isBusy}
                >
                  {__("Publish now")}
                </Button>
              )
            : (
                <>
                  <Button
                    type="submit"
                    variant="secondary"
                    icon={IconUpload}
                    onClick={() => { actionRef.current = "publish"; }}
                    disabled={isBusy}
                  >
                    {__("Publish now")}
                  </Button>
                  <Button
                    type="submit"
                    icon={IconSend}
                    onClick={() => { actionRef.current = "request-approval"; }}
                    disabled={isBusy || approverIds.length === 0}
                  >
                    {__("Request approval")}
                  </Button>
                </>
              )}
        </DialogFooter>
      </form>
    </Dialog>
  );
}
