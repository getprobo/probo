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

import { Button, IconChevronRight, useConfirm, useToast } from "@probo/ui";
import { type ChangeEventHandler, useRef } from "react";
import { useTranslation } from "react-i18next";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageNDASectionFragment$key } from "#/__generated__/core/CompliancePageNDASectionFragment.graphql";
import { useDeleteCompliancePageNDAMutation, useUploadCompliancePageNDAMutation } from "#/hooks/graph/CompliancePageGraph";

import { CompliancePageNDACard } from "./CompliancePageNDACard";

const fragment = graphql`
  fragment CompliancePageNDASectionFragment on Organization {
    compliancePage: compliancePortal {
      id
      nda {
        fileName
      }
      canUploadNDA: permission(action: "compliance-portal:portal:upload-nda")
      ...CompliancePageNDACard_compliancePage
    }
  }
`;

export interface CompliancePageNDASectionProps {
  fragmentRef: CompliancePageNDASectionFragment$key;
}

export function CompliancePageNDASection(props: CompliancePageNDASectionProps) {
  const { fragmentRef } = props;

  const { t } = useTranslation("organizations/compliance-page");
  const { toast } = useToast();
  const confirm = useConfirm();
  const fileInputRef = useRef<HTMLInputElement>(null);

  const organization = useFragment<CompliancePageNDASectionFragment$key>(fragment, fragmentRef);

  const [uploadNDA, isUploadingNDA] = useUploadCompliancePageNDAMutation();
  const [deleteNDA, isDeletingNDA] = useDeleteCompliancePageNDAMutation();

  const handleNDAUpload = async (file: File) => {
    if (!organization.compliancePage?.id) {
      toast({
        title: t("ndaSection.errors.title"),
        description: t("ndaSection.errors.notFound"),
        variant: "error",
      });
      return;
    }

    await uploadNDA({
      variables: {
        input: {
          compliancePortalId: organization.compliancePage.id,
          fileName: file.name,
          file: null,
        },
      },
      uploadables: {
        "input.file": file,
      },
    });
  };

  const handleNDAFileChange: ChangeEventHandler<HTMLInputElement> = (e) => {
    const file = e.target.files?.[0];
    e.target.value = "";

    if (!file) return;

    if (file.type !== "application/pdf") {
      toast({
        title: t("ndaSection.errors.invalidFileType.title"),
        description: t("ndaSection.errors.invalidFileType.description"),
        variant: "error",
      });
      return;
    }

    if (file.size > 10 * 1024 * 1024) {
      toast({
        title: t("ndaSection.errors.fileTooLarge.title"),
        description: t("ndaSection.errors.fileTooLarge.description"),
        variant: "error",
      });
      return;
    }

    void handleNDAUpload(file);
  };

  const handleNDADelete = () => {
    const compliancePortalId = organization.compliancePage?.id;
    if (!compliancePortalId) {
      toast({
        title: t("ndaSection.errors.title"),
        description: t("ndaSection.errors.notFound"),
        variant: "error",
      });
      return;
    }

    confirm(
      () => deleteNDA({ variables: { input: { compliancePortalId } } }),
      {
        title: t("ndaSection.delete.title"),
        message: t("ndaSection.delete.description"),
        label: t("ndaSection.delete.actions.delete"),
        variant: "danger",
      },
    );
  };

  const compliancePage = organization.compliancePage;
  const hasNDA = !!compliancePage?.nda?.fileName;
  const canUploadNDA = compliancePage?.canUploadNDA;
  const isBusy = isUploadingNDA || isDeletingNDA;

  return (
    <section className="space-y-4">
      <div>
        <h2 className="text-base font-medium">
          {t("ndaSection.title")}
        </h2>
        <p className="text-sm text-txt-tertiary">
          {t("ndaSection.uploadDescription")}
        </p>
      </div>

      <div className="space-y-3">
        {hasNDA && compliancePage
          ? (
              <CompliancePageNDACard
                compliancePageKey={compliancePage}
                isBusy={isBusy}
                isUploading={isUploadingNDA}
                onFileChange={handleNDAFileChange}
                onDelete={handleNDADelete}
              />
            )
          : canUploadNDA
            ? (
                <div className="flex flex-col items-center justify-center gap-3 rounded-lg border border-dashed border-border-solid px-4 py-8">
                  <p className="max-w-md text-center text-sm text-txt-tertiary">
                    {t("ndaSection.emptyDescription")}
                  </p>
                  <Button
                    iconAfter={IconChevronRight}
                    disabled={isBusy}
                    onClick={() => fileInputRef.current?.click()}
                  >
                    {isUploadingNDA ? t("brandPage.actions.uploading") : t("ndaSection.actions.upload")}
                  </Button>
                  <input
                    ref={fileInputRef}
                    type="file"
                    hidden
                    accept="application/pdf,.pdf"
                    onChange={handleNDAFileChange}
                  />
                </div>
              )
            : (
                <p className="text-sm text-txt-tertiary">
                  {t("ndaSection.empty")}
                </p>
              )}
      </div>
    </section>
  );
}
