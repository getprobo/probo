import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import { PageHeader, TabLink, Tabs } from "@probo/ui";
import { Outlet } from "react-router";

import { useOrganizationId } from "#/hooks/useOrganizationId";

export default function ContextLayout() {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const prefix = `/organizations/${organizationId}/context`;

  usePageTitle(__("Context"));

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Context")}
        description={__(
          "Structured company information and meetings for AI assistants and compliance workflows.",
        )}
      />
      <Tabs>
        <TabLink to={`${prefix}/overview`}>{__("Context")}</TabLink>
        <TabLink to={`${prefix}/meetings`}>{__("Meetings")}</TabLink>
      </Tabs>
      <Outlet />
    </div>
  );
}
