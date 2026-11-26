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

import type { ReactElement } from "react";
import { useTranslation } from "react-i18next";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { PublishStatementOfApplicabilityDialogMutation } from "#/__generated__/core/PublishStatementOfApplicabilityDialogMutation.graphql";
import {
  PublishListDialog,
  type PublishListDialogApprovers,
  type PublishListDialogInput,
} from "#/components/dialogs/PublishListDialog";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const publishMutation = graphql`
  mutation PublishStatementOfApplicabilityDialogMutation(
    $input: PublishStatementOfApplicabilityInput!
  ) {
    publishStatementOfApplicability(input: $input) {
      documentEdge {
        node {
          id
        }
      }
    }
  }
`;

type Props = {
  children: ReactElement;
  statementOfApplicabilityId: string;
  defaultApprovers: PublishListDialogApprovers;
  onPublished?: (documentId: string) => void;
};

export function PublishStatementOfApplicabilityDialog({
  children,
  statementOfApplicabilityId,
  defaultApprovers,
  onPublished,
}: Props) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const [publish, isPublishing]
    = useMutation<PublishStatementOfApplicabilityDialogMutation>(publishMutation);

  const onPublish = ({ minor, approverIds }: PublishListDialogInput) =>
    new Promise<string | null | undefined>((resolve, reject) => {
      publish({
        variables: {
          input: { statementOfApplicabilityId, minor, approverIds },
        },
        onCompleted: (response) => {
          resolve(response.publishStatementOfApplicability?.documentEdge?.node?.id);
        },
        onError: reject,
      });
    });

  return (
    <PublishListDialog
      organizationId={organizationId}
      defaultApprovers={defaultApprovers}
      isPublishing={isPublishing}
      onPublish={onPublish}
      onPublished={onPublished}
      title={t("publishStatementOfApplicabilityDialog.title")}
      description={t("publishStatementOfApplicabilityDialog.description")}
      publishedMessage={t("publishStatementOfApplicabilityDialog.messages.published")}
      publishError={t("publishStatementOfApplicabilityDialog.errors.publish")}
    >
      {children}
    </PublishListDialog>
  );
}
