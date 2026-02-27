import { formatError } from "@probo/helpers";
import { useSystemTheme } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  Button,
  Card,
  IconBlock,
  IconLock,
  IconMedal,
  useToast,
} from "@probo/ui";
import { type PropsWithChildren, use } from "react";
import { useFragment, useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import { Viewer } from "#/providers/Viewer";

import type { OrganizationSidebar_requestAllAccessesMutation } from "./__generated__/OrganizationSidebar_requestAllAccessesMutation.graphql";
import type { OrganizationSidebarFragment$key } from "./__generated__/OrganizationSidebarFragment.graphql";
import { AuditRowAvatar } from "./AuditRow";

const sidebarFragment = graphql`
  fragment OrganizationSidebarFragment on TrustCenter {
    logoFileUrl
    darkLogoFileUrl
    organization {
      name
      description
      websiteUrl
      email
      headquarterAddress
    }
    complianceBadges(first: 50) {
      edges {
        node {
          id
          name
          iconUrl
        }
      }
    }
    audits(first: 50) {
      edges {
        node {
          id
          ...AuditRowFragment
        }
      }
    }
  }
`;

const requestAllAccessesMutation = graphql`
  mutation OrganizationSidebar_requestAllAccessesMutation {
    requestAllAccesses {
      trustCenterAccess {
        id
      }
    }
  }
`;

export function OrganizationSidebar({
  trustCenter,
}: {
  trustCenter: OrganizationSidebarFragment$key | null | undefined;
}) {
  const data = useFragment(sidebarFragment, trustCenter ?? null);
  const { __ } = useTranslate();
  const isAuthenticated = !!use(Viewer);
  const { toast } = useToast();
  const theme = useSystemTheme();

  const logoFileUrl = theme === "dark" ? (data?.darkLogoFileUrl ?? data?.logoFileUrl) : data?.logoFileUrl;

  const [requestAllAccesses, isRequestingAccess]
    = useMutation<OrganizationSidebar_requestAllAccessesMutation>(
      requestAllAccessesMutation,
    );

  const handleRequestAllAccesses = () => {
    requestAllAccesses({
      variables: {},
      onCompleted: (_, errors) => {
        if (errors?.length) {
          toast({
            title: __("Error"),
            description: formatError(__("Cannot request access"), errors),
            variant: "error",
          });
          return;
        }
        toast({
          title: __("Success"),
          description: __("Access request submitted successfully."),
          variant: "success",
        });
      },
      onError: (error) => {
        toast({
          title: __("Error"),
          description: error.message ?? __("Cannot request access"),
          variant: "error",
        });
      },
    });
  };

  if (!data) {
    return null;
  }

  return (
    <Card className="p-6 relative overflow-hidden border-b border-border-low isolate">
      <div className="h-21 bg-[#044E4114] absolute top-0 left-0 right-0 -z-1"></div>
      {logoFileUrl
        ? (
            <img
              alt=""
              src={logoFileUrl}
              className="size-24 rounded-2xl border border-border-mid shadow-mid bg-level-1"
            />
          )
        : (
            <div className="size-24 rounded-2xl border border-border-mid shadow-mid bg-level-1" />
          )}
      <h1 className="text-2xl mt-6">{data.organization.name}</h1>
      <p className="text-sm text-txt-secondary mt-1">
        {data.organization.description}
      </p>

      <hr className="my-6 -mx-6 h-px bg-border-low border-none" />

      {/* Business information */}
      <div className="space-y-4">
        <h2 className="text-xs text-txt-secondary flex gap-1 items-center">
          <IconBlock size={16} />
          {__("Business information")}
        </h2>
        {data.organization.websiteUrl && (
          <BusinessInfo label={__("Website")}>
            <a
              href={data.organization.websiteUrl}
              target="_blank"
              rel="noopener noreferrer"
            >
              <span className="text-txt-info hover:underline ">
                {new URL(data.organization.websiteUrl).host}
              </span>
            </a>
          </BusinessInfo>
        )}
        {data.organization.email && (
          <BusinessInfo label={__("Contact")}>
            <a href={`mailto:${data.organization.email}`}>
              <span className="text-txt-info hover:underline ">
                {data.organization.email}
              </span>
            </a>
          </BusinessInfo>
        )}
        {data.organization.headquarterAddress && (
          <BusinessInfo label={__("HQ address")}>
            {data.organization.headquarterAddress}
          </BusinessInfo>
        )}

        <hr className="my-6 -mx-6 h-px bg-border-low border-none" />

        {/* Badges or Frameworks */}
        {data.complianceBadges.edges.length > 0
          ? (
              <>
                <div className="space-y-4">
                  <h2 className="text-xs text-txt-secondary flex gap-1 items-center">
                    <IconMedal size={16} />
                    {__("Frameworks")}
                  </h2>
                  <div className="flex flex-wrap gap-3">
                    {data.complianceBadges.edges.map(({ node: badge }) => (
                      <div key={badge.id} className="flex flex-col items-center gap-1.5" title={badge.name}>
                        <img
                          src={badge.iconUrl}
                          alt={badge.name}
                          className="size-10 object-contain"
                        />
                        <span className="text-xs text-txt-secondary text-center leading-tight max-w-[72px] truncate">
                          {badge.name}
                        </span>
                      </div>
                    ))}
                  </div>
                </div>
                <hr className="my-6 -mx-6 h-px bg-border-low border-none" />
              </>
            )
          : data.audits.edges.length > 0
            ? (
                <>
                  <div className="space-y-4">
                    <h2 className="text-xs text-txt-secondary flex gap-1 items-center">
                      <IconMedal size={16} />
                      {__("Frameworks")}
                    </h2>
                    <div
                      className="grid grid-cols-4 gap-4"
                      style={{
                        gridTemplateColumns: "repeat(auto-fit, 75px)",
                      }}
                    >
                      {data.audits.edges.map(audit => (
                        <AuditRowAvatar key={audit.node.id} audit={audit.node} />
                      ))}
                    </div>
                  </div>
                  <hr className="my-6 -mx-6 h-px bg-border-low border-none" />
                </>
              )
            : null}

        {/* Actions */}
        {isAuthenticated
          ? (
              <Button
                disabled={isRequestingAccess}
                variant="primary"
                icon={IconLock}
                className="w-full h-10"
                onClick={handleRequestAllAccesses}
              >
                {__("Request access")}
              </Button>
            )
          : (
              <Button
                variant="primary"
                icon={IconLock}
                className="w-full h-10"
                to="/connect"
              >
                {__("Request access")}
              </Button>
            )}
      </div>
    </Card>
  );
}

function BusinessInfo({
  children,
  label,
}: PropsWithChildren<{ label: string }>) {
  return (
    <div>
      <div className="text-xs text-txt-secondary">{label}</div>
      <div className="text-sm">{children}</div>
    </div>
  );
}
