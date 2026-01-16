import { useTranslate } from "@probo/i18n";
import {
  Button,
  Card,
  Checkbox,
  Dropzone,
  Spinner,
  useToast,
  IconTrashCan,
} from "@probo/ui";
import { useOutletContext } from "react-router";
import {
  useUpdateTrustCenterMutation,
  useUploadTrustCenterNDAMutation,
  useDeleteTrustCenterNDAMutation,
} from "/hooks/graph/TrustCenterGraph";
import type { TrustCenterGraphQuery$data } from "/__generated__/core/TrustCenterGraphQuery.graphql";
import { useState } from "react";
import { SlackConnections } from "../../../components/organizations/SlackConnection";

export default function TrustCenterOverviewTab() {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const { organization } = useOutletContext<TrustCenterGraphQuery$data>();

  const [updateTrustCenter, isUpdating] = useUpdateTrustCenterMutation();
  const [uploadNDA, isUploadingNDA] = useUploadTrustCenterNDAMutation();
  const [deleteNDA, isDeletingNDA] = useDeleteTrustCenterNDAMutation();
  const [isActive, setIsActive] = useState(
    organization.trustCenter?.active || false,
  );

  const canUpdateTrustCenter = organization.trustCenter?.canUpdate;

  const handleToggleActive = async (active: boolean) => {
    if (!organization.trustCenter?.id) {
      toast({
        title: __("Error"),
        description: __("Trust center not found"),
        variant: "error",
      });
      return;
    }

    setIsActive(active);

    updateTrustCenter({
      variables: {
        input: {
          trustCenterId: organization.trustCenter.id,
          active,
        },
      },
      onError: () => {
        setIsActive(!active);
      },
    });
  };

  const handleNDAUpload = async (files: File[]) => {
    if (!organization.trustCenter?.id) {
      toast({
        title: __("Error"),
        description: __("Trust center not found"),
        variant: "error",
      });
      return;
    }

    if (files.length === 0) return;

    const file = files[0];

    await uploadNDA({
      variables: {
        input: {
          trustCenterId: organization.trustCenter.id,
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
    if (!organization.trustCenter?.id) {
      toast({
        title: __("Error"),
        description: __("Trust center not found"),
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
          trustCenterId: organization.trustCenter.id,
        },
      },
    });
  };

  const trustCenterUrl = organization.trustCenter?.id
    ? organization.customDomain?.domain
      ? `https://${organization.customDomain.domain}`
      : `${window.location.origin}/trust/${organization.trustCenter.id}`
    : null;

  return (
    <div className="space-y-6">
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-base font-medium">{__("Trust Center Status")}</h2>
          {isUpdating && <Spinner />}
        </div>
        <Card padded className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="space-y-1">
              <h3 className="font-medium">{__("Activate Trust Center")}</h3>
              <p className="text-sm text-txt-tertiary">
                {__(
                  "Make your trust center publicly accessible to build customer confidence",
                )}
              </p>
            </div>
            <Checkbox
              checked={isActive}
              onChange={handleToggleActive}
              disabled={!canUpdateTrustCenter}
            />
          </div>

          {isActive && trustCenterUrl && (
            <div className="mt-4 p-4 bg-accent-light rounded-lg border border-accent">
              <div className="flex items-center justify-between">
                <div>
                  <h4 className="font-medium text-accent-dark">
                    {__("Your Trust Center is Live!")}
                  </h4>
                  <p className="text-sm text-accent-dark mt-1">
                    {__("Your customers can now access your trust center at:")}
                  </p>
                  <a
                    href={trustCenterUrl}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-sm font-mono text-accent underline hover:no-underline"
                  >
                    {trustCenterUrl}
                  </a>
                </div>
                <Button
                  variant="secondary"
                  onClick={() =>
                    window.open(trustCenterUrl, "_blank", "noopener,noreferrer")
                  }
                >
                  {__("View")}
                </Button>
              </div>
            </div>
          )}

          {!isActive && (
            <div className="mt-4 p-4 bg-tertiary rounded-lg border border-border-solid">
              <h4 className="font-medium text-txt-secondary">
                {__("Trust Center is Inactive")}
              </h4>
              <p className="text-sm text-txt-tertiary mt-1">
                {__(
                  "Your trust center is currently not accessible to the public. Enable it to start sharing your compliance status.",
                )}
              </p>
            </div>
          )}
        </Card>
      </div>
      {organization.trustCenter?.canGetNDA && (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-base font-medium">
              {__("Non-Disclosure Agreement")}
            </h2>
            {(isUploadingNDA || isDeletingNDA) && <Spinner />}
          </div>
          <Card padded className="space-y-4">
            <div className="space-y-2">
              {!organization.trustCenter?.ndaFileName &&
              organization.trustCenter.canUploadNDA ? (
                <p className="text-sm text-txt-tertiary">
                  {__(
                    "Upload a Non-Disclosure Agreement that visitors must accept before accessing your trust center",
                  )}
                </p>
              ) : (
                <></>
              )}
              {organization.trustCenter?.ndaFileName ? (
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <div className="space-y-1">
                      <div className="flex items-center gap-2">
                        <p className="text-sm font-medium">
                          {organization.trustCenter.ndaFileName ||
                            __("Non-Disclosure Agreement")}
                        </p>
                      </div>
                      <p className="text-xs text-txt-tertiary">
                        {__(
                          "Visitors will need to accept this NDA before accessing your trust center",
                        )}
                      </p>
                    </div>
                    <div className="flex items-center gap-2">
                      <Button
                        type="button"
                        variant="secondary"
                        onClick={() => {
                          if (organization.trustCenter?.ndaFileUrl) {
                            window.open(
                              organization.trustCenter.ndaFileUrl,
                              "_blank",
                            );
                          }
                        }}
                      >
                        {__("Download PDF")}
                      </Button>
                      {organization.trustCenter?.canDeleteNDA && (
                        <Button
                          variant="quaternary"
                          icon={IconTrashCan}
                          onClick={handleNDADelete}
                          disabled={isDeletingNDA}
                        />
                      )}
                    </div>
                  </div>
                </div>
              ) : (
                <>
                  {canUpdateTrustCenter ? (
                    <Dropzone
                      description={__("Upload PDF files up to 10MB")}
                      isUploading={isUploadingNDA}
                      onDrop={handleNDAUpload}
                      accept={{
                        "application/pdf": [".pdf"],
                      }}
                      maxSize={10}
                    />
                  ) : (
                    <p className="text-sm text-txt-tertiary">
                      {__("No NDA file uploaded")}
                    </p>
                  )}
                </>
              )}
            </div>
          </Card>
        </div>
      )}

      <div className="space-y-4">
        <h2 className="text-base font-medium">{__("Integrations")}</h2>
        <Card padded>
          <SlackConnections
            canUpdate={!!organization.trustCenter?.canUpdate}
            organizationId={organization.id!}
            slackConnections={
              organization.slackConnections?.edges?.map((edge) => edge.node) ??
              []
            }
          />
        </Card>
      </div>
    </div>
  );
}
