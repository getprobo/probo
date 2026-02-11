import { useTranslate } from "@probo/i18n";
import {
  IconBank,
  IconBook,
  IconBox,
  IconCalendar1,
  IconCircleProgress,
  IconClock,
  IconCrossLargeX,
  IconFire3,
  IconGroup1,
  IconInboxEmpty,
  IconListStack,
  IconLock,
  IconMedal,
  IconPageCheck,
  IconPageTextLine,
  IconRotateCw,
  IconSettingsGear2,
  IconShield,
  IconStore,
  IconTodo,
  SidebarItem,
} from "@probo/ui";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { SidebarFragment$key } from "#/__generated__/iam/SidebarFragment.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const fragment = graphql`
    fragment SidebarFragment on Organization {
        canListMeetings: permission(action: "core:meeting:list")
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
        canListNonconformities: permission(action: "core:nonconformity:list")
        canListObligations: permission(action: "core:obligation:list")
        canListContinualImprovements: permission(
            action: "core:continual-improvement:list"
        )
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
    <ul className="space-y-[2px]">
      {organization.canListMeetings && (
        <SidebarItem
          label={__("Meetings")}
          icon={IconCalendar1}
          to={`${prefix}/meetings`}
        />
      )}
      {organization.canListTasks && (
        <SidebarItem
          label={__("Tasks")}
          icon={IconInboxEmpty}
          to={`${prefix}/tasks`}
        />
      )}
      {organization.canListMeasures && (
        <SidebarItem
          label={__("Measures")}
          icon={IconTodo}
          to={`${prefix}/measures`}
        />
      )}
      {organization.canListRisks && (
        <SidebarItem
          label={__("Risks")}
          icon={IconFire3}
          to={`${prefix}/risks`}
        />
      )}
      {organization.canListFrameworks && (
        <SidebarItem
          label={__("Frameworks")}
          icon={IconBank}
          to={`${prefix}/frameworks`}
        />
      )}
      {organization.canListMembers && (
        <SidebarItem
          label={__("People")}
          icon={IconGroup1}
          to={`${prefix}/people`}
        />
      )}
      {organization.canListVendors && (
        <SidebarItem
          label={__("Vendors")}
          icon={IconStore}
          to={`${prefix}/vendors`}
        />
      )}
      {organization.canListDocuments && (
        <SidebarItem
          label={__("Documents")}
          icon={IconPageTextLine}
          to={`${prefix}/documents`}
        />
      )}
      {organization.canListAssets && (
        <SidebarItem
          label={__("Assets")}
          icon={IconBox}
          to={`${prefix}/assets`}
        />
      )}
      {organization.canListData && (
        <SidebarItem
          label={__("Data")}
          icon={IconListStack}
          to={`${prefix}/data`}
        />
      )}
      {organization.canListAudits && (
        <SidebarItem
          label={__("Audits")}
          icon={IconMedal}
          to={`${prefix}/audits`}
        />
      )}
      {organization.canListNonconformities && (
        <SidebarItem
          label={__("Nonconformities")}
          icon={IconCrossLargeX}
          to={`${prefix}/nonconformities`}
        />
      )}
      {organization.canListObligations && (
        <SidebarItem
          label={__("Obligations")}
          icon={IconBook}
          to={`${prefix}/obligations`}
        />
      )}
      {organization.canListContinualImprovements && (
        <SidebarItem
          label={__("Continual Improvements")}
          icon={IconRotateCw}
          to={`${prefix}/continual-improvements`}
        />
      )}
      {organization.canListProcessingActivities && (
        <SidebarItem
          label={__("Processing Activities")}
          icon={IconCircleProgress}
          to={`${prefix}/processing-activities`}
        />
      )}
      {organization.canListStatesOfApplicability && (
        <SidebarItem
          label={__("States of Applicability")}
          icon={IconPageCheck}
          to={`${prefix}/states-of-applicability`}
        />
      )}
      {organization.canListRightsRequests && (
        <SidebarItem
          label={__("Rights Requests")}
          icon={IconLock}
          to={`${prefix}/rights-requests`}
        />
      )}
      {organization.canListSnapshots && (
        <SidebarItem
          label={__("Snapshots")}
          icon={IconClock}
          to={`${prefix}/snapshots`}
        />
      )}
      {organization.canGetTrustCenter && (
        <SidebarItem
          label={__("Compliance Page")}
          icon={IconShield}
          to={`${prefix}/compliance-page`}
        />
      )}
      {organization.canUpdateOrganization && (
        <SidebarItem
          label={__("Settings")}
          icon={IconSettingsGear2}
          to={`${prefix}/settings`}
        />
      )}
    </ul>
  );
}
