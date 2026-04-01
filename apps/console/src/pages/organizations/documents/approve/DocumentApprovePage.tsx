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

import { formatError, type GraphQLError } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  Badge,
  Button,
  Card,
  Dialog,
  DialogContent,
  DialogFooter,
  IconCircleCheck,
  IconCircleX,
  IconRadioUnchecked,
  Spinner,
  Textarea,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { clsx } from "clsx";
import { useEffect, useRef, useState } from "react";
import {
  type PreloadedQuery,
  useFragment,
  useMutation,
  usePreloadedQuery,
} from "react-relay";
import { useNavigate } from "react-router";
import { graphql } from "relay-runtime";
import { useWindowSize } from "usehooks-ts";

import type { DocumentApprovePage_approveMutation } from "#/__generated__/core/DocumentApprovePage_approveMutation.graphql";
import type { DocumentApprovePage_rejectMutation } from "#/__generated__/core/DocumentApprovePage_rejectMutation.graphql";
import type { DocumentApprovePageDecisionFragment$key } from "#/__generated__/core/DocumentApprovePageDecisionFragment.graphql";
import type { DocumentApprovePageDocumentFragment$key } from "#/__generated__/core/DocumentApprovePageDocumentFragment.graphql";
import type { DocumentApprovePageExportEmployeePDFMutation } from "#/__generated__/core/DocumentApprovePageExportEmployeePDFMutation.graphql";
import type { DocumentApprovePageQuery } from "#/__generated__/core/DocumentApprovePageQuery.graphql";
import type { DocumentApprovePageVersionRowFragment$key } from "#/__generated__/core/DocumentApprovePageVersionRowFragment.graphql";
import { PDFPreview } from "#/components/documents/PDFPreview";
import { useOrganizationId } from "#/hooks/useOrganizationId";

export const documentApprovePageQuery = graphql`
  query DocumentApprovePageQuery($documentId: ID!) {
    viewer @required(action: THROW) {
      approvableDocument(id: $documentId) {
        ...DocumentApprovePageDocumentFragment
      }
    }
  }
`;

const documentFragment = graphql`
  fragment DocumentApprovePageDocumentFragment on EmployeeDocument {
    id
    title
    versions(first: 100, orderBy: { field: CREATED_AT, direction: DESC })
      @required(action: THROW) {
      edges @required(action: THROW) {
        node @required(action: THROW) {
          id
          ...DocumentApprovePageVersionRowFragment
          approvalDecision {
            ...DocumentApprovePageDecisionFragment
          }
        }
      }
    }
  }
`;

const versionRowFragment = graphql`
  fragment DocumentApprovePageVersionRowFragment on EmployeeDocumentVersion {
    id
    major
    minor
    publishedAt
    approvalDecision {
      id
      state
    }
  }
`;

const decisionFragment = graphql`
  fragment DocumentApprovePageDecisionFragment on DocumentVersionApprovalDecision {
    id
    state
    canApprove: permission(action: "core:document-version:approve")
    canReject: permission(action: "core:document-version:reject")
  }
`;

const approveDocumentVersionMutation = graphql`
  mutation DocumentApprovePage_approveMutation(
    $input: ApproveDocumentVersionInput!
  ) {
    approveDocumentVersion(input: $input) {
      approvalDecision {
        ...DocumentApprovePageDecisionFragment
      }
    }
  }
`;

const rejectDocumentVersionMutation = graphql`
  mutation DocumentApprovePage_rejectMutation(
    $input: RejectDocumentVersionInput!
  ) {
    rejectDocumentVersion(input: $input) {
      approvalDecision {
        ...DocumentApprovePageDecisionFragment
      }
    }
  }
`;

const exportPDFMutation = graphql`
  mutation DocumentApprovePageExportEmployeePDFMutation(
    $input: ExportEmployeeDocumentVersionPDFInput!
  ) {
    exportEmployeeDocumentVersionPDF(input: $input) {
      data
    }
  }
`;

export function DocumentApprovePage(props: {
  queryRef: PreloadedQuery<DocumentApprovePageQuery>;
}) {
  const { queryRef } = props;
  const data = usePreloadedQuery<DocumentApprovePageQuery>(
    documentApprovePageQuery,
    queryRef,
  );

  const document = data.viewer.approvableDocument;
  if (!document) {
    return (
      <div className="flex items-center justify-center h-full">
        <Spinner />
      </div>
    );
  }

  return <DocumentApproveContent fKey={document} />;
}

