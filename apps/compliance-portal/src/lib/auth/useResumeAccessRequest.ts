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

import { Toast } from "@base-ui/react/toast";
import { useEffect, useRef } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate, useSearchParams } from "react-router";
import type { PayloadError } from "relay-runtime";
import { graphql } from "relay-runtime";

import {
  buildRequestAccessContinueUrl,
  buildRequestAllContinueUrl,
  gateRedirectPath,
  REQUEST_ALL_PARAM,
  REQUEST_DOCUMENT_PARAM,
  REQUEST_FILE_PARAM,
  REQUEST_REPORT_PARAM,
} from "#/lib/auth/continueUrl";
import { useMutation } from "#/lib/relay/useMutation";

import type { useResumeAccessRequest_documentMutation } from "./__generated__/useResumeAccessRequest_documentMutation.graphql";
import type { useResumeAccessRequest_fileMutation } from "./__generated__/useResumeAccessRequest_fileMutation.graphql";
import type { useResumeAccessRequest_reportMutation } from "./__generated__/useResumeAccessRequest_reportMutation.graphql";
import type { useResumeAccessRequestMutation } from "./__generated__/useResumeAccessRequestMutation.graphql";

const requestAllAccessesMutation = graphql`
  mutation useResumeAccessRequestMutation {
    requestAllAccesses {
      trustCenterAccess {
        id
      }
    }
  }
`;

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
    }
  }
`;

const requestFileMutation = graphql`
  mutation useResumeAccessRequest_fileMutation($input: RequestTrustCenterFileAccessInput!) {
    requestTrustCenterFileAccess(input: $input) {
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

// After a user signs in through the dialog, they land back on the page that
// carried a deferred access marker. This hook fires the matching mutation once
// (when authenticated) — request-all from the top bar, or a single
// document / report / file requested from a locked row — routes to the
// full-name gate when the backend asks for it, and clears the marker so a
// refresh never re-triggers it.
export function useResumeAccessRequest(isAuthenticated: boolean) {
  const [searchParams, setSearchParams] = useSearchParams();
  const navigate = useNavigate();
  const toast = Toast.useToastManager();
  const { t } = useTranslation();
  const firedRef = useRef(false);

  const [requestAllAccesses] = useMutation<useResumeAccessRequestMutation>(
    requestAllAccessesMutation,
    { errorToast: false },
  );
  const [requestDocumentAccess] = useMutation<useResumeAccessRequest_documentMutation>(
    requestDocumentMutation,
    { errorToast: false },
  );
  const [requestReportAccess] = useMutation<useResumeAccessRequest_reportMutation>(
    requestReportMutation,
    { errorToast: false },
  );
  const [requestFileAccess] = useMutation<useResumeAccessRequest_fileMutation>(
    requestFileMutation,
    { errorToast: false },
  );

  useEffect(() => {
    if (!isAuthenticated || firedRef.current) {
      return;
    }

    const documentId = searchParams.get(REQUEST_DOCUMENT_PARAM);
    const reportId = searchParams.get(REQUEST_REPORT_PARAM);
    const fileId = searchParams.get(REQUEST_FILE_PARAM);
    const all = searchParams.get(REQUEST_ALL_PARAM) === "true";

    if (!documentId && !reportId && !fileId && !all) {
      return;
    }

    firedRef.current = true;

    // Shared outcome handling. The full-name and NDA gates are thrown by the
    // fetch layer, so they arrive in `onError` and deep-link to their gate page,
    // preserving the marker so the request resumes once cleared. Other failures
    // toast; success confirms.
    const makeHandlers = (continueUrl: string) => ({
      onCompleted: (_response: unknown, errors: PayloadError[] | null) => {
        if (errors && errors.length > 0) {
          toast.add({ title: t("auth.errors.requestFailed"), type: "error" });
          return;
        }
        toast.add({ title: t("auth.requestAccess.success"), type: "success" });
      },
      onError: (error: Error) => {
        const gatePath = gateRedirectPath(error, continueUrl);
        if (gatePath) {
          void navigate(gatePath);
          return;
        }
        toast.add({ title: t("auth.errors.requestFailed"), type: "error" });
      },
    });

    // Drop the marker up front so a reload can't queue a second request.
    const clear = (param: string) => {
      searchParams.delete(param);
      setSearchParams(searchParams, { replace: true });
    };

    if (documentId) {
      const continueUrl = buildRequestAccessContinueUrl(REQUEST_DOCUMENT_PARAM, documentId);
      clear(REQUEST_DOCUMENT_PARAM);
      void requestDocumentAccess({
        variables: { input: { documentId } },
        ...makeHandlers(continueUrl),
      }).catch(() => {});
      return;
    }

    if (reportId) {
      const continueUrl = buildRequestAccessContinueUrl(REQUEST_REPORT_PARAM, reportId);
      clear(REQUEST_REPORT_PARAM);
      void requestReportAccess({
        variables: { input: { reportId } },
        ...makeHandlers(continueUrl),
      }).catch(() => {});
      return;
    }

    if (fileId) {
      const continueUrl = buildRequestAccessContinueUrl(REQUEST_FILE_PARAM, fileId);
      clear(REQUEST_FILE_PARAM);
      void requestFileAccess({
        variables: { input: { trustCenterFileId: fileId } },
        ...makeHandlers(continueUrl),
      }).catch(() => {});
      return;
    }

    const allContinueUrl = buildRequestAllContinueUrl();
    clear(REQUEST_ALL_PARAM);
    void requestAllAccesses({
      variables: {},
      ...makeHandlers(allContinueUrl),
    }).catch(() => {});
  }, [
    isAuthenticated,
    navigate,
    requestAllAccesses,
    requestDocumentAccess,
    requestReportAccess,
    requestFileAccess,
    searchParams,
    setSearchParams,
    t,
    toast,
  ]);
}
