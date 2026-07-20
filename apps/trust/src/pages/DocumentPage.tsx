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
import { useSystemTheme } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { UnAuthenticatedError } from "@probo/relay";
import {
  Button,
  IconArrowDown,
  IconChevronLeft,
  IconLock,
  Spinner,
  useToast,
} from "@probo/ui";
import { useEffect, useState } from "react";
import {
  type PreloadedQuery,
  useMutation,
  usePreloadedQuery,
} from "react-relay";
import { Link, useLocation, useNavigate, useSearchParams } from "react-router";
import { graphql } from "relay-runtime";

import { PDFPreview } from "#/components/PDFPreview";

import type { DocumentPageExportDocumentMutation } from "./__generated__/DocumentPageExportDocumentMutation.graphql";
import type { DocumentPageExportReportMutation } from "./__generated__/DocumentPageExportReportMutation.graphql";
import type { DocumentPageExportCompliancePortalFileMutation } from "./__generated__/DocumentPageExportCompliancePortalFileMutation.graphql";
import type { DocumentPageQuery as DocumentPageQueryType } from "./__generated__/DocumentPageQuery.graphql";
import type { DocumentPageRequestDocumentAccessMutation } from "./__generated__/DocumentPageRequestDocumentAccessMutation.graphql";
import type { DocumentPageRequestReportAccessMutation } from "./__generated__/DocumentPageRequestReportAccessMutation.graphql";
import type { DocumentPageRequestCompliancePortalFileAccessMutation } from "./__generated__/DocumentPageRequestCompliancePortalFileAccessMutation.graphql";

export const documentPageQuery = graphql`
  query DocumentPageQuery($alias: String!) {
    currentCompliancePortal {
      logo {
        downloadUrl
      }
      darkLogo {
        downloadUrl
      }
    }
    aliasedNode(alias: $alias) @required(action: THROW) {
      __typename
      ... on Document {
        id
        title
        isUserAuthorized
        access {
          id
          status
        }
      }
      ... on CompliancePortalFile {
        id
        name
        isUserAuthorized
        access {
          id
          status
        }
      }
      ... on AuditReport {
        id
        fileName
        isUserAuthorized
        access {
          id
          status
        }
      }
    }
  }
`;

const exportDocumentMutation = graphql`
  mutation DocumentPageExportDocumentMutation(
    $input: ExportDocumentPDFInput!
  ) {
    exportDocumentPDF(input: $input) {
      data
    }
  }
`;

const exportCompliancePortalFileMutation = graphql`
  mutation DocumentPageExportCompliancePortalFileMutation(
    $input: ExportCompliancePortalFileInput!
  ) {
    exportCompliancePortalFile(input: $input) {
      data
    }
  }
`;

const exportReportMutation = graphql`
  mutation DocumentPageExportReportMutation(
    $input: ExportReportPDFInput!
  ) {
    exportReportPDF(input: $input) {
      data
    }
  }
`;

const requestDocumentAccessMutation = graphql`
  mutation DocumentPageRequestDocumentAccessMutation(
    $input: RequestDocumentAccessInput!
  ) {
    requestDocumentAccess(input: $input) {
      document {
        access {
          id
          status
        }
      }
    }
  }
`;

const requestCompliancePortalFileAccessMutation = graphql`
  mutation DocumentPageRequestCompliancePortalFileAccessMutation(
    $input: RequestCompliancePortalFileAccessInput!
  ) {
    requestCompliancePortalFileAccess(input: $input) {
      file {
        access {
          id
          status
        }
      }
    }
  }
`;

const requestReportAccessMutation = graphql`
  mutation DocumentPageRequestReportAccessMutation(
    $input: RequestReportAccessInput!
  ) {
    requestReportAccess(input: $input) {
      audit {
        reportFile {
          access {
            id
            status
          }
        }
      }
    }
  }
`;

type Props = {
  queryRef: PreloadedQuery<DocumentPageQueryType>;
};

function isPdfDataUri(dataUri: string): boolean {
  return dataUri.startsWith("data:application/pdf;");
}

function extractBase64Data(dataUri: string): string {
  const commaIndex = dataUri.indexOf(",");
  if (commaIndex === -1) return dataUri;
  return dataUri.substring(commaIndex + 1);
}

function extractMimeType(dataUri: string): string {
  const match = dataUri.match(/^data:([^;]+);/);
  return match?.[1] ?? "application/octet-stream";
}

function getNodeTitle(node: DocumentPageQueryType["response"]["aliasedNode"]): string | undefined {
  switch (node.__typename) {
    case "Document":
      return node.title;
    case "CompliancePortalFile":
      return node.name;
    case "AuditReport":
      return node.fileName;
    default:
      return undefined;
  }
}

