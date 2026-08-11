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

import type { PublishAssetListDialogMutation } from "#/__generated__/core/PublishAssetListDialogMutation.graphql";
import {
  PublishListDialog,
  type PublishListDialogInput,
} from "#/components/dialogs/PublishListDialog";

const publishMutation = graphql`
  mutation PublishAssetListDialogMutation(
    $input: PublishAssetListInput!
  ) {
    publishAssetList(input: $input) {
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

export function PublishAssetListDialog({
  children,
  organizationId,
  defaultApproverIds,
  onPublished,
}: Props) {
  const { t } = useTranslation();
  const [publish, isPublishing] = useMutation<PublishAssetListDialogMutation>(publishMutation);

  const onPublish = (input: PublishListDialogInput) =>
    new Promise<string | null | undefined>((resolve, reject) => {
      publish({
        variables: { input },
        onCompleted: (response) => {
          resolve(response.publishAssetList?.documentEdge?.node?.id);
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
        title: t("publishAssetListDialog.title"),
        description: t("publishAssetListDialog.description"),
        approvers: t("publishAssetListDialog.fields.approvers"),
        approversPlaceholder: t("publishAssetListDialog.fields.approversPlaceholder"),
        publishMinor: t("publishAssetListDialog.actions.publishMinor"),
        publish: t("publishAssetListDialog.actions.publish"),
        requestApproval: t("publishAssetListDialog.actions.requestApproval"),
        successTitle: t("publishAssetListDialog.messages.success"),
        published: t("publishAssetListDialog.messages.published"),
        approvalRequested: t("publishAssetListDialog.messages.approvalRequested"),
        errorTitle: t("publishAssetListDialog.messages.error"),
        publishError: t("publishAssetListDialog.errors.publish"),
      }}
    >
      {children}
    </PublishListDialog>
  );
}
