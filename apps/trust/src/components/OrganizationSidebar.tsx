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
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import { Viewer } from "#/providers/Viewer";
import type { TrustGraphCurrentQuery$data } from "#/queries/__generated__/TrustGraphCurrentQuery.graphql";

import type { OrganizationSidebar_requestAllAccessesMutation } from "./__generated__/OrganizationSidebar_requestAllAccessesMutation.graphql";
import { AuditRowAvatar } from "./AuditRow";

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
  trustCenter: TrustGraphCurrentQuery$data["currentTrustCenter"];
}) {
  const { __ } = useTranslate();
  const isAuthenticated = !!use(Viewer);
  const { toast } = useToast();
  const theme = useSystemTheme();

  const logoFileUrl = theme === "dark" ? (trustCenter?.darkLogoFileUrl ?? trustCenter?.logoFileUrl) : trustCenter?.logoFileUrl;

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

  if (!trustCenter) {
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
      <h1 className="text-2xl mt-6">{trustCenter.organization.name}</h1>
      <p className="text-sm text-txt-secondary mt-1">
        {trustCenter.organization.description}
      </p>

      <hr className="my-6 -mx-6 h-px bg-border-low border-none" />

      {/* Business information */}
      <div className="space-y-4">
        <h2 className="text-xs text-txt-secondary flex gap-1 items-center">
          <IconBlock size={16} />
          {__("Business information")}
        </h2>
        {trustCenter.organization.websiteUrl && (
          <BusinessInfo label={__("Website")}>
            <a
              href={trustCenter.organization.websiteUrl}
              target="_blank"
              rel="noopener noreferrer"
            >
              <span className="text-txt-info hover:underline ">
                {new URL(trustCenter.organization.websiteUrl).host}
              </span>
            </a>
          </BusinessInfo>
        )}
        {trustCenter.organization.email && (
          <BusinessInfo label={__("Contact")}>
            <a href={`mailto:${trustCenter.organization.email}`}>
              <span className="text-txt-info hover:underline ">
                {trustCenter.organization.email}
              </span>
            </a>
          </BusinessInfo>
        )}
        {trustCenter.organization.headquarterAddress && (
          <BusinessInfo label={__("HQ address")}>
            {trustCenter.organization.headquarterAddress}
          </BusinessInfo>
        )}

        <hr className="my-6 -mx-6 h-px bg-border-low border-none" />

        {/* Certifications */}
        {trustCenter.audits.edges.length > 0 && (
          <>
            <div className="space-y-4">
              <h2 className="text-xs text-txt-secondary flex gap-1 items-center">
                <IconMedal size={16} />
                {__("Frameworks")}
              </h2>
              <div
                className="grid grid-cols-4 gap-4"
                style={{
                  gridTemplateColumns: "repeat(auto-fit, 75px",
                }}
              >
                {trustCenter.audits.edges.map(audit => (
                  <AuditRowAvatar key={audit.node.id} audit={audit.node} />
                ))}
              </div>
            </div>

            <hr className="my-6 -mx-6 h-px bg-border-low border-none" />
          </>
        )}

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