function getNodeId(node: DocumentPageQueryType["response"]["aliasedNode"]): string | undefined {
  switch (node.__typename) {
    case "Document":
    case "CompliancePortalFile":
    case "AuditReport":
      return node.id;
    default:
      return undefined;
  }
}

export function DocumentPage({ queryRef }: Props) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const theme = useSystemTheme();
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const [pdfData, setPdfData] = useState<string | null>(null);
  const [fileData, setFileData] = useState<string | null>(null);
  const [exportError, setExportError] = useState<string | null>(null);

  const data = usePreloadedQuery<DocumentPageQueryType>(documentPageQuery, queryRef);
  const compliancePortal = data.currentCompliancePortal;
  const node = data.aliasedNode;

  if (
    node.__typename !== "Document"
    && node.__typename !== "CompliancePortalFile"
    && node.__typename !== "AuditReport"
  ) {
    throw new Error(`Unexpected node type: ${node.__typename}`);
  }

  const nodeTitle = getNodeTitle(node);
  const nodeId = getNodeId(node);

  const logoFileUrl = theme === "dark"
    ? (compliancePortal?.darkLogo?.downloadUrl ?? compliancePortal?.logo?.downloadUrl)
    : compliancePortal?.logo?.downloadUrl;

  const [exportDocument, isExportingDocument]
    = useMutation<DocumentPageExportDocumentMutation>(exportDocumentMutation);
  const [exportFile, isExportingFile]
    = useMutation<DocumentPageExportCompliancePortalFileMutation>(exportCompliancePortalFileMutation);
  const [exportReport, isExportingReport]
    = useMutation<DocumentPageExportReportMutation>(exportReportMutation);
  const [requestAccess, isRequestingAccess]
    = useMutation<DocumentPageRequestDocumentAccessMutation>(requestDocumentAccessMutation);
  const [requestFileAccess, isRequestingFileAccess]
    = useMutation<DocumentPageRequestCompliancePortalFileAccessMutation>(requestCompliancePortalFileAccessMutation);
  const [requestReportAccess, isRequestingReportAccess]
    = useMutation<DocumentPageRequestReportAccessMutation>(requestReportAccessMutation);

  const isExporting = isExportingDocument || isExportingFile || isExportingReport;

  const [prevNodeId, setPrevNodeId] = useState(nodeId);
  if (prevNodeId !== nodeId) {
    setPrevNodeId(nodeId);
    setPdfData(null);
    setFileData(null);
    setExportError(null);
  }

  useEffect(() => {
    if (!node.isUserAuthorized || pdfData || fileData || exportError) return;

    const onError = (error: Error) => {
      setExportError(error.message ?? __("Cannot export document"));
    };

    const onCompletedErrors = (errors: readonly { message: string }[] | null | undefined) => {
      if (errors?.length) {
        setExportError(formatError(__("Cannot export document"), [...errors]));
        return true;
      }
      return false;
    };

    switch (node.__typename) {
      case "Document":
        exportDocument({
          variables: { input: { documentId: node.id } },
          onCompleted: (response, errors) => {
            if (onCompletedErrors(errors)) return;
            setPdfData(response.exportDocumentPDF.data);
          },
          onError,
        });
        break;
      case "CompliancePortalFile":
        exportFile({
          variables: { input: { compliancePortalFileId: node.id } },
          onCompleted: (response, errors) => {
            if (onCompletedErrors(errors)) return;
            if (isPdfDataUri(response.exportCompliancePortalFile.data)) {
              setPdfData(response.exportCompliancePortalFile.data);
            } else {
              setFileData(response.exportCompliancePortalFile.data);
            }
          },
          onError,
        });
        break;
      case "AuditReport":
        exportReport({
          variables: { input: { reportId: node.id } },
          onCompleted: (response, errors) => {
            if (onCompletedErrors(errors)) return;
            setPdfData(response.exportReportPDF.data);
          },
          onError,
        });
        break;
    }
  }, [node, pdfData, fileData, exportError, exportDocument, exportFile, exportReport, __]);

  const handleRequestAccess = () => {
    const onError = (error: Error) => {
      if (error instanceof UnAuthenticatedError) {
        const urlSearchParams = new URLSearchParams([[
          "continue",
          window.location.origin + location.pathname + "?" + searchParams.toString(),
        ]]);
        void navigate(`/connect?${urlSearchParams.toString()}`);
        return;
      }
      toast({
        title: __("Error"),
        description: error.message ?? __("Cannot request access"),
        variant: "error",
      });
    };

    const onCompleted = (_: unknown, errors: readonly { message: string }[] | null | undefined) => {
      if (errors?.length) {
        toast({
          title: __("Error"),
          description: formatError(__("Cannot request access"), [...errors]),
          variant: "error",
        });
        return;
      }
      toast({
        title: __("Success"),
        description: __("Access request submitted successfully."),
        variant: "success",
      });
    };

    switch (node.__typename) {
      case "Document":
        requestAccess({
          variables: { input: { documentId: node.id } },
          onCompleted,
          onError,
        });
        break;
      case "CompliancePortalFile":
        requestFileAccess({
          variables: { input: { compliancePortalFileId: node.id } },
          onCompleted,
          onError,
        });
        break;
      case "AuditReport":
        requestReportAccess({
          variables: { input: { reportId: node.id } },
          onCompleted,
          onError,
        });
        break;
    }
  };

  const isRequesting = isRequestingAccess || isRequestingFileAccess || isRequestingReportAccess;
  const hasRequested = node.access?.status === "REQUESTED";
  const isPdf = node.__typename === "Document" || node.__typename === "AuditReport" || (node.__typename === "CompliancePortalFile" && pdfData !== null);

  const handleDownload = () => {
    if (!fileData || !nodeTitle) return;
    const base64 = extractBase64Data(fileData);
    const mimeType = extractMimeType(fileData);
    const byteCharacters = atob(base64);
    const byteNumbers = new Uint8Array(byteCharacters.length);
    for (let i = 0; i < byteCharacters.length; i++) {
      byteNumbers[i] = byteCharacters.charCodeAt(i);
    }
    const blob = new Blob([byteNumbers], { type: mimeType });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = nodeTitle;
    a.click();
    URL.revokeObjectURL(url);
  };

  return (
    <div className="flex flex-col h-screen bg-level-2">
      <header className="flex items-center h-12 gap-3 border-b border-border-solid px-4 flex-none bg-level-1">
        <Link
          to="/documents"
          className="size-8 grid place-items-center hover:bg-secondary-hover rounded-sm transition-all"
        >
          <IconChevronLeft size={16} />
        </Link>
        {logoFileUrl && (
          <img
            alt=""
            src={logoFileUrl}
            className="h-6 w-auto"
          />
        )}
        <span className="text-sm font-medium truncate">
          {nodeTitle}
        </span>
      </header>

      <main className="flex-1 min-h-0">
        {exportError
          ? (
              <div className="flex items-center justify-center h-full">
                <div className="text-center max-w-sm">
                  <h2 className="text-lg font-medium mb-2">
                    {__("Failed to load document")}
                  </h2>
                  <p className="text-sm text-txt-secondary">
                    {exportError}
                  </p>
                </div>
              </div>
            )
          : node.isUserAuthorized
            ? (
                isPdf
                  ? (
                      isExporting || !pdfData
                        ? (
                            <div className="flex items-center justify-center h-full">
                              <Spinner />
                            </div>
                          )
                        : (
                            <PDFPreview src={pdfData} name={nodeTitle ?? ""} />
                          )
                    )
                  : (
                      isExporting || !fileData
                        ? (
                            <div className="flex items-center justify-center h-full">
                              <Spinner />
                            </div>
                          )
                        : (
                            <div className="flex items-center justify-center h-full">
                              <div className="text-center max-w-sm">
                                <h2 className="text-lg font-medium mb-2">
                                  {nodeTitle}
                                </h2>
                                <p className="text-sm text-txt-secondary mb-6">
                                  {__("This file cannot be previewed in the browser.")}
                                </p>
                                <Button
                                  icon={IconArrowDown}
                                  onClick={handleDownload}
                                >
                                  {__("Download file")}
                                </Button>
                              </div>
                            </div>
                          )
                    )
              )
            : (
                <div className="flex items-center justify-center h-full">
                  <div className="text-center max-w-sm">
                    <IconLock size={32} className="mx-auto text-txt-tertiary mb-4" />
                    <h2 className="text-lg font-medium mb-2">
                      {nodeTitle}
                    </h2>
                    <p className="text-sm text-txt-secondary mb-6">
                      {__("This document requires access approval before viewing.")}
                    </p>
                    <Button
                      disabled={hasRequested || isRequesting}
                      icon={IconLock}
                      onClick={handleRequestAccess}
                    >
                      {hasRequested
                        ? __("Access requested")
                        : __("Request access")}
                    </Button>
                  </div>
                </div>
              )}
      </main>
    </div>
  );
}