function VersionRow({
  fKey,
  isSelected,
  onSelect,
}: {
  fKey: DocumentApprovePageVersionRowFragment$key;
  isSelected: boolean;
  onSelect: () => void;
}) {
  const { __ } = useTranslate();
  const versionData = useFragment(versionRowFragment, fKey);
  const approvalDecision = versionData.approvalDecision;
  const state = approvalDecision?.state;
  const isApproved = state === "APPROVED";
  const isRejected = state === "REJECTED";

  return (
    <div
      onClick={onSelect}
      className={clsx(
        "flex items-center gap-3 py-3 px-4 transition-colors cursor-pointer",
        isSelected
          ? "bg-blue-50 border-l-4 border-blue-500"
          : "bg-transparent hover:bg-level-1",
      )}
    >
      <div className="flex items-center justify-center w-8 h-8 rounded-full bg-level-2 flex-shrink-0">
        {isApproved
          ? <IconCircleCheck size={20} className="text-txt-success" />
          : isRejected
            ? <IconCircleX size={20} className="text-txt-danger" />
            : <IconRadioUnchecked size={20} className="text-txt-tertiary" />}
      </div>
      <div className="flex-1 min-w-0">
        <p
          className={clsx(
            "text-sm font-medium truncate",
            (isApproved || isRejected) ? "text-txt-tertiary" : "text-txt-primary",
          )}
        >
          {versionData.publishedAt
            ? `v${versionData.major}.${versionData.minor} - ${(() => {
              const date = new Date(versionData.publishedAt);
              const day = String(date.getDate()).padStart(2, "0");
              const month = String(date.getMonth() + 1).padStart(2, "0");
              const year = date.getFullYear();
              return `${day}/${month}/${year}`;
            })()}`
            : `v${versionData.major}.${versionData.minor}`}
        </p>
      </div>
      <div className="flex-shrink-0">
        {isApproved
          ? <Badge variant="success">{__("Approved")}</Badge>
          : isRejected
            ? <Badge variant="danger">{__("Rejected")}</Badge>
            : isSelected
              ? <Badge variant="info">{__("In review")}</Badge>
              : <Badge variant="warning">{__("Pending")}</Badge>}
      </div>
    </div>
  );
}

function ViewerDecision(props: {
  fragmentRef: DocumentApprovePageDecisionFragment$key;
  versionId: string;
  onBack: () => void;
}) {
  const { fragmentRef, versionId, onBack } = props;
  const { __ } = useTranslate();
  const decision = useFragment(decisionFragment, fragmentRef);
  const rejectDialogRef = useDialogRef();
  const [rejectComment, setRejectComment] = useState("");
  const { toast } = useToast();

  const [approveVersion, isApproving] = useMutation<DocumentApprovePage_approveMutation>(
    approveDocumentVersionMutation,
  );

  const [rejectVersion, isRejecting] = useMutation<DocumentApprovePage_rejectMutation>(
    rejectDocumentVersionMutation,
  );

  const isPending = decision.state === "PENDING";
  const isApproved = decision.state === "APPROVED";
  const isRejected = decision.state === "REJECTED";

  if (!decision.canApprove && !decision.canReject) {
    return (
      <Button onClick={onBack} className="h-10 w-full" variant="secondary">
        {__("Back to Documents")}
      </Button>
    );
  }

  if (isApproved) {
    return (
      <>
        <div className="flex items-center gap-2 text-sm text-txt-accent mb-4">
          <IconCircleCheck size={20} />
          <span>{__("You have approved this document.")}</span>
        </div>
        <Button onClick={onBack} className="h-10 w-full" variant="secondary">
          {__("Back to Documents")}
        </Button>
      </>
    );
  }

  if (isRejected) {
    return (
      <>
        <div className="flex items-center gap-2 text-sm text-txt-danger mb-4">
          <IconCircleX size={20} />
          <span>{__("You have rejected this document.")}</span>
        </div>
        <Button onClick={onBack} className="h-10 w-full" variant="secondary">
          {__("Back to Documents")}
        </Button>
      </>
    );
  }

  if (!isPending) {
    return null;
  }

  return (
    <>
      <div className="space-y-3">
        <div className="flex gap-3">
          {decision.canReject && (
            <Button
              variant="danger"
              className="flex-1"
              disabled={isApproving || isRejecting}
              onClick={() => rejectDialogRef.current?.open()}
            >
              {__("Reject")}
            </Button>
          )}
          {decision.canApprove && (
            <Button
              className="flex-1"
              disabled={isApproving || isRejecting}
              icon={isApproving ? Spinner : undefined}
              onClick={() => {
                approveVersion({
                  variables: {
                    input: {
                      documentVersionId: versionId,
                    },
                  },
                  onCompleted(_, errors) {
                    if (errors?.length) {
                      toast({
                        title: __("Error"),
                        description: formatError(__("Failed to approve document"), errors),
                        variant: "error",
                      });
                    } else {
                      toast({
                        title: __("Success"),
                        description: __("Document approved successfully"),
                        variant: "success",
                      });
                    }
                  },
                  onError(error) {
                    toast({
                      title: __("Error"),
                      description: error.message,
                      variant: "error",
                    });
                  },
                });
              }}
            >
              {__("Approve")}
            </Button>
          )}
        </div>
        <p className="text-xs text-txt-tertiary">
          {__("By clicking Approve, I consent to approve this document electronically and agree that my electronic signature has the same legal validity as a handwritten signature.")}
        </p>
        <Button onClick={onBack} className="w-full" variant="secondary">
          {__("Back to Documents")}
        </Button>
      </div>

      <Dialog ref={rejectDialogRef} title={__("Reject Document")}>
        <DialogContent padded>
          <p className="text-sm text-txt-secondary mb-4">
            {__("Please provide a reason for rejecting this document. The document will be sent back to draft status.")}
          </p>
          <Textarea
            placeholder={__("Reason for rejection...")}
            value={rejectComment}
            onChange={e => setRejectComment(e.target.value)}
            rows={4}
          />
        </DialogContent>
        <DialogFooter>
          <Button
            variant="danger"
            disabled={isRejecting}
            icon={isRejecting ? Spinner : undefined}
            onClick={() => {
              rejectVersion({
                variables: {
                  input: {
                    documentVersionId: versionId,
                    comment: rejectComment || undefined,
                  },
                },
                onCompleted(_, errors) {
                  if (errors?.length) {
                    toast({
                      title: __("Error"),
                      description: formatError(__("Failed to reject document"), errors),
                      variant: "error",
                    });
                  } else {
                    toast({
                      title: __("Success"),
                      description: __("Document rejected"),
                      variant: "success",
                    });
                    rejectDialogRef.current?.close();
                  }
                },
                onError(error) {
                  toast({
                    title: __("Error"),
                    description: error.message,
                    variant: "error",
                  });
                },
              });
            }}
          >
            {__("Reject Document")}
          </Button>
        </DialogFooter>
      </Dialog>
    </>
  );
}

