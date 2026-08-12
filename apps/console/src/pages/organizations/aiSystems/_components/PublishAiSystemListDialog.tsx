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
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { PublishAiSystemListDialog_organization$key } from "#/__generated__/core/PublishAiSystemListDialog_organization.graphql";
import type { PublishAiSystemListDialogMutation } from "#/__generated__/core/PublishAiSystemListDialogMutation.graphql";
import {
  PublishListDialog,
  type PublishListDialogInput,
} from "#/components/dialogs/PublishListDialog";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { useMutation } from "#/lib/relay/useMutation";

const publishAiSystemListDialogFragment = graphql`
  fragment PublishAiSystemListDialog_organization on Organization {
    aiSystemsDocument {
      defaultApprovers {
        id
      }
    }
  }
`;

const publishMutation = graphql`
  mutation PublishAiSystemListDialogMutation(
    $input: PublishAiSystemListInput!
  ) {
    publishAiSystemList(input: $input) {
      documentEdge {
        node {
          id
        }
      }
    }
  }
`;

interface PublishAiSystemListDialogProps {
  children: ReactNode;
  organizationKey: PublishAiSystemListDialog_organization$key;
  onPublished?: (documentId: string) => void;
}

export function PublishAiSystemListDialog({
  children,
  organizationKey,
  onPublished,
}: PublishAiSystemListDialogProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const organization = useFragment(
    publishAiSystemListDialogFragment,
    organizationKey,
  );
  const defaultApproverIds = (
    organization.aiSystemsDocument?.defaultApprovers ?? []
  ).map(approver => approver.id);

  const [publish, isPublishing] = useMutation<PublishAiSystemListDialogMutation>(
    publishMutation,
    {
      errorToast: false,
    },
  );

  const onPublish = async (input: PublishListDialogInput) => {
    const response = await publish({
      variables: { input },
    });
    return response.publishAiSystemList?.documentEdge?.node?.id;
  };

  return (
    <PublishListDialog
      organizationId={organizationId}
      defaultApproverIds={defaultApproverIds}
      isPublishing={isPublishing}
      onPublish={onPublish}
      onPublished={onPublished}
      title={t("publishAiSystemListDialog.title")}
      publishedMessage={t("publishAiSystemListDialog.messages.published")}
      publishError={t("publishAiSystemListDialog.errors.publish")}
    >
      {children}
    </PublishListDialog>
  );
}
