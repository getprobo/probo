import { useTranslate } from "@probo/i18n";
import { usePageTitle } from "@probo/hooks";
import {
  Button,
  Card,
  Checkbox,
  Field,
  Input,
  PageHeader,
  Spinner,
  useToast,
  Tabs,
  TabLink,
  TabItem,
} from "@probo/ui";
import { usePreloadedQuery, type PreloadedQuery } from "react-relay";
import { trustCenterQuery, useUpdateTrustCenterMutation } from "/hooks/graph/TrustCenterGraph";
import type { TrustCenterGraphQuery } from "/hooks/graph/__generated__/TrustCenterGraphQuery.graphql";
import { useState } from "react";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { Outlet, useLocation, Link } from "react-router";

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
  const [isActive, setIsActive] = useState(organization.trustCenter?.active || false);
  const [slug, setSlug] = useState(organization.trustCenter?.slug || "");
  const [isUpdatingSlug, setIsUpdatingSlug] = useState(false);

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

  const handleSlugUpdate = async () => {
    if (!organization.trustCenter?.id) {
      toast({
        title: __("Error"),
        description: __("Trust center not found"),
        variant: "error",
      });
      return;
    }

    if (!slug.trim()) {
      toast({
        title: __("Error"),
        description: __("Slug cannot be empty"),
        variant: "error",
      });
      return;
    }

    setIsUpdatingSlug(true);

    updateTrustCenter({
      variables: {
        input: {
          trustCenterId: organization.trustCenter.id,
          slug: slug.trim(),
        },
      },
      onCompleted: () => {
        setIsUpdatingSlug(false);
      },
      onError: () => {
        setIsUpdatingSlug(false);
        setSlug(organization.trustCenter?.slug || "");
      },
    });
  };

  const trustCenterUrl = organization.trustCenter?.slug
    ? `${window.location.origin}/trust/${organization.trustCenter.slug}`
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
          <h2 className="text-base font-medium">{__("Configuration")}</h2>
          {isUpdatingSlug && <Spinner />}
        </div>
        <Card padded className="space-y-4">
          <div className="space-y-2">
            <Field
              label={__("Slug")}
              help={__("The unique identifier for your trust center URL")}
            >
              <div className="flex items-center space-x-2">
                <div className="flex items-center">
                  <Input
                    value={slug}
                    onChange={(e) => setSlug(e.target.value)}
                    placeholder={__("your-organization")}
                    className="min-w-[200px]"
                  />
                </div>
                <Button
                  variant="secondary"
                  onClick={handleSlugUpdate}
                  disabled={
                    isUpdatingSlug ||
                    !slug.trim() ||
                    slug === organization.trustCenter?.slug
                  }
                >
                  {__("Update")}
                </Button>
              </div>
            </Field>
          </div>
        </Card>
      </div>
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
