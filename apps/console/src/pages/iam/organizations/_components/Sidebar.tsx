import { useTranslate } from "@probo/i18n";
import {
  IconBank,
  IconBook,
  IconBox,
  IconCircleProgress,
  IconClock,
  IconFire3,
  IconGroup1,
  IconHome,
  IconInboxEmpty,
  IconListStack,
  IconLock,
  IconMagnifyingGlass,
  IconMedal,
  IconPageCheck,
  IconPageTextLine,
  IconPageTextSolid,
  IconSettingsGear2,
  IconShield,
  IconStore,
  IconTodo,
  SidebarGroup,
  SidebarItem,
} from "@probo/ui";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { SidebarFragment$key } from "#/__generated__/iam/SidebarFragment.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const fragment = graphql`
    fragment SidebarFragment on Organization {
        canGetContext: permission(action: "core:organization-context:get")
        canListTasks: permission(action: "core:task:list")
        canListMeasures: permission(action: "core:measure:list")
        canListRisks: permission(action: "core:risk:list")
        canListFrameworks: permission(action: "core:framework:list")
        canListMembers: permission(action: "iam:membership:list")
        canListVendors: permission(action: "core:vendor:list")
        canListDocuments: permission(action: "core:document:list")
        canListAssets: permission(action: "core:asset:list")
        canListData: permission(action: "core:datum:list")
        canListAudits: permission(action: "core:audit:list")
        canListFindings: permission(action: "core:finding:list")
        canListObligations: permission(action: "core:obligation:list")
        canListProcessingActivities: permission(
            action: "core:processing-activity:list"
        )
        canListRightsRequests: permission(action: "core:rights-request:list")
        canListSnapshots: permission(action: "core:snapshot:list")
        canGetTrustCenter: permission(action: "core:trust-center:get")
        canUpdateOrganization: permission(action: "iam:organization:update")
        canListStatesOfApplicability: permission(
            action: "core:state-of-applicability:list"
        )
    }
`;

export function Sidebar(props: { fKey: SidebarFragment$key }) {
  const { fKey } = props;

  const { __ } = useTranslate();
  const organizationId = useOrganizationId();

  const organization = useFragment<SidebarFragment$key>(fragment, fKey);

  const prefix = `/organizations/${organizationId}`;

  return (
    <div className="space-y-1">
      {/* Overview - standalone top item */}
      <ul className="space-y-0.5">
        <SidebarItem
          label={__("Overview")}
          icon={IconHome}
          to={`${prefix}/tasks`}
        />
      </ul>

      {/* Work */}
      {(organization.canListTasks || organization.canGetContext) && (
        <SidebarGroup label={__("Work")}>
          {organization.canListTasks && (
            <SidebarItem
              label={__("Tasks")}
              icon={IconInboxEmpty}
              to={`${prefix}/tasks`}
            />
          )}
          {organization.canGetContext && (
            <SidebarItem
              label={__("Context")}
              icon={IconPageTextSolid}
              to={`${prefix}/context`}
            />
          )}
        </SidebarGroup>
      )}

      {/* Governance */}
      {(organization.canListFrameworks
        || organization.canListMeasures
        || organization.canListAudits
        || organization.canListStatesOfApplicability
        || organization.canGetTrustCenter) && (
        <SidebarGroup label={__("Governance")}>
          {organization.canGetTrustCenter && (
            <SidebarItem
              label={__("Compliance")}
              icon={IconShield}
              to={`${prefix}/compliance-page`}
            />
          )}
          {organization.canListFrameworks && (
            <SidebarItem
              label={__("Frameworks")}
              icon={IconBank}
              to={`${prefix}/frameworks`}
            />
          )}
          {organization.canListMeasures && (
            <SidebarItem
              label={__("Measures")}
              icon={IconTodo}
              to={`${prefix}/measures`}
            />
          )}
          {organization.canListAudits && (
            <SidebarItem
              label={__("Audits")}
              icon={IconMedal}
              to={`${prefix}/audits`}
            />
          )}
          {organization.canListStatesOfApplicability && (
            <SidebarItem
              label={__("States of Applicability")}
              icon={IconPageCheck}
              to={`${prefix}/states-of-applicability`}
            />
          )}
        </SidebarGroup>
      )}

      {/* Risk Management */}
      {(organization.canListRisks || organization.canListObligations) && (
        <SidebarGroup label={__("Risk Management")}>
          {organization.canListRisks && (
            <SidebarItem
              label={__("Risks")}
              icon={IconFire3}
              to={`${prefix}/risks`}
            />
          )}
          {organization.canListObligations && (
            <SidebarItem
              label={__("Obligations")}
              icon={IconBook}
              to={`${prefix}/obligations`}
            />
          )}
        </SidebarGroup>
      )}

      {/* Asset Management */}
      {(organization.canListAssets || organization.canListRightsRequests) && (
        <SidebarGroup label={__("Asset Management")}>
          {organization.canListAssets && (
            <SidebarItem
              label={__("Assets")}
              icon={IconBox}
              to={`${prefix}/assets`}
            />
          )}
          {organization.canListRightsRequests && (
            <SidebarItem
              label={__("Rights Requests")}
              icon={IconLock}
              to={`${prefix}/rights-requests`}
            />
          )}
        </SidebarGroup>
      )}

      {/* Company */}
      {(organization.canListVendors || organization.canListMembers || organization.canListFindings) && (
        <SidebarGroup label={__("Company")}>
          {organization.canListVendors && (
            <SidebarItem
              label={__("Vendors")}
              icon={IconStore}
              to={`${prefix}/vendors`}
            />
          )}
          {organization.canListMembers && (
            <SidebarItem
              label={__("People")}
              icon={IconGroup1}
              to={`${prefix}/people`}
            />
          )}
          {organization.canListFindings && (
            <SidebarItem
              label={__("Findings")}
              icon={IconMagnifyingGlass}
              to={`${prefix}/findings`}
            />
          )}
        </SidebarGroup>
      )}

      {/* Evidence */}
      {(organization.canListDocuments || organization.canListSnapshots) && (
        <SidebarGroup label={__("Evidence")}>
          {organization.canListDocuments && (
            <SidebarItem
              label={__("Documents")}
              icon={IconPageTextLine}
              to={`${prefix}/documents`}
            />
          )}
          {organization.canListSnapshots && (
            <SidebarItem
              label={__("Snapshots")}
              icon={IconClock}
              to={`${prefix}/snapshots`}
            />
          )}
        </SidebarGroup>
      )}

      {/* Data Management */}
      {(organization.canListData || organization.canListProcessingActivities) && (
        <SidebarGroup label={__("Data Management")}>
          {organization.canListData && (
            <SidebarItem
              label={__("Data")}
              icon={IconListStack}
              to={`${prefix}/data`}
            />
          )}
          {organization.canListProcessingActivities && (
            <SidebarItem
              label={__("Processing Activities")}
              icon={IconCircleProgress}
              to={`${prefix}/processing-activities`}
            />
          )}
        </SidebarGroup>
      )}

      {/* Settings - standalone bottom item */}
      {organization.canUpdateOrganization && (
        <ul className="space-y-0.5 mt-3">
          <SidebarItem
            label={__("Settings")}
            icon={IconSettingsGear2}
            to={`${prefix}/settings`}
          />
        </ul>
      )}
    </div>
  );
}
