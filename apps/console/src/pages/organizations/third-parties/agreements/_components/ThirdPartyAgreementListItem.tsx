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

import { downloadFile, formatError } from "@probo/helpers";
import { dateFormat } from "@probo/i18n";
import {
  Button,
  IconPageTextLine,
  IconPencil,
  IconTrashCan,
  useConfirm,
} from "@probo/ui";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment, useMutation } from "react-relay";

import type { ThirdPartyAgreementListItem_agreement$key } from "#/__generated__/core/ThirdPartyAgreementListItem_agreement.graphql";
import type { ThirdPartyAgreementListItemDeleteMutation } from "#/__generated__/core/ThirdPartyAgreementListItemDeleteMutation.graphql";

import { EditThirdPartyAgreementDialog } from "./EditThirdPartyAgreementDialog";
import { ThirdPartyAgreementPreviewDialog } from "./ThirdPartyAgreementPreviewDialog";

const agreementFragment = graphql`
  fragment ThirdPartyAgreementListItem_agreement on ThirdPartyAgreement {
    id
    name
    type
    validity {
      start
      end
    }
    file {
      fileName
      downloadUrl
    }
    canUpdate: permission(action: "core:thirdParty-agreement:update")
    canDelete: permission(action: "core:thirdParty-agreement:delete")
    ...EditThirdPartyAgreementDialog_agreement
    ...ThirdPartyAgreementPreviewDialog_agreement
  }
`;

const deleteAgreementMutation = graphql`
  mutation ThirdPartyAgreementListItemDeleteMutation(
    $input: DeleteThirdPartyAgreementInput!
    $connections: [ID!]!
  ) {
    deleteThirdPartyAgreement(input: $input) {
      deletedThirdPartyAgreementId @deleteEdge(connections: $connections)
    }
  }
`;

interface ThirdPartyAgreementListItemProps {
  agreementKey: ThirdPartyAgreementListItem_agreement$key;
  connectionId: string;
}

export function ThirdPartyAgreementListItem({
  agreementKey,
  connectionId,
}: ThirdPartyAgreementListItemProps) {
  const { t, i18n } = useTranslation();
  const agreement = useFragment(agreementFragment, agreementKey);
  const confirm = useConfirm();
  const toast = useToast();
  const [previewOpen, setPreviewOpen] = useState(false);
  const [deleteAgreement] = useMutation<ThirdPartyAgreementListItemDeleteMutation>(
    deleteAgreementMutation,
  );

  const handleDelete = () => {
    confirm(
      () =>
        new Promise<void>((resolve) => {
          void deleteAgreement({
            variables: {
              connections: [connectionId],
              input: { agreementId: agreement.id },
            },
            onCompleted(_response, errors) {
              if (errors) {
                toast.add({
                  title: t("thirdPartyAgreementListItem.messages.error"),
                  description: formatError(
                    t("thirdPartyAgreementListItem.errors.delete"),
                    errors,
                  ),
                  type: "error",
                });
              }
              resolve();
            },
            onError(error) {
              toast.add({
                title: t("thirdPartyAgreementListItem.messages.error"),
                description: formatError(
                  t("thirdPartyAgreementListItem.errors.delete"),
                  error,
                ),
                type: "error",
              });
              resolve();
            },
          });
        }),
      {
        message: t("thirdPartyAgreementListItem.deleteConfirmation", {
          name: agreement.name,
        }),
      },
    );
  };

  const validity = formatValidity(
    agreement.validity?.start,
    agreement.validity?.end,
    i18n.language,
    t,
  );

  return (
    <div className="flex items-center justify-between gap-4 p-4 border border-border-low rounded-lg">
      <div className="flex-1 min-w-0">
        <h3 className="font-medium text-txt-primary">{agreement.name}</h3>
        <p className="text-sm text-txt-secondary mt-1">
          {t(`thirdPartyAgreementsPage.types.${agreement.type}`)}
          {" · "}
          {agreement.file.fileName}
        </p>
        {validity && (
          <p className="text-xs text-txt-secondary mt-1">{validity}</p>
        )}
      </div>
      <div className="flex items-center gap-2 flex-wrap justify-end">
        <Button
          type="button"
          variant="secondary"
          icon={IconPageTextLine}
          onClick={() => setPreviewOpen(true)}
        >
          {t("thirdPartyAgreementListItem.actions.preview")}
        </Button>
        <Button
          type="button"
          variant="secondary"
          onClick={() =>
            downloadFile(agreement.file.downloadUrl, agreement.file.fileName)}
        >
          {t("thirdPartyAgreementListItem.actions.download")}
        </Button>
        {agreement.canUpdate && (
          <EditThirdPartyAgreementDialog agreementKey={agreement}>
            <Button variant="quaternary" icon={IconPencil} />
          </EditThirdPartyAgreementDialog>
        )}
        {agreement.canDelete && (
          <Button variant="quaternary" icon={IconTrashCan} onClick={handleDelete} />
        )}
      </div>
      {previewOpen && (
        <ThirdPartyAgreementPreviewDialog
          agreementKey={agreement}
          onClose={() => setPreviewOpen(false)}
        />
      )}
    </div>
  );
}

function formatValidity(
  validFrom: string | null | undefined,
  validUntil: string | null | undefined,
  language: string,
  t: ReturnType<typeof useTranslation>["t"],
) {
  if (validFrom && validUntil) {
    return t("thirdPartyAgreementsPage.validity.range", {
      from: dateFormat(language, validFrom),
      until: dateFormat(language, validUntil),
    });
  }
  if (validFrom) {
    return t("thirdPartyAgreementsPage.validity.from", {
      date: dateFormat(language, validFrom),
    });
  }
  if (validUntil) {
    return t("thirdPartyAgreementsPage.validity.until", {
      date: dateFormat(language, validUntil),
    });
  }
  return "";
}
