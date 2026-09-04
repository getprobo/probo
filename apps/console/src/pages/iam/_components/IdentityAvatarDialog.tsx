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

import { EditAvatarDialog } from "@probo/ui/src/v2/EditAvatarDialog/EditAvatarDialog";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { IdentityAvatarDialog_deleteMutation } from "#/__generated__/iam/IdentityAvatarDialog_deleteMutation.graphql";
import type { IdentityAvatarDialog_identity$key } from "#/__generated__/iam/IdentityAvatarDialog_identity.graphql";
import type { IdentityAvatarDialog_updateMutation } from "#/__generated__/iam/IdentityAvatarDialog_updateMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

export const identityAvatarDialogFragment = graphql`
  fragment IdentityAvatarDialog_identity on Identity {
    fullName
    avatar {
      downloadUrl
    }
  }
`;

const updateIdentityAvatarMutation = graphql`
  mutation IdentityAvatarDialog_updateMutation(
    $input: UpdateIdentityAvatarInput!
  ) {
    updateIdentityAvatar(input: $input) {
      identity {
        id
        avatar {
          downloadUrl
        }
      }
    }
  }
`;

const deleteIdentityAvatarMutation = graphql`
  mutation IdentityAvatarDialog_deleteMutation {
    deleteIdentityAvatar {
      identity {
        id
        avatar {
          downloadUrl
        }
      }
    }
  }
`;

export interface IdentityAvatarDialogProps {
  identityKey: IdentityAvatarDialog_identity$key;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function IdentityAvatarDialog({
  identityKey,
  open,
  onOpenChange,
}: IdentityAvatarDialogProps) {
  const { t } = useTranslation();
  const identity = useFragment(identityAvatarDialogFragment, identityKey);
  const [updateAvatar, isUploading] = useMutation<IdentityAvatarDialog_updateMutation>(
    updateIdentityAvatarMutation,
    {
      successMessage: t("editAvatar.messages.updated"),
      errorToast: t("editAvatar.errors.update"),
    },
  );
  const [deleteAvatar, isRemoving] = useMutation<IdentityAvatarDialog_deleteMutation>(
    deleteIdentityAvatarMutation,
    {
      successMessage: t("editAvatar.messages.removed"),
      errorToast: t("editAvatar.errors.remove"),
    },
  );

  return (
    <EditAvatarDialog
      open={open}
      fullName={identity.fullName}
      src={identity.avatar?.downloadUrl}
      uploading={isUploading}
      removing={isRemoving}
      title={t("editAvatar.title")}
      description={t("editAvatar.description")}
      uploadLabel={t("editAvatar.actions.upload")}
      replaceLabel={t("editAvatar.actions.replace")}
      removeLabel={t("editAvatar.actions.remove")}
      closeLabel={t("editAvatar.actions.close")}
      onOpenChange={onOpenChange}
      onUpload={(file) => {
        void updateAvatar({
          variables: { input: { file: null } },
          uploadables: { "input.file": file },
        });
      }}
      onRemove={() => {
        void deleteAvatar({ variables: {} });
      }}
    />
  );
}
