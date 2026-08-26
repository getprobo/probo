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

import { WarningCircleIcon } from "@phosphor-icons/react";
import { Callout } from "@probo/ui/src/v2/Callout/Callout";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePortalNDASectionFragment$key } from "#/__generated__/core/CompliancePortalNDASectionFragment.graphql";
import { useUploadCompliancePortalNDAMutation } from "#/hooks/graph/CompliancePortalGraph";

import type { NdaUploadError } from "../_lib/useNdaDropzone";
import { ndaSection } from "../variants";

import { CompliancePortalNDACard } from "./CompliancePortalNDACard";

const fragment = graphql`
  fragment CompliancePortalNDASectionFragment on CompliancePortal {
    id
    nda {
      fileName
    }
    canUploadNDA: permission(action: "compliance-portal:portal:upload-nda")
    ...CompliancePortalNDACard_compliancePortal
  }
`;

export interface CompliancePortalNDASectionProps {
  compliancePortalKey: CompliancePortalNDASectionFragment$key;
}

export function CompliancePortalNDASection({
  compliancePortalKey,
}: CompliancePortalNDASectionProps) {
  const { t } = useTranslation("organizations/compliance-portals");
  const { root, intro, body, grid, errorCopy } = ndaSection();
  const [uploadError, setUploadError] = useState<NdaUploadError | null>(null);

  const compliancePortal = useFragment<CompliancePortalNDASectionFragment$key>(
    fragment,
    compliancePortalKey,
  );

  const [uploadNDA, isUploadingNDA] = useUploadCompliancePortalNDAMutation();

  const handleNDAUpload = (file: File) => {
    setUploadError(null);
    void uploadNDA({
      variables: {
        input: {
          compliancePortalId: compliancePortal.id,
          fileName: file.name,
          file: null,
        },
      },
      uploadables: {
        "input.file": file,
      },
    });
  };

  const hasNDA = compliancePortal.nda?.fileName != null;
  const canUploadNDA = compliancePortal.canUploadNDA;

  return (
    <section className={root()}>
      <div className={intro()}>
        <Heading level={2} size={4} weight="medium" highContrast>
          {t("ndaSection.title")}
        </Heading>
      </div>

      <div className={body()}>
        {uploadError != null && (
          <Callout color="red" icon={<WarningCircleIcon weight="fill" />}>
            <div className={errorCopy()}>
              <Text size={2} weight="medium" color="current" highContrast>
                {t(`ndaSection.errors.${uploadError}.title`)}
              </Text>
              <Text size={2} color="current">
                {t(`ndaSection.errors.${uploadError}.description`)}
              </Text>
            </div>
          </Callout>
        )}

        {hasNDA || canUploadNDA
          ? (
              <div className={grid()}>
                <CompliancePortalNDACard
                  compliancePortalKey={compliancePortal}
                  isUploading={isUploadingNDA}
                  onFile={handleNDAUpload}
                  onReject={setUploadError}
                />
              </div>
            )
          : (
              <Text size={2} color="faint">
                {t("ndaSection.empty")}
              </Text>
            )}
      </div>
    </section>
  );
}
