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

import type { PublishRiskListDialogMutation } from "#/__generated__/core/PublishRiskListDialogMutation.graphql";
import {
  PublishListDialog,
  type PublishListDialogInput,
} from "#/components/dialogs/PublishListDialog";

const publishMutation = graphql`
  mutation PublishRiskListDialogMutation(
    $input: PublishRiskListInput!
  ) {
    publishRiskList(input: $input) {
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

export function PublishRiskListDialog({
  children,
  organizationId,
  defaultApproverIds,
  onPublished,
}: Props) {
  const { t } = useTranslation();
  const [publish, isPublishing] = useMutation<PublishRiskListDialogMutation>(publishMutation);

  const onPublish = (input: PublishListDialogInput) =>
    new Promise<string | null | undefined>((resolve, reject) => {
      publish({
        variables: { input },
        onCompleted: (response) => {
          resolve(response.publishRiskList?.documentEdge?.node?.id);
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
        title: t("publishRiskListDialog.title"),
        description: t("publishRiskListDialog.description"),
        approvers: t("publishRiskListDialog.fields.approvers"),
        approversPlaceholder: t("publishRiskListDialog.fields.approversPlaceholder"),
        publishMinor: t("publishRiskListDialog.actions.publishMinor"),
        publish: t("publishRiskListDialog.actions.publish"),
        requestApproval: t("publishRiskListDialog.actions.requestApproval"),
        successTitle: t("publishRiskListDialog.messages.success"),
        published: t("publishRiskListDialog.messages.published"),
        approvalRequested: t("publishRiskListDialog.messages.approvalRequested"),
        errorTitle: t("publishRiskListDialog.messages.error"),
        publishError: t("publishRiskListDialog.errors.publish"),
      }}
    >
      {children}
    </PublishListDialog>
  );
}
