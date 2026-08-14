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

import { downloadFile } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { dateFormat } from "@probo/i18n";
import {
  Button,
  Card,
  IconPencil,
  IconPlusLarge,
  IconTrashCan,
  PageHeader,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, useFragment, usePreloadedQuery } from "react-relay";

import type { ThirdPartyAgreementsPageBusinessAssociateAgreementFragment$key } from "#/__generated__/core/ThirdPartyAgreementsPageBusinessAssociateAgreementFragment.graphql";
import type { ThirdPartyAgreementsPageDataPrivacyAgreementFragment$key } from "#/__generated__/core/ThirdPartyAgreementsPageDataPrivacyAgreementFragment.graphql";
import type { ThirdPartyAgreementsPageQuery } from "#/__generated__/core/ThirdPartyAgreementsPageQuery.graphql";

import { DeleteBusinessAssociateAgreementDialog } from "../_components/DeleteBusinessAssociateAgreementDialog";
import { DeleteDataPrivacyAgreementDialog } from "../_components/DeleteDataPrivacyAgreementDialog";
import { EditBusinessAssociateAgreementDialog } from "../_components/EditBusinessAssociateAgreementDialog";
import { EditDataPrivacyAgreementDialog } from "../_components/EditDataPrivacyAgreementDialog";
import { UploadBusinessAssociateAgreementDialog } from "../_components/UploadBusinessAssociateAgreementDialog";
import { UploadDataPrivacyAgreementDialog } from "../_components/UploadDataPrivacyAgreementDialog";

const thirdPartyBusinessAssociateAgreementFragment = graphql`
  fragment ThirdPartyAgreementsPageBusinessAssociateAgreementFragment on ThirdParty {
    businessAssociateAgreement {
      id
      file {
        fileName
        downloadUrl
      }
      validFrom
      validUntil
      canUpdate: permission(
        action: "core:thirdParty-business-associate-agreement:update"
      )
      canDelete: permission(
        action: "core:thirdParty-business-associate-agreement:delete"
      )
    }
  }
`;

const thirdPartyDataPrivacyAgreementFragment = graphql`
  fragment ThirdPartyAgreementsPageDataPrivacyAgreementFragment on ThirdParty {
    dataPrivacyAgreement {
      id
      file {
        fileName
        downloadUrl
      }
      validFrom
      validUntil
      canUpdate: permission(action: "core:thirdParty-data-privacy-agreement:update")
      canDelete: permission(action: "core:thirdParty-data-privacy-agreement:delete")
    }
  }
`;

export const thirdPartyAgreementsPageQuery = graphql`
  query ThirdPartyAgreementsPageQuery($thirdPartyId: ID!) {
    node(id: $thirdPartyId) {
      __typename
      ... on ThirdParty {
        id
        name
        canUploadBAA: permission(
          action: "core:thirdParty-business-associate-agreement:upload"
        )
        canUploadDPA: permission(
          action: "core:thirdParty-data-privacy-agreement:upload"
        )
        ...ThirdPartyAgreementsPageBusinessAssociateAgreementFragment
        ...ThirdPartyAgreementsPageDataPrivacyAgreementFragment
      }
    }
  }
`;

interface ThirdPartyAgreementsPageProps {
  queryRef: PreloadedQuery<ThirdPartyAgreementsPageQuery>;
}

