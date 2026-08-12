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
      compliancePortal {
        id
        viewerHasRequestedAccess
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
// single mutation. Auth / full-name / NDA gates are consumed by useMutation
// (current URL as continue — no batch marker, so the selection is not resumed
// after the gate). Success and non-gate failures use the shared toast feedback.
export function useBulkRequestAccess(onSuccess?: () => void): BulkAccessRequest {
  const { t } = useTranslation();
  const [mutate, isRequesting] = useMutation<useBulkRequestAccessMutation>(bulkMutation, {
    successMessage: t("auth.requestAccess.success"),
    errorToast: t("auth.errors.requestFailed"),
  });

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
        onCompleted: (_response, errors) => {
          if (!errors || errors.length === 0) {
            onSuccess?.();
          }
        },
      }).catch(() => {});
    },
    [mutate, onSuccess],
  );

  return { requestAccess, isRequesting };
}
