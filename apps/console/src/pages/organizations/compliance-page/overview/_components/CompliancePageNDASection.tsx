// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { useTranslate } from "@probo/i18n";
import { Button, Dropzone, IconTrashCan, Spinner, useToast } from "@probo/ui";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { CompliancePageNDASectionFragment$key } from "#/__generated__/core/CompliancePageNDASectionFragment.graphql";
import { useDeleteTrustCenterNDAMutation, useUploadTrustCenterNDAMutation } from "#/hooks/graph/TrustCenterGraph";

const fragment = graphql`
  fragment CompliancePageNDASectionFragment on Organization {
    compliancePage: trustCenter {
      id
      ndaFileName
      ndaFileUrl
      canUploadNDA: permission(action: "core:trust-center:upload-nda")
      canDeleteNDA: permission(action: "core:trust-center:delete-nda")
    }
  }
`;

export function CompliancePageNDASection(props: { fragmentRef: CompliancePageNDASectionFragment$key }) {
  const { fragmentRef } = props;

  const { __ } = useTranslate();
  const { toast } = useToast();

  const organization = useFragment<CompliancePageNDASectionFragment$key>(fragment, fragmentRef);

  const [uploadNDA, isUploadingNDA] = useUploadTrustCenterNDAMutation();
  const [deleteNDA, isDeletingNDA] = useDeleteTrustCenterNDAMutation();

  const handleNDAUpload = async (files: File[]) => {
    if (!organization.compliancePage?.id) {
      toast({
        title: __("Error"),
        description: __("Compliance page not found"),
        variant: "error",
      });
      return;
    }

    if (files.length === 0) return;

    const file = files[0];

    await uploadNDA({
      variables: {
        input: {
          trustCenterId: organization.compliancePage.id,
          fileName: file.name,
          file: null,
        },
      },
      uploadables: {
        "input.file": file,
      },
    });
  };

  const handleNDADelete = async () => {
    if (!organization.compliancePage?.id) {
      toast({
        title: __("Error"),
        description: __("Compliance page not found"),
        variant: "error",
      });
      return;
    }

    if (!confirm(__("Are you sure you want to delete the NDA file?"))) {
      return;
    }

    await deleteNDA({
      variables: {
        input: {
          trustCenterId: organization.compliancePage.id,
        },
      },
    });
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-base font-medium">
          {__("Non-Disclosure Agreement")}
        </h2>
        {(isUploadingNDA || isDeletingNDA) && <Spinner />}
      </div>
      <div className="space-y-2">
        {!organization.compliancePage?.ndaFileName
          && organization.compliancePage?.canUploadNDA
          ? (
              <p className="text-sm text-txt-tertiary">
                {__(
                  "Upload a Non-Disclosure Agreement that visitors must accept before accessing your compliance page",
                )}
              </p>
            )
          : (
              <></>
            )}
        {organization.compliancePage?.ndaFileName
          ? (
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <div className="space-y-1">
                    <div className="flex items-center gap-2">
                      <p className="text-sm font-medium">
                        {organization.compliancePage.ndaFileName
                          || __("Non-Disclosure Agreement")}
                      </p>
                    </div>
                    <p className="text-xs text-txt-tertiary">
                      {__(
                        "Visitors will need to accept this NDA before accessing your compliance page",
                      )}
                    </p>
                  </div>
                  <div className="flex items-center gap-2">
                    <Button
                      type="button"
                      variant="secondary"
                      onClick={() => {
                        if (organization.compliancePage?.ndaFileUrl) {
                          window.open(
                            organization.compliancePage.ndaFileUrl,
                            "_blank",
                            "noopener,noreferrer",
                          );
                        }
                      }}
                    >
                      {__("Download PDF")}
                    </Button>
                    {organization.compliancePage?.canDeleteNDA && (
                      <Button
                        variant="quaternary"
                        icon={IconTrashCan}
                        onClick={() => void handleNDADelete()}
                        disabled={isDeletingNDA}
                      />
                    )}
                  </div>
                </div>
              </div>
            )
          : (
              <>
                {organization.compliancePage?.canUploadNDA
                  ? (
                      <Dropzone
                        description={__("Upload PDF files up to 10MB")}
                        isUploading={isUploadingNDA}
                        onDrop={files => void handleNDAUpload(files)}
                        accept={{
                          "application/pdf": [".pdf"],
                        }}
                        maxSize={10}
                      />
                    )
                  : (
                      <p className="text-sm text-txt-tertiary">
                        {__("No NDA file uploaded")}
                      </p>
                    )}
              </>
            )}
      </div>
    </div>
  );
}
