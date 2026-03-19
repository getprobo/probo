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

import type { MeetingDetailPageDeleteMutation } from "#/__generated__/core/MeetingDetailPageDeleteMutation.graphql";
import type { MeetingDetailPageMeetingFragment$key } from "#/__generated__/core/MeetingDetailPageMeetingFragment.graphql";
import type { MeetingDetailPageQuery } from "#/__generated__/core/MeetingDetailPageQuery.graphql";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import {
  UpdateMeetingMinutesDialog,
  type UpdateMeetingMinutesDialogRef,
} from "./dialogs/UpdateMeetingMinutesDialog";

export const meetingDetailPageQuery = graphql`
  query MeetingDetailPageQuery($meetingId: ID!) {
    node(id: $meetingId) {
      ...MeetingDetailPageMeetingFragment
    }
  }
`;

const meetingFragment = graphql`
  fragment MeetingDetailPageMeetingFragment on Meeting {
    id
    name
    date
    # eslint-disable-next-line relay/unused-fields
    minutes
    canUpdate: permission(action: "core:meeting:update")
    canDelete: permission(action: "core:meeting:delete")
    attendees {
      id
      fullName
    }
  }
`;

const deleteMeetingMutation = graphql`
  mutation MeetingDetailPageDeleteMutation($input: DeleteMeetingInput!) {
    deleteMeeting(input: $input) {
      deletedMeetingId @deleteRecord
    }
  }
`;

function useDeleteMeetingMutation() {
  const { __ } = useTranslate();

  return useMutationWithToasts<MeetingDetailPageDeleteMutation>(
    deleteMeetingMutation,
    {
      successMessage: __("Meeting deleted successfully."),
      errorMessage: __("Failed to delete meeting"),
    },
  );
}

type Props = {
  queryRef: PreloadedQuery<MeetingDetailPageQuery>;
};

export default function MeetingDetailPage(props: Props) {
  const node = usePreloadedQuery(meetingDetailPageQuery, props.queryRef).node;
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
            void navigate(`/organizations/${organizationId}/context/meetings`);
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
                to: `/organizations/${organizationId}/context/meetings`,
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