function DocumentApproveContent({
  fKey,
}: {
  fKey: DocumentApprovePageDocumentFragment$key;
}) {
  const { __ } = useTranslate();
  const navigate = useNavigate();
  const { width } = useWindowSize();
  const isMobile = width < 1100;
  const isDesktop = !isMobile;
  const organizationId = useOrganizationId();
  const { toast } = useToast();

  const documentData = useFragment(documentFragment, fKey);
  const versions = documentData.versions.edges.map(({ node }) => node);

  const [selectedVersionId, setSelectedVersionId] = useState<
    string | undefined
  >(() => versions[0]?.id);

  const selectedVersion = versions.find(v => v?.id === selectedVersionId);

  usePageTitle(__("Review and Approve Document"));

  const [exportPDF] = useMutation<DocumentApprovePageExportEmployeePDFMutation>(
    exportPDFMutation,
  );

  const [pdfUrl, setPdfUrl] = useState<string | null>(null);
  const pdfUrlRef = useRef<string | null>(null);

  useEffect(() => {
    if (!selectedVersion?.id) return;

    exportPDF({
      variables: {
        input: {
          documentVersionId: selectedVersion.id,
        },
      },
      onCompleted: (data, errors): void => {
        if (errors) {
          toast({
            title: __("Error"),
            description: formatError(
              __("Failed to load PDF"),
              errors as GraphQLError[],
            ),
            variant: "error",
          });
          return;
        }
        if (data.exportEmployeeDocumentVersionPDF?.data) {
          const dataUrl = data.exportEmployeeDocumentVersionPDF.data;
          pdfUrlRef.current = dataUrl;
          setPdfUrl(dataUrl);
        }
      },
      onError: (error) => {
        toast({
          title: __("Error"),
          description: formatError(
            __("Failed to load PDF"),
            error as GraphQLError,
          ),
          variant: "error",
        });
      },
    });

    return () => {
      pdfUrlRef.current = null;
    };
  }, [selectedVersion?.id, exportPDF, toast, __]);

  return (
    <div
      className="fixed bg-level-2 flex flex-col"
      style={{ top: "3rem", left: 0, right: 0, bottom: 0 }}
    >
      <div className="grid lg:grid-cols-2 min-h-0 h-full">
        <div className="w-full lg:w-[440px] mx-auto py-20 overflow-y-auto scrollbar-hide">
          <h1 className="text-2xl font-semibold mb-6">
            {documentData.title || ""}
          </h1>

          <Card className="mb-6 overflow-hidden">
            <div className="divide-y divide-border-solid">
              {versions.map(version => (
                <VersionRow
                  key={version.id}
                  fKey={version}
                  isSelected={version.id === selectedVersionId}
                  onSelect={() => setSelectedVersionId(version.id)}
                />
              ))}
            </div>
          </Card>

          <p className="text-txt-secondary text-sm mb-6">
            {__("Please review the document carefully before making your decision.")}
          </p>

          <div className="min-h-[60px]">
            {(() => {
              const decision = selectedVersion?.approvalDecision;
              return decision
                ? (
                    <ViewerDecision
                      fragmentRef={decision}
                      versionId={selectedVersion.id}
                      onBack={() =>
                        void navigate(`/organizations/${organizationId}/employee/approvals`)}
                    />
                  )
                : null;
            })()}
          </div>
        </div>

        {isDesktop && (
          <div className="bg-subtle h-full border-l border-border-solid min-h-0">
            {pdfUrl && (
              <PDFPreview src={pdfUrl} name={documentData.title || ""} />
            )}
          </div>
        )}
      </div>
    </div>
  );
}
