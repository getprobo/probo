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

import { useEffect, useRef } from "react";
import { useTranslation } from "react-i18next";
import { useSearchParams } from "react-router";
import { graphql } from "relay-runtime";

import {
  buildRequestAccessContinueUrl,
  REQUEST_DOCUMENT_PARAM,
  REQUEST_FILE_PARAM,
  REQUEST_REPORT_PARAM,
} from "#/lib/auth/continueUrl";
import { useMutation } from "#/lib/relay/useMutation";

import type { useResumeAccessRequest_documentMutation } from "./__generated__/useResumeAccessRequest_documentMutation.graphql";
import type { useResumeAccessRequest_fileMutation } from "./__generated__/useResumeAccessRequest_fileMutation.graphql";
import type { useResumeAccessRequest_reportMutation } from "./__generated__/useResumeAccessRequest_reportMutation.graphql";

const requestDocumentMutation = graphql`
  mutation useResumeAccessRequest_documentMutation($input: RequestDocumentAccessInput!) {
    requestDocumentAccess(input: $input) {
      document {
        id
        access {
          id
          status
        }
      }
      compliancePortal {
        id
        viewerHasRequestedAccess
      }
    }
  }
`;

const requestReportMutation = graphql`
  mutation useResumeAccessRequest_reportMutation($input: RequestReportAccessInput!) {
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
      compliancePortal {
        id
        viewerHasRequestedAccess
      }
    }
  }
`;

const requestFileMutation = graphql`
  mutation useResumeAccessRequest_fileMutation($input: RequestCompliancePortalFileAccessInput!) {
    requestCompliancePortalFileAccess(input: $input) {
      file {
        id
        access {
          id
          status
        }
      }
      compliancePortal {
        id
        viewerHasRequestedAccess
      }
    }
  }
`;

// After a user signs in through OAuth /initiate, they land back on the page that
// carried a deferred access marker. This hook fires the matching mutation once
// (when authenticated) — a single document / report / file requested from a
// locked row — and clears the marker so a refresh never re-triggers it. Auth
// gates (full-name / NDA) are consumed by useMutation with a continueUrl that
// still carries the marker so the request can resume after the next gate.
export function useResumeAccessRequest(isAuthenticated: boolean) {
  const [searchParams, setSearchParams] = useSearchParams();
  const { t } = useTranslation();
  const firedRef = useRef(false);

  const feedback = {
    successMessage: t("auth.requestAccess.success"),
    errorToast: t("auth.errors.requestFailed"),
  };

  const [requestDocumentAccess] = useMutation<useResumeAccessRequest_documentMutation>(
    requestDocumentMutation,
    feedback,
  );
  const [requestReportAccess] = useMutation<useResumeAccessRequest_reportMutation>(
    requestReportMutation,
    feedback,
  );
  const [requestFileAccess] = useMutation<useResumeAccessRequest_fileMutation>(
    requestFileMutation,
    feedback,
  );

  useEffect(() => {
    if (!isAuthenticated || firedRef.current) {
      return;
    }

    const documentId = searchParams.get(REQUEST_DOCUMENT_PARAM);
    const reportId = searchParams.get(REQUEST_REPORT_PARAM);
    const fileId = searchParams.get(REQUEST_FILE_PARAM);

    if (!documentId && !reportId && !fileId) {
      return;
    }

    firedRef.current = true;

    // Drop the marker up front so a reload can't queue a second request. The
    // continueUrl passed to the mutation still includes the marker so a further
    // full-name / NDA gate can re-queue the same resume.
    const clear = (param: string) => {
      searchParams.delete(param);
      setSearchParams(searchParams, { replace: true });
    };

    if (documentId) {
      const continueUrl = buildRequestAccessContinueUrl(REQUEST_DOCUMENT_PARAM, documentId);
      clear(REQUEST_DOCUMENT_PARAM);
      void requestDocumentAccess(
        { variables: { input: { documentId } } },
        { continueUrl },
      ).catch(() => {});
      return;
    }

    if (reportId) {
      const continueUrl = buildRequestAccessContinueUrl(REQUEST_REPORT_PARAM, reportId);
      clear(REQUEST_REPORT_PARAM);
      void requestReportAccess(
        { variables: { input: { reportId } } },
        { continueUrl },
      ).catch(() => {});
      return;
    }

    if (fileId) {
      const continueUrl = buildRequestAccessContinueUrl(REQUEST_FILE_PARAM, fileId);
      clear(REQUEST_FILE_PARAM);
      void requestFileAccess(
        { variables: { input: { compliancePortalFileId: fileId } } },
        { continueUrl },
      ).catch(() => {});
    }
  }, [
    isAuthenticated,
    requestDocumentAccess,
    requestReportAccess,
    requestFileAccess,
    searchParams,
    setSearchParams,
  ]);
}
