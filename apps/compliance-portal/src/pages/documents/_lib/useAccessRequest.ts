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

import { useCallback } from "react";
import { useTranslation } from "react-i18next";
import { graphql } from "relay-runtime";

import {
  buildRequestAccessContinueUrl,
  REQUEST_DOCUMENT_PARAM,
  REQUEST_FILE_PARAM,
  REQUEST_REPORT_PARAM,
} from "#/lib/auth/continueUrl";
import { useMutation } from "#/lib/relay/useMutation";

import type { useAccessRequestDocumentMutation } from "./__generated__/useAccessRequestDocumentMutation.graphql";
import type { useAccessRequestFileMutation } from "./__generated__/useAccessRequestFileMutation.graphql";
import type { useAccessRequestReportMutation } from "./__generated__/useAccessRequestReportMutation.graphql";
import type { DocumentKind } from "./useDocumentExport";

export interface AccessRequest {
  requestAccess: () => void;
  isRequesting: boolean;
}

// Each mutation echoes the updated access record so Relay flips the row to its
// "requested" state in place, without a refetch.
const documentMutation = graphql`
  mutation useAccessRequestDocumentMutation($input: RequestDocumentAccessInput!) {
    requestDocumentAccess(input: $input) {
      document {
        id
        access {
          id
          status
        }
      }
    }
  }
`;

const reportMutation = graphql`
  mutation useAccessRequestReportMutation($input: RequestReportAccessInput!) {
    requestReportAccess(input: $input) {
      audit {
        id
        reportFile {
          id
          access {
            id
            status
          }
        }
      }
    }
  }
`;

const fileMutation = graphql`
  mutation useAccessRequestFileMutation($input: RequestCompliancePortalFileAccessInput!) {
    requestCompliancePortalFileAccess(input: $input) {
      file {
        id
        access {
          id
          status
        }
      }
    }
  }
`;

// Auth, full-name, and NDA gates are consumed by useMutation (with a
// marker-bearing continueUrl so the request resumes after the gate). Success
// and non-gate failures use the shared toast feedback.
function useAccessRequestFeedback() {
  const { t } = useTranslation();
  return {
    successMessage: t("auth.requestAccess.success"),
    errorToast: t("auth.errors.requestFailed"),
  };
}

export function useRequestDocumentAccess(id: string): AccessRequest {
  const feedback = useAccessRequestFeedback();
  const [mutate, isRequesting] = useMutation<useAccessRequestDocumentMutation>(
    documentMutation,
    feedback,
  );

  const requestAccess = useCallback(() => {
    void mutate(
      { variables: { input: { documentId: id } } },
      { continueUrl: buildRequestAccessContinueUrl(REQUEST_DOCUMENT_PARAM, id) },
    ).catch(() => {});
  }, [mutate, id]);

  return { requestAccess, isRequesting };
}

export function useRequestReportAccess(id: string): AccessRequest {
  const feedback = useAccessRequestFeedback();
  const [mutate, isRequesting] = useMutation<useAccessRequestReportMutation>(
    reportMutation,
    feedback,
  );

  const requestAccess = useCallback(() => {
    void mutate(
      { variables: { input: { reportId: id } } },
      { continueUrl: buildRequestAccessContinueUrl(REQUEST_REPORT_PARAM, id) },
    ).catch(() => {});
  }, [mutate, id]);

  return { requestAccess, isRequesting };
}

export function useRequestFileAccess(id: string): AccessRequest {
  const feedback = useAccessRequestFeedback();
  const [mutate, isRequesting] = useMutation<useAccessRequestFileMutation>(
    fileMutation,
    feedback,
  );

  const requestAccess = useCallback(() => {
    void mutate(
      { variables: { input: { compliancePortalFileId: id } } },
      { continueUrl: buildRequestAccessContinueUrl(REQUEST_FILE_PARAM, id) },
    ).catch(() => {});
  }, [mutate, id]);

  return { requestAccess, isRequesting };
}

// Resolves the right request hook for a viewer node resolved by kind. Safe to
// call unconditionally (all three hooks run); returns the one matching `kind`.
export function useAccessRequest(kind: DocumentKind, id: string): AccessRequest {
  const document = useRequestDocumentAccess(id);
  const report = useRequestReportAccess(id);
  const file = useRequestFileAccess(id);

  switch (kind) {
    case "Document":
      return document;
    case "AuditReport":
      return report;
    case "CompliancePortalFile":
      return file;
  }
}
