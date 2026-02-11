import { formatDate, sprintf } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Avatar,
  Breadcrumb,
  DropdownItem,
  IconPencil,
  IconTrashCan,
  PageHeader,
  useConfirm,
} from "@probo/ui";
import { useRef } from "react";
import type { PreloadedQuery } from "react-relay";
import { graphql, useFragment, usePreloadedQuery } from "react-relay";
import { Link, Outlet, useNavigate } from "react-router";

import type { MeetingDetailPageMeetingFragment$key } from "#/__generated__/core/MeetingDetailPageMeetingFragment.graphql";
import type { MeetingGraphNodeQuery } from "#/__generated__/core/MeetingGraphNodeQuery.graphql";
import {
  meetingNodeQuery,
  useDeleteMeetingMutation,
} from "#/hooks/graph/MeetingGraph";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import {
  UpdateMeetingMinutesDialog,
  type UpdateMeetingMinutesDialogRef,
} from "./dialogs/UpdateMeetingMinutesDialog";

const meetingFragment = graphql`
  fragment MeetingDetailPageMeetingFragment on Meeting {
    id
    name
    date
    minutes
    canUpdate: permission(action: "core:meeting:update")
    canDelete: permission(action: "core:meeting:delete")
    attendees {
      id
      fullName
    }
  }
`;

type Props = {
  queryRef: PreloadedQuery<MeetingGraphNodeQuery>;
};

export default function MeetingDetailPage(props: Props) {
  const node = usePreloadedQuery(meetingNodeQuery, props.queryRef).node;
  const meeting = useFragment<MeetingDetailPageMeetingFragment$key>(
    meetingFragment,
    node,
  );
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();

  const [deleteMeeting, isDeleting] = useDeleteMeetingMutation();
  const confirm = useConfirm();
  const updateMinutesDialogRef = useRef<UpdateMeetingMinutesDialogRef>(null);

  usePageTitle(meeting.name);

  const hasAnyAction = meeting.canUpdate || meeting.canDelete;

  const handleDelete = () => {
    confirm(
      () =>
        deleteMeeting({
          variables: {
            input: { meetingId: meeting.id },
          },
          onSuccess: () => {
            void navigate(`/organizations/${organizationId}/meetings`);
          },
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete the meeting \"%s\". This action cannot be undone.",
          ),
          meeting.name,
        ),
      },
    );
  };

  return (
    <>
      <UpdateMeetingMinutesDialog
        ref={updateMinutesDialogRef}
        meeting={meeting}
      />
      <div className="space-y-6">
        <div className="flex justify-between items-center mb-4">
          <Breadcrumb
            items={[
              {
                label: __("Meetings"),
                to: `/organizations/${organizationId}/meetings`,
              },
              {
                label: meeting.name,
              },
            ]}
          />
          {hasAnyAction && (
            <ActionDropdown variant="secondary">
              {meeting.canUpdate && (
                <DropdownItem
                  onClick={() => updateMinutesDialogRef.current?.open()}
                  icon={IconPencil}
                >
                  {__("Edit minutes")}
                </DropdownItem>
              )}
              {meeting.canDelete && (
                <DropdownItem
                  variant="danger"
                  icon={IconTrashCan}
                  disabled={isDeleting}
                  onClick={handleDelete}
                >
                  {__("Delete meeting")}
                </DropdownItem>
              )}
            </ActionDropdown>
          )}
        </div>
        <PageHeader
          title={meeting.name}
          description={formatDate(meeting.date)}
        />
        {meeting.attendees && meeting.attendees.length > 0 && (
          <div className="flex gap-2 items-center flex-wrap">
            {meeting.attendees.map(attendee => (
              <div key={attendee.id} className="flex gap-2 items-center">
                <Avatar name={attendee.fullName ?? ""} />
                <Link
                  to={`/organizations/${organizationId}/people/${attendee.id}`}
                  className="text-sm hover:underline"
                >
                  {attendee.fullName}
                </Link>
              </div>
            ))}
          </div>
        )}
        <Outlet context={{ meeting }} />
      </div>
    </>
  );
}
