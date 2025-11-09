import type { PreloadedQuery } from "react-relay";
import { graphql, useFragment, usePreloadedQuery } from "react-relay";
import type { MeetingGraphNodeQuery } from "/hooks/graph/__generated__/MeetingGraphNodeQuery.graphql";
import { usePageTitle } from "@probo/hooks";
import type { MeetingDetailPageMeetingFragment$key } from "./__generated__/MeetingDetailPageMeetingFragment.graphql";
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
import { useOrganizationId } from "/hooks/useOrganizationId";
import { sprintf, formatDate } from "@probo/helpers";
import { Link, Outlet, useNavigate } from "react-router";
import {
  UpdateMeetingMinutesDialog,
  type UpdateMeetingMinutesDialogRef,
} from "./dialogs/UpdateMeetingMinutesDialog";
import { useRef } from "react";
import {
  meetingNodeQuery,
  useDeleteMeetingMutation,
} from "/hooks/graph/MeetingGraph";

const meetingFragment = graphql`
  fragment MeetingDetailPageMeetingFragment on Meeting {
    id
    name
    date
    minutes
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
    node
  );
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();

  if (!meeting) {
    return <div>{__("Meeting not found")}</div>;
  }

  const [deleteMeeting, isDeleting] = useDeleteMeetingMutation();
  const confirm = useConfirm();
  const updateMinutesDialogRef = useRef<UpdateMeetingMinutesDialogRef>(null);

  usePageTitle(meeting.name);

  const handleDelete = () => {
    confirm(
      () =>
        deleteMeeting({
          variables: {
            input: { meetingId: meeting.id },
          },
          onSuccess: () => {
            navigate(`/organizations/${organizationId}/meetings`);
          },
        }),
      {
        message: sprintf(
          __(
            'This will permanently delete the meeting "%s". This action cannot be undone.'
          ),
          meeting.name
        ),
      }
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
          <ActionDropdown variant="secondary">
            <DropdownItem
              onClick={() => updateMinutesDialogRef.current?.open()}
              icon={IconPencil}
            >
              {__("Edit minutes")}
            </DropdownItem>
            <DropdownItem
              variant="danger"
              icon={IconTrashCan}
              disabled={isDeleting}
              onClick={handleDelete}
            >
              {__("Delete meeting")}
            </DropdownItem>
          </ActionDropdown>
        </div>
        <PageHeader
          title={meeting.name}
          description={formatDate(meeting.date)}
        />
        {meeting.attendees && meeting.attendees.length > 0 && (
          <div className="flex gap-2 items-center flex-wrap">
            {meeting.attendees.map((attendee) => (
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
