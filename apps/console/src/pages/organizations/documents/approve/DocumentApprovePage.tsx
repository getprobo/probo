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
import { usePageTitle } from "@probo/hooks";
import { dateFormat } from "@probo/i18n";
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
import { useTranslation } from "react-i18next";
import {
  type PreloadedQuery,
  useFragment,
  useMutation,
  usePreloadedQuery,
} from "react-relay";
import { Navigate, useNavigate } from "react-router";
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
    consentText
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
  const organizationId = useOrganizationId();
  const data = usePreloadedQuery<DocumentApprovePageQuery>(
    documentApprovePageQuery,
    queryRef,
  );

  const document = data.viewer.approvableDocument;
  if (!document) {
    return (
      <Navigate
        to={`/organizations/${organizationId}/employee/approvals`}
        replace
      />
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
  const { t, i18n } = useTranslation();
  const versionData = useFragment(versionRowFragment, fKey);
  const approvalDecision = versionData.approvalDecision;
  const state = approvalDecision?.state;
  const isApproved = state === "APPROVED";
  const isRejected = state === "REJECTED";
  const isVoided = state === "VOIDED";

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
            : isVoided
              ? <IconRadioUnchecked size={20} className="text-txt-secondary" />
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
            ? t("documentApprovePage.versionWithDate", {
                major: versionData.major,
                minor: versionData.minor,
                date: dateFormat(i18n.language, versionData.publishedAt),
              })
            : t("documentApprovePage.version", {
                major: versionData.major,
                minor: versionData.minor,
              })}
        </p>
      </div>
      <div className="flex-shrink-0">
        {isApproved
          ? <Badge variant="success">{t("documentApprovePage.status.approved")}</Badge>
          : isRejected
            ? <Badge variant="danger">{t("documentApprovePage.status.rejected")}</Badge>
            : isVoided
              ? <Badge variant="neutral">{t("documentApprovePage.status.voided")}</Badge>
              : isSelected
                ? <Badge variant="info">{t("documentApprovePage.status.inReview")}</Badge>
                : <Badge variant="warning">{t("documentApprovePage.status.pending")}</Badge>}
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
  const { t } = useTranslation();
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
  const isVoided = decision.state === "VOIDED";

  if (isVoided) {
    return (
      <>
        <div className="flex items-center gap-2 text-sm text-txt-secondary mb-4">
          <span>{t("documentApprovePage.messages.noLongerRequired")}</span>
        </div>
        <Button onClick={onBack} className="h-10 w-full" variant="secondary">
          {t("documentApprovePage.actions.back")}
        </Button>
      </>
    );
  }

  if (!decision.canApprove && !decision.canReject) {
    return (
      <Button onClick={onBack} className="h-10 w-full" variant="secondary">
        {t("documentApprovePage.actions.back")}
      </Button>
    );
  }

  if (isApproved) {
    return (
      <>
        <div className="flex items-center gap-2 text-sm text-txt-accent mb-4">
          <IconCircleCheck size={20} />
          <span>{t("documentApprovePage.messages.approved")}</span>
        </div>
        <Button onClick={onBack} className="h-10 w-full" variant="secondary">
          {t("documentApprovePage.actions.back")}
        </Button>
      </>
    );
  }

  if (isRejected) {
    return (
      <>
        <div className="flex items-center gap-2 text-sm text-txt-danger mb-4">
          <IconCircleX size={20} />
          <span>{t("documentApprovePage.messages.rejected")}</span>
        </div>
        <Button onClick={onBack} className="h-10 w-full" variant="secondary">
          {t("documentApprovePage.actions.back")}
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
              {t("documentApprovePage.actions.reject")}
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
                        title: t("documentApprovePage.errors.title"),
                        description: formatError(t("documentApprovePage.errors.approve"), errors),
                        variant: "error",
                      });
                    } else {
                      toast({
                        title: t("documentApprovePage.messages.successTitle"),
                        description: t("documentApprovePage.messages.approvalSuccess"),
                        variant: "success",
                      });
                    }
                  },
                  onError(error) {
                    toast({
                      title: t("documentApprovePage.errors.title"),
                      description: error.message,
                      variant: "error",
                    });
                  },
                });
              }}
            >
              {t("documentApprovePage.actions.approve")}
            </Button>
          )}
        </div>
        <p className="text-xs text-txt-tertiary">
          {decision.consentText}
        </p>
        <Button onClick={onBack} className="w-full" variant="secondary">
          {t("documentApprovePage.actions.back")}
        </Button>
      </div>

      <Dialog ref={rejectDialogRef} title={t("documentApprovePage.rejectDialog.title")}>
        <DialogContent padded>
          <p className="text-sm text-txt-secondary mb-4">
            {t("documentApprovePage.rejectDialog.description")}
          </p>
          <Textarea
            placeholder={t("documentApprovePage.rejectDialog.placeholder")}
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
                      title: t("documentApprovePage.errors.title"),
                      description: formatError(t("documentApprovePage.errors.reject"), errors),
                      variant: "error",
                    });
                  } else {
                    toast({
                      title: t("documentApprovePage.messages.successTitle"),
                      description: t("documentApprovePage.messages.rejectionSuccess"),
                      variant: "success",
                    });
                    rejectDialogRef.current?.close();
                  }
                },
                onError(error) {
                  toast({
                    title: t("documentApprovePage.errors.title"),
                    description: error.message,
                    variant: "error",
                  });
                },
              });
            }}
          >
            {t("documentApprovePage.rejectDialog.submit")}
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
  const { t } = useTranslation();
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

  usePageTitle(t("documentApprovePage.pageTitle"));

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
            title: t("documentApprovePage.errors.title"),
            description: formatError(
              t("documentApprovePage.errors.loadPdf"),
              errors,
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
          title: t("documentApprovePage.errors.title"),
          description: formatError(
            t("documentApprovePage.errors.loadPdf"),
            error,
          ),
          variant: "error",
        });
      },
    });

    return () => {
      pdfUrlRef.current = null;
    };
  }, [selectedVersion?.id, exportPDF, toast, t]);

  if (versions.length === 0) {
    return (
      <Navigate
        to={`/organizations/${organizationId}/employee/approvals`}
        replace
      />
    );
  }

  return (
    <div className="fixed inset-0 top-12 bg-level-2 flex flex-col">
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
            {t("documentApprovePage.reviewHint")}
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
