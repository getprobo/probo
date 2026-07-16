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
import {
  FullNameRequiredError,
  NDASignatureRequiredError,
  UnAuthenticatedError,
} from "@probo/relay";
import { useCallback, useMemo } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router";
import type { PayloadError } from "relay-runtime";
import { graphql } from "relay-runtime";

import {
  buildRequestAccessContinueUrl,
  REQUEST_DOCUMENT_PARAM,
  REQUEST_FILE_PARAM,
  REQUEST_REPORT_PARAM,
} from "#/lib/auth/continueUrl";
import { useSignInDialog } from "#/lib/auth/signInDialogContext";
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
  mutation useAccessRequestFileMutation($input: RequestTrustCenterFileAccessInput!) {
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

// Shared success / error handling for a single access request. The auth,
// full-name, and NDA gates are thrown by the fetch layer, so they surface in
// `onError` (not `onError`'s GraphQL-errors argument): unauthenticated opens the
// sign-in dialog, full-name deep-links to its gate (both deferring the request
// via the continue URL), NDA is a toast (its primary path is the query-load
// boundary), and everything else is a generic toast.
function useAccessRequestHandlers(param: string, id: string) {
  const { openSignIn } = useSignInDialog();
  const navigate = useNavigate();
  const toast = Toast.useToastManager();
  const { t } = useTranslation();

  return useMemo(
    () => ({
      onCompleted: (_response: unknown, errors: PayloadError[] | null) => {
        if (errors && errors.length > 0) {
          toast.add({ title: t("auth.errors.requestFailed"), type: "error" });
          return;
        }
        toast.add({ title: t("auth.requestAccess.success"), type: "success" });
      },
      onError: (error: Error) => {
        // Not signed in: open the dialog, deferring this request until the user
        // lands back authenticated (see useResumeAccessRequest).
        if (error instanceof UnAuthenticatedError) {
          openSignIn({ continueTo: buildRequestAccessContinueUrl(param, id) });
          return;
        }
        // Missing profile name: send them to the full-name gate, preserving the
        // marker so the request resumes afterwards.
        if (error instanceof FullNameRequiredError) {
          const continueUrl = buildRequestAccessContinueUrl(param, id);
          void navigate(`/full-name?continue=${encodeURIComponent(continueUrl)}`);
          return;
        }
        // NDA is enforced at query load (the route boundary redirects to /nda);
        // here we only inform, matching the trust app.
        if (error instanceof NDASignatureRequiredError) {
          toast.add({ title: t("auth.errors.ndaRequired"), type: "error" });
          return;
        }
        toast.add({ title: t("auth.errors.requestFailed"), type: "error" });
      },
    }),
    [openSignIn, navigate, toast, t, param, id],
  );
}

export function useRequestDocumentAccess(id: string): AccessRequest {
  const handlers = useAccessRequestHandlers(REQUEST_DOCUMENT_PARAM, id);
  const [mutate, isRequesting] = useMutation<useAccessRequestDocumentMutation>(
    documentMutation,
    { errorToast: false },
  );

  const requestAccess = useCallback(() => {
    void mutate({ variables: { input: { documentId: id } }, ...handlers }).catch(() => {});
  }, [mutate, id, handlers]);

  return { requestAccess, isRequesting };
}

export function useRequestReportAccess(id: string): AccessRequest {
  const handlers = useAccessRequestHandlers(REQUEST_REPORT_PARAM, id);
  const [mutate, isRequesting] = useMutation<useAccessRequestReportMutation>(
    reportMutation,
    { errorToast: false },
  );

  const requestAccess = useCallback(() => {
    void mutate({ variables: { input: { reportId: id } }, ...handlers }).catch(() => {});
  }, [mutate, id, handlers]);

  return { requestAccess, isRequesting };
}

export function useRequestFileAccess(id: string): AccessRequest {
  const handlers = useAccessRequestHandlers(REQUEST_FILE_PARAM, id);
  const [mutate, isRequesting] = useMutation<useAccessRequestFileMutation>(
    fileMutation,
    { errorToast: false },
  );

  const requestAccess = useCallback(() => {
    void mutate({ variables: { input: { trustCenterFileId: id } }, ...handlers }).catch(() => {});
  }, [mutate, id, handlers]);

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
    case "TrustCenterFile":
      return file;
  }
}
