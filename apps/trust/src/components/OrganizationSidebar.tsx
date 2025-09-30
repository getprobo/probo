import { useTranslate } from "@probo/i18n";
import { Button, Card, IconBlock, IconLock, IconMedal } from "@probo/ui";
import type { TrustGraphQuery$data } from "/queries/__generated__/TrustGraphQuery.graphql";
import type { PropsWithChildren } from "react";
import { domain } from "@probo/helpers";
import { AuditRowAvatar } from "./AuditRow";
import { RequestAccessDialog } from "./RequestAccessDialog";
import { useIsAuthenticated } from "/hooks/useIsAuthenticated";

export function OrganizationSidebar({
  trustCenter,
}: {
  trustCenter: TrustGraphQuery$data["trustCenterBySlug"];
}) {
  const { __ } = useTranslate();
  const isAuthenticated = useIsAuthenticated();

  if (!trustCenter) {
    return null;
  }

  return (
    <Card className="p-6 relative overflow-hidden border-b-1 border-border-low isolate">
      <div className="h-21 bg-[#044E4114] absolute top-0 left-0 right-0 -z-1"></div>
      {trustCenter.organization.logoUrl ? (
        <img
          alt=""
          src={trustCenter.organization.logoUrl}
          className="size-24 rounded-2xl border border-border-mid shadow-mid"
        />
      ) : (
        <div className="size-24 rounded-2xl border border-border-mid bg-level-1 shadow-mid" />
      )}
      <h1 className="text-2xl mt-6">{trustCenter.organization.name}</h1>
      <p className="text-sm text-txt-secondary mt-1">
        {trustCenter.organization.description}
      </p>

      <hr className="my-6 -mx-6 h-[1px] bg-border-low border-none" />

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
                {domain(trustCenter.organization.websiteUrl)}
              </span>
            </a>
          </BusinessInfo>
        )}
        {trustCenter.organization.email && (
          <BusinessInfo label={__("Contact")}>
            {trustCenter.organization.email}
          </BusinessInfo>
        )}
        {trustCenter.organization.headquarterAddress && (
          <BusinessInfo label={__("HQ address")}>
            {trustCenter.organization.headquarterAddress}
          </BusinessInfo>
        )}

        <hr className="my-6 -mx-6 h-[1px] bg-border-low border-none" />

        {/* Certifications */}
        <div className="space-y-4">
          <h2 className="text-xs text-txt-secondary flex gap-1 items-center">
            <IconMedal size={16} />
            {__("Certifications")}
          </h2>
          <div
            className="grid grid-cols-4 gap-4"
            style={{
              gridTemplateColumns: "repeat(auto-fit, 75px",
            }}
          >
            {trustCenter.audits.edges.map((audit) => (
              <AuditRowAvatar key={audit.node.id} audit={audit.node} />
            ))}
          </div>
        </div>

        <hr className="my-6 -mx-6 h-[1px] bg-border-low border-none" />

        {/* Actions */}
        {!isAuthenticated && (
          <RequestAccessDialog>
            <Button variant="primary" icon={IconLock} className="w-full h-10">
              {__("Request access")}
            </Button>
          </RequestAccessDialog>
        )}
        {/* <Button variant="secondary" icon={IconMail} className="w-full h-10">
          {__("Subscribe to updates")}
        </Button>*/}
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
