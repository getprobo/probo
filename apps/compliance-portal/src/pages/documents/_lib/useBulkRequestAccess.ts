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
import { UnAuthenticatedError } from "@probo/relay";
import { useCallback } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router";
import type { PayloadError } from "relay-runtime";
import { graphql } from "relay-runtime";

import { gateRedirectPath, getSafeContinueUrl, redirectToInitiate } from "#/lib/auth/continueUrl";
import { useLocale } from "#/lib/i18n/useLocale";
import { useMutation } from "#/lib/relay/useMutation";

import type { useBulkRequestAccessMutation } from "./__generated__/useBulkRequestAccessMutation.graphql";
import type { DocumentKind } from "./DocumentSelectionContext";

// One selection-scoped call for the whole batch. The payload echoes each
// affected node's updated access record so Relay flips every requested row to
// its "pending" state in place, without a refetch.
const bulkMutation = graphql`
  mutation useBulkRequestAccessMutation($input: RequestAccessesInput!) {
    requestAccesses(input: $input) {
      documents {
        id
        access {
          id
          status
        }
      }
      audits {
        id
        reportFile {
          id
          access {
            id
            status
          }
        }
      }
      files {
        id
        access {
          id
          status
        }
      }
    }
  }
`;

export interface BulkAccessRequestEntry {
  id: string;
  kind: DocumentKind;
}

export interface BulkAccessRequest {
  requestAccess: (entries: BulkAccessRequestEntry[]) => void;
  isRequesting: boolean;
}

// Requests access for a mixed selection of documents / reports / files in a
// single mutation. Auth, full-name, and NDA gates are thrown by the fetch layer
// and surface in `onError`: unauthenticated redirects to OAuth /initiate, while
// full-name and NDA deep-link to their gate page. Unlike the single-row flow
// this is a "simple redirect": the current URL carries no batch marker, so the
// selection is not resumed after the gate is cleared (the user re-selects).
export function useBulkRequestAccess(onSuccess?: () => void): BulkAccessRequest {
  const navigate = useNavigate();
  const locale = useLocale();
  const toast = Toast.useToastManager();
  const { t } = useTranslation();
  const [mutate, isRequesting] = useMutation<useBulkRequestAccessMutation>(
    bulkMutation,
    { errorToast: false },
  );

  const requestAccess = useCallback(
    (entries: BulkAccessRequestEntry[]) => {
      const documentIds: string[] = [];
      const reportIds: string[] = [];
      const compliancePortalFileIds: string[] = [];

      for (const entry of entries) {
        switch (entry.kind) {
          case "Document":
            documentIds.push(entry.id);
            break;
          case "AuditReport":
            reportIds.push(entry.id);
            break;
          case "CompliancePortalFile":
            compliancePortalFileIds.push(entry.id);
            break;
        }
      }

      void mutate({
        variables: { input: { documentIds, reportIds, compliancePortalFileIds } },
        onCompleted: (_response: unknown, errors: PayloadError[] | null) => {
          if (errors && errors.length > 0) {
            toast.add({ title: t("auth.errors.requestFailed"), type: "error" });
            return;
          }
          toast.add({ title: t("auth.requestAccess.success"), type: "success" });
          onSuccess?.();
        },
        onError: (error: Error) => {
          const continueUrl = getSafeContinueUrl(window.location.href);

          if (error instanceof UnAuthenticatedError) {
            redirectToInitiate(continueUrl);
            return;
          }

          const gatePath = gateRedirectPath(error, continueUrl, locale);
          if (gatePath) {
            void navigate(gatePath);
            return;
          }

          toast.add({ title: t("auth.errors.requestFailed"), type: "error" });
        },
      }).catch(() => {});
    },
    [mutate, toast, t, navigate, locale, onSuccess],
  );

  return { requestAccess, isRequesting };
}