export function ThirdPartyAgreementsPage({ queryRef }: ThirdPartyAgreementsPageProps) {
  const data = usePreloadedQuery<ThirdPartyAgreementsPageQuery>(
    thirdPartyAgreementsPageQuery,
    queryRef,
  );
  if (data.node?.__typename !== "ThirdParty") {
    throw new Error("Third party not found");
  }
  const thirdParty = data.node;
  const { t, i18n } = useTranslation();

  const thirdPartyWithBAA
    = useFragment<ThirdPartyAgreementsPageBusinessAssociateAgreementFragment$key>(
      thirdPartyBusinessAssociateAgreementFragment,
      thirdParty,
    );
  const businessAssociateAgreement = thirdPartyWithBAA.businessAssociateAgreement;

  const thirdPartyWithDPA
    = useFragment<ThirdPartyAgreementsPageDataPrivacyAgreementFragment$key>(
      thirdPartyDataPrivacyAgreementFragment,
      thirdParty,
    );
  const dataPrivacyAgreement = thirdPartyWithDPA.dataPrivacyAgreement;

  usePageTitle(t("thirdPartyAgreementsPage.pageTitle", { name: thirdParty.name }));

  return (
    <div className="space-y-6">
      <PageHeader
        title={t("thirdPartyAgreementsPage.title")}
        description={t("thirdPartyAgreementsPage.description")}
      />
      <Card className="space-y-4" padded>
        <div className="flex items-center justify-between p-4 border border-border-low rounded-lg">
          <div className="flex-1">
            <h3 className="font-medium text-txt-primary">
              {t("thirdPartyAgreementsPage.agreements.businessAssociate")}
            </h3>
            <p className="text-sm text-txt-secondary mt-1">
              {businessAssociateAgreement
                ? businessAssociateAgreement.file.fileName
                : t("thirdPartyAgreementsPage.agreements.noBusinessAssociate")}
            </p>
            {(businessAssociateAgreement?.validFrom
              || businessAssociateAgreement?.validUntil) && (
              <p className="text-xs text-txt-secondary mt-1">
                {formatValidity(
                  businessAssociateAgreement.validFrom,
                  businessAssociateAgreement.validUntil,
                  i18n.language,
                  t,
                )}
              </p>
            )}
          </div>
          <div className="flex items-center gap-2">
            {businessAssociateAgreement
              ? (
                  <>
                    <Button
                      type="button"
                      variant="secondary"
                      onClick={() =>
                        downloadFile(
                          businessAssociateAgreement.file.downloadUrl,
                          businessAssociateAgreement.file.fileName,
                        )}
                    >
                      {t("thirdPartyAgreementsPage.actions.downloadPdf")}
                    </Button>
                    {businessAssociateAgreement.canUpdate && (
                      <EditBusinessAssociateAgreementDialog
                        thirdPartyId={thirdParty.id}
                        agreement={{
                          validFrom: businessAssociateAgreement.validFrom,
                          validUntil: businessAssociateAgreement.validUntil,
                        }}
                        onSuccess={() => window.location.reload()}
                      >
                        <Button variant="quaternary" icon={IconPencil} />
                      </EditBusinessAssociateAgreementDialog>
                    )}
                    {businessAssociateAgreement.canDelete && (
                      <DeleteBusinessAssociateAgreementDialog
                        thirdPartyId={thirdParty.id}
                        fileName={businessAssociateAgreement.file.fileName}
                        onSuccess={() => window.location.reload()}
                      >
                        <Button variant="quaternary" icon={IconTrashCan} />
                      </DeleteBusinessAssociateAgreementDialog>
                    )}
                  </>
                )
              : (
                  thirdParty.canUploadBAA && (
                    <UploadBusinessAssociateAgreementDialog
                      thirdPartyId={thirdParty.id}
                      onSuccess={() => window.location.reload()}
                    >
                      <Button variant="secondary" icon={IconPlusLarge}>
                        {t("thirdPartyAgreementsPage.actions.upload")}
                      </Button>
                    </UploadBusinessAssociateAgreementDialog>
                  )
                )}
          </div>
        </div>

        <div className="flex items-center justify-between p-4 border border-border-low rounded-lg">
          <div className="flex-1">
            <h3 className="font-medium text-txt-primary">
              {t("thirdPartyAgreementsPage.agreements.dataPrivacy")}
            </h3>
            <p className="text-sm text-txt-secondary mt-1">
              {dataPrivacyAgreement
                ? dataPrivacyAgreement.file.fileName
                : t("thirdPartyAgreementsPage.agreements.noDataPrivacy")}
            </p>
            {(dataPrivacyAgreement?.validFrom
              || dataPrivacyAgreement?.validUntil) && (
              <p className="text-xs text-txt-secondary mt-1">
                {formatValidity(
                  dataPrivacyAgreement.validFrom,
                  dataPrivacyAgreement.validUntil,
                  i18n.language,
                  t,
                )}
              </p>
            )}
          </div>
          <div className="flex items-center gap-2">
            {dataPrivacyAgreement
              ? (
                  <>
                    <Button
                      type="button"
                      variant="secondary"
                      onClick={() =>
                        downloadFile(
                          dataPrivacyAgreement.file.downloadUrl,
                          dataPrivacyAgreement.file.fileName,
                        )}
                    >
                      {t("thirdPartyAgreementsPage.actions.downloadPdf")}
                    </Button>
                    {dataPrivacyAgreement.canUpdate && (
                      <EditDataPrivacyAgreementDialog
                        thirdPartyId={thirdParty.id}
                        agreement={{
                          validFrom: dataPrivacyAgreement.validFrom,
                          validUntil: dataPrivacyAgreement.validUntil,
                        }}
                        onSuccess={() => window.location.reload()}
                      >
                        <Button variant="quaternary" icon={IconPencil} />
                      </EditDataPrivacyAgreementDialog>
                    )}
                    {dataPrivacyAgreement.canDelete && (
                      <DeleteDataPrivacyAgreementDialog
                        thirdPartyId={thirdParty.id}
                        fileName={dataPrivacyAgreement.file.fileName}
                        onSuccess={() => window.location.reload()}
                      >
                        <Button variant="quaternary" icon={IconTrashCan} />
                      </DeleteDataPrivacyAgreementDialog>
                    )}
                  </>
                )
              : (
                  thirdParty.canUploadDPA && (
                    <UploadDataPrivacyAgreementDialog
                      thirdPartyId={thirdParty.id}
                      onSuccess={() => window.location.reload()}
                    >
                      <Button variant="secondary" icon={IconPlusLarge}>
                        {t("thirdPartyAgreementsPage.actions.upload")}
                      </Button>
                    </UploadDataPrivacyAgreementDialog>
                  )
                )}
          </div>
        </div>
      </Card>
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
    return t("thirdPartyAgreementsPage.agreements.validity.range", {
      from: dateFormat(language, validFrom),
      until: dateFormat(language, validUntil),
    });
  }
  if (validFrom) {
    return t("thirdPartyAgreementsPage.agreements.validity.from", { date: dateFormat(language, validFrom) });
  }
  if (validUntil) {
    return t("thirdPartyAgreementsPage.agreements.validity.until", { date: dateFormat(language, validUntil) });
  }
  return "";
}
