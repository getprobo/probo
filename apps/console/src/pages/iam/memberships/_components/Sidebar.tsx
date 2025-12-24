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
import { useTranslate } from "@probo/i18n";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { graphql } from "relay-runtime";
import { useFragment } from "react-relay";
import type { SidebarFragment$key } from "./__generated__/SidebarFragment.graphql";

const fragment = graphql`
  fragment SidebarFragment on Identity
  @argumentDefinitions(organizationId: { type: "ID!" }) {
    canListMeetings: permission(
      action: "core:meeting:list"
      id: $organizationId
    )
    canListTasks: permission(action: "core:task:list", id: $organizationId)
    canListMeasures: permission(
      action: "core:measures:list"
      id: $organizationId
    )
    canListRisks: permission(action: "core:risk:list", id: $organizationId)
    canListFrameworks: permission(
      action: "core:frameworks:list"
      id: $organizationId
    )
    canListPeople: permission(action: "core:people:list", id: $organizationId)
    canListVendors: permission(action: "core:vendor:list", id: $organizationId)
    canListDocuments: permission(
      action: "core:document:list"
      id: $organizationId
    )
    canListAssets: permission(action: "core:asset:list", id: $organizationId)
    canListData: permission(action: "core:datum:list", id: $organizationId)
    canListAudits: permission(action: "core:audit:list", id: $organizationId)
    canListNonconformities: permission(
      action: "core:nonconformity:list"
      id: $organizationId
    )
    canListObligations: permission(
      action: "core:obligation:list"
      id: $organizationId
    )
    canListContinualImprovements: permission(
      action: "core:continual-improvement:list"
      id: $organizationId
    )
    canListProcessingActivities: permission(
      action: "core:processing-activity:list"
      id: $organizationId
    )
    canListStatesOfApplicability: permission(
      action: "core:state-of-applicability:list"
      id: $organizationId
    )
    canListSnapshots: permission(
      action: "core:snapshot:list"
      id: $organizationId
    )
    canGetTrustCenter: permission(
      action: "core:trust-center:get"
      id: $organizationId
    )
    canUpdateOrganization: permission(
      action: "iam:organization:update"
      id: $organizationId
    )
  }
`;

export function Sidebar(props: { fKey: SidebarFragment$key }) {
  const { fKey } = props;

  const { __ } = useTranslate();
  const organizationId = useOrganizationId();

  const permissions = useFragment<SidebarFragment$key>(fragment, fKey);

  const prefix = `/organizations/${organizationId}`;

  return (
    <ul className="space-y-[2px]">
      {permissions.canListMeetings && (
        <SidebarItem
          label={__("Meetings")}
          icon={IconCalendar1}
          to={`${prefix}/meetings`}
        />
      )}
      {permissions.canListTasks && (
        <SidebarItem
          label={__("Tasks")}
          icon={IconInboxEmpty}
          to={`${prefix}/tasks`}
        />
      )}
      {permissions.canListMeasures && (
        <SidebarItem
          label={__("Measures")}
          icon={IconTodo}
          to={`${prefix}/measures`}
        />
      )}
      {permissions.canListRisks && (
        <SidebarItem
          label={__("Risks")}
          icon={IconFire3}
          to={`${prefix}/risks`}
        />
      )}
      {permissions.canListFrameworks && (
        <SidebarItem
          label={__("Frameworks")}
          icon={IconBank}
          to={`${prefix}/frameworks`}
        />
      )}
      {permissions.canListPeople && (
        <SidebarItem
          label={__("People")}
          icon={IconGroup1}
          to={`${prefix}/people`}
        />
      )}
      {permissions.canListVendors && (
        <SidebarItem
          label={__("Vendors")}
          icon={IconStore}
          to={`${prefix}/vendors`}
        />
      )}
      {permissions.canListDocuments && (
        <SidebarItem
          label={__("Documents")}
          icon={IconPageTextLine}
          to={`${prefix}/documents`}
        />
      )}
      {permissions.canListAssets && (
        <SidebarItem
          label={__("Assets")}
          icon={IconBox}
          to={`${prefix}/assets`}
        />
      )}
      {permissions.canListData && (
        <SidebarItem
          label={__("Data")}
          icon={IconListStack}
          to={`${prefix}/data`}
        />
      )}
      {permissions.canListAudits && (
        <SidebarItem
          label={__("Audits")}
          icon={IconMedal}
          to={`${prefix}/audits`}
        />
      )}
      {permissions.canListNonconformities && (
        <SidebarItem
          label={__("Nonconformities")}
          icon={IconCrossLargeX}
          to={`${prefix}/nonconformities`}
        />
      )}
      {permissions.canListObligations && (
        <SidebarItem
          label={__("Obligations")}
          icon={IconBook}
          to={`${prefix}/obligations`}
        />
      )}
      {permissions.canListContinualImprovements && (
        <SidebarItem
          label={__("Continual Improvements")}
          icon={IconRotateCw}
          to={`${prefix}/continual-improvements`}
        />
      )}
      {permissions.canListProcessingActivities && (
        <SidebarItem
          label={__("Processing Activities")}
          icon={IconCircleProgress}
          to={`${prefix}/processing-activities`}
        />
      )}
      {permissions.canListStatesOfApplicability && (
        <SidebarItem
          label={__("States of Applicability")}
          icon={IconPageCheck}
          to={`${prefix}/states-of-applicability`}
        />
      )}
      {permissions.canListSnapshots && (
        <SidebarItem
          label={__("Snapshots")}
          icon={IconClock}
          to={`${prefix}/snapshots`}
        />
      )}
      {permissions.canGetTrustCenter && (
        <SidebarItem
          label={__("Trust Center")}
          icon={IconShield}
          to={`${prefix}/trust-center`}
        />
      )}
      {permissions.canUpdateOrganization && (
        <SidebarItem
          label={__("Settings")}
          icon={IconSettingsGear2}
          to={`${prefix}/settings`}
        />
      )}
    </ul>
  );
}
