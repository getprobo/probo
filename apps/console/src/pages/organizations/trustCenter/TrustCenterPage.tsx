import { useTranslate } from "@probo/i18n";
import { usePageTitle } from "@probo/hooks";
import {
  Button,
  Card,
  Checkbox,
  Dropzone,
  PageHeader,
  Spinner,
  useToast,
  Tabs,
  TabLink,
  TabItem,
  IconTrashCan,
} from "@probo/ui";
import { usePreloadedQuery, type PreloadedQuery } from "react-relay";
import { trustCenterQuery, useUpdateTrustCenterMutation, useUploadTrustCenterNDAMutation, useDeleteTrustCenterNDAMutation } from "/hooks/graph/TrustCenterGraph";
import type { TrustCenterGraphQuery } from "/hooks/graph/__generated__/TrustCenterGraphQuery.graphql";
import { useState } from "react";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { Outlet, useLocation, Link } from "react-router";
import { TrustCenterReferencesSection } from "/components/trustCenter/TrustCenterReferencesSection";

type Props = {
  queryRef: PreloadedQuery<TrustCenterGraphQuery>;
};

export default function TrustCenterPage({ queryRef }: Props) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const organizationId = useOrganizationId();
  const location = useLocation();
  const { organization } = usePreloadedQuery(trustCenterQuery, queryRef);

  const [updateTrustCenter, isUpdating] = useUpdateTrustCenterMutation();
  const [uploadNDA, isUploadingNDA] = useUploadTrustCenterNDAMutation();
  const [deleteNDA, isDeletingNDA] = useDeleteTrustCenterNDAMutation();
  const [isActive, setIsActive] = useState(organization.trustCenter?.active || false);

  usePageTitle(__("Trust Center"));

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
      <PageHeader
        title={__("Trust Center")}
        description={__(
          "Configure your public trust center to showcase your security and compliance posture."
        )}
      />

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
                {__("Make your trust center publicly accessible to build customer confidence")}
              </p>
            </div>
            <Checkbox
              checked={isActive}
              onChange={handleToggleActive}
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
                  onClick={() => window.open(trustCenterUrl, '_blank', 'noopener,noreferrer')}
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
                {__("Your trust center is currently not accessible to the public. Enable it to start sharing your compliance status.")}
              </p>
            </div>
          )}
        </Card>
      </div>

      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-base font-medium">{__("Non-Disclosure Agreement")}</h2>
          {(isUploadingNDA || isDeletingNDA) && <Spinner />}
        </div>
        <Card padded className="space-y-4">
          <div className="space-y-2">
            {!organization.trustCenter?.ndaFileName ?  (
              <p className="text-sm text-txt-tertiary">
                {__("Upload a Non-Disclosure Agreement that visitors must accept before accessing your trust center")}
              </p>
            ) : (<></>)}
            {organization.trustCenter?.ndaFileName ? (
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <div className="space-y-1">
                    <div className="flex items-center gap-2">
                      <p className="text-sm font-medium">
                        {organization.trustCenter.ndaFileName || __("Non-Disclosure Agreement")}
                      </p>
                    </div>
                    <p className="text-xs text-txt-tertiary">
                      {__("Visitors will need to accept this NDA before accessing your trust center")}
                    </p>
                  </div>
                  <div className="flex items-center gap-2">
                    <Button
                      type="button"
                      variant="secondary"
                      onClick={() => {
                        if (organization.trustCenter?.ndaFileUrl) {
                          window.open(organization.trustCenter.ndaFileUrl, '_blank');
                        }
                      }}
                    >
                      {__("Download PDF")}
                    </Button>
                    <Button
                      variant="quaternary"
                      icon={IconTrashCan}
                      onClick={handleNDADelete}
                      disabled={isDeletingNDA}
                    />
                  </div>
                </div>
              </div>
            ) : (
              <Dropzone
                description={__("Upload PDF files up to 10MB")}
                isUploading={isUploadingNDA}
                onDrop={handleNDAUpload}
                accept={{
                  "application/pdf": [".pdf"],
                }}
                maxSize={10}
              />
            )}
          </div>
        </Card>
      </div>

      {organization.trustCenter?.id && (
        <TrustCenterReferencesSection trustCenterId={organization.trustCenter.id} />
      )}

      <div className="space-y-4">
        <Tabs>
          <TabItem
            asChild
            active={
              location.pathname === `/organizations/${organizationId}/trust-center` ||
              location.pathname === `/organizations/${organizationId}/trust-center/audits`
            }
          >
            <Link to={`/organizations/${organizationId}/trust-center/audits`}>
              {__("Audits")}
            </Link>
          </TabItem>
          <TabLink to={`/organizations/${organizationId}/trust-center/vendors`}>
            {__("Vendors")}
          </TabLink>
          <TabLink to={`/organizations/${organizationId}/trust-center/documents`}>
            {__("Documents")}
          </TabLink>
          <TabLink to={`/organizations/${organizationId}/trust-center/access`}>
            {__("Access")}
          </TabLink>
        </Tabs>

        <Outlet context={{ organization }} />
      </div>
    </div>
  );
}
