import { useTranslate } from "@probo/i18n";
import { PageHeader, TabLink, Tabs } from "@probo/ui";
import { Outlet } from "react-router";

export default function EmployeeTabsLayout() {
  const { __ } = useTranslate();

  return (
    <div className="space-y-6">
      <PageHeader title={__("Documents")} />
      <Tabs>
        <TabLink to="signatures" end>
          {__("Signatures")}
        </TabLink>
        <TabLink to="approvals" end>
          {__("Approvals")}
        </TabLink>
      </Tabs>
      <Outlet />
    </div>
  );
}
