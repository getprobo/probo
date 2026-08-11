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

import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { PublishDataProtectionImpactAssessmentListDialogMutation } from "#/__generated__/core/PublishDataProtectionImpactAssessmentListDialogMutation.graphql";
import {
  PublishListDialog,
  type PublishListDialogInput,
} from "#/components/dialogs/PublishListDialog";

const publishMutation = graphql`
  mutation PublishDataProtectionImpactAssessmentListDialogMutation(
    $input: PublishDataProtectionImpactAssessmentListInput!
  ) {
    publishDataProtectionImpactAssessmentList(input: $input) {
      documentEdge {
        node {
          id
        }
      }
    }
  }
`;

type Props = {
  children: ReactNode;
  organizationId: string;
  defaultApproverIds?: string[];
  onPublished?: (documentId: string) => void;
};

export function PublishDataProtectionImpactAssessmentListDialog({
  children,
  organizationId,
  defaultApproverIds,
  onPublished,
}: Props) {
  const { t } = useTranslation();
  const [publish, isPublishing] = useMutation<PublishDataProtectionImpactAssessmentListDialogMutation>(publishMutation);

  const onPublish = (input: PublishListDialogInput) =>
    new Promise<string | null | undefined>((resolve, reject) => {
      publish({
        variables: { input },
        onCompleted: (response) => {
          resolve(response.publishDataProtectionImpactAssessmentList?.documentEdge?.node?.id);
        },
        onError: reject,
      });
    });

  return (
    <PublishListDialog
      organizationId={organizationId}
      defaultApproverIds={defaultApproverIds}
      isPublishing={isPublishing}
      onPublish={onPublish}
      onPublished={onPublished}
      labels={{
        title: t("publishDpiaListDialog.title"),
        description: t("publishDpiaListDialog.description"),
        approvers: t("publishDpiaListDialog.fields.approvers"),
        approversPlaceholder: t("publishDpiaListDialog.fields.approversPlaceholder"),
        publishMinor: t("publishDpiaListDialog.actions.publishMinor"),
        publish: t("publishDpiaListDialog.actions.publish"),
        requestApproval: t("publishDpiaListDialog.actions.requestApproval"),
        successTitle: t("publishDpiaListDialog.messages.success"),
        published: t("publishDpiaListDialog.messages.published"),
        approvalRequested: t("publishDpiaListDialog.messages.approvalRequested"),
        errorTitle: t("publishDpiaListDialog.messages.error"),
        publishError: t("publishDpiaListDialog.errors.publish"),
      }}
    >
      {children}
    </PublishListDialog>
  );
}
