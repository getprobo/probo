import { use } from "react";
import { PermissionsContext } from "/providers/PermissionsContext";
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

export function Sidebar() {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();

  const { isAuthorized } = use(PermissionsContext);

  const prefix = `/organizations/${organizationId}`;

  console.log(isAuthorized("Organization", "listMeetings"));
  return (
    <ul className="space-y-[2px]">
      {isAuthorized("Organization", "listMeetings") && (
        <SidebarItem
          label={__("Meetings")}
          icon={IconCalendar1}
          to={`${prefix}/meetings`}
        />
      )}
      {isAuthorized("Organization", "listTasks") && (
        <SidebarItem
          label={__("Tasks")}
          icon={IconInboxEmpty}
          to={`${prefix}/tasks`}
        />
      )}
      {isAuthorized("Organization", "listMeasures") && (
        <SidebarItem
          label={__("Measures")}
          icon={IconTodo}
          to={`${prefix}/measures`}
        />
      )}
      {isAuthorized("Organization", "listRisks") && (
        <SidebarItem
          label={__("Risks")}
          icon={IconFire3}
          to={`${prefix}/risks`}
        />
      )}
      {isAuthorized("Organization", "listFrameworks") && (
        <SidebarItem
          label={__("Frameworks")}
          icon={IconBank}
          to={`${prefix}/frameworks`}
        />
      )}
      {isAuthorized("Organization", "listPeople") && (
        <SidebarItem
          label={__("People")}
          icon={IconGroup1}
          to={`${prefix}/people`}
        />
      )}
      {isAuthorized("Organization", "listVendors") && (
        <SidebarItem
          label={__("Vendors")}
          icon={IconStore}
          to={`${prefix}/vendors`}
        />
      )}
      {isAuthorized("Organization", "listDocuments") && (
        <SidebarItem
          label={__("Documents")}
          icon={IconPageTextLine}
          to={`${prefix}/documents`}
        />
      )}
      {isAuthorized("Organization", "listAssets") && (
        <SidebarItem
          label={__("Assets")}
          icon={IconBox}
          to={`${prefix}/assets`}
        />
      )}
      {isAuthorized("Organization", "listData") && (
        <SidebarItem
          label={__("Data")}
          icon={IconListStack}
          to={`${prefix}/data`}
        />
      )}
      {isAuthorized("Organization", "listAudits") && (
        <SidebarItem
          label={__("Audits")}
          icon={IconMedal}
          to={`${prefix}/audits`}
        />
      )}
      {isAuthorized("Organization", "listNonconformities") && (
        <SidebarItem
          label={__("Nonconformities")}
          icon={IconCrossLargeX}
          to={`${prefix}/nonconformities`}
        />
      )}
      {isAuthorized("Organization", "listObligations") && (
        <SidebarItem
          label={__("Obligations")}
          icon={IconBook}
          to={`${prefix}/obligations`}
        />
      )}
      {isAuthorized("Organization", "listContinualImprovements") && (
        <SidebarItem
          label={__("Continual Improvements")}
          icon={IconRotateCw}
          to={`${prefix}/continual-improvements`}
        />
      )}
      {isAuthorized("Organization", "listProcessingActivities") && (
        <SidebarItem
          label={__("Processing Activities")}
          icon={IconCircleProgress}
          to={`${prefix}/processing-activities`}
        />
      )}
      {isAuthorized("Organization", "listSnapshots") && (
        <SidebarItem
          label={__("Snapshots")}
          icon={IconClock}
          to={`${prefix}/snapshots`}
        />
      )}
      {isAuthorized("Organization", "getTrustCenter") && (
        <SidebarItem
          label={__("Trust Center")}
          icon={IconShield}
          to={`${prefix}/trust-center`}
        />
      )}
      {isAuthorized("Organization", "listMembers") && (
        <SidebarItem
          label={__("Settings")}
          icon={IconSettingsGear2}
          to={`${prefix}/settings`}
        />
      )}
    </ul>
  );
}
