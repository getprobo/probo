import type { PreloadedQuery } from "react-relay";
import { graphql, useFragment, usePreloadedQuery } from "react-relay";
import type { MeetingGraphNodeQuery } from "/__generated__/core/MeetingGraphNodeQuery.graphql";
import { usePageTitle } from "@probo/hooks";
import type { MeetingDetailPageMeetingFragment$key } from "/__generated__/core/MeetingDetailPageMeetingFragment.graphql";
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
import { useRef, useState, useEffect, use } from "react";
import {
  meetingNodeQuery,
  useDeleteMeetingMutation,
} from "/hooks/graph/MeetingGraph";
import { PermissionsContext } from "/providers/PermissionsContext";

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
    node,
  );
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();
  const { isAuthorized } = use(PermissionsContext);

  const [deleteMeeting, isDeleting] = useDeleteMeetingMutation();
  const confirm = useConfirm();
  const updateMinutesDialogRef = useRef<UpdateMeetingMinutesDialogRef>(null);

  const [canUpdate, setCanUpdate] = useState<boolean>(false);
  const [canDelete, setCanDelete] = useState<boolean>(false);

  useEffect(() => {
    if (!organizationId) {
      setCanUpdate(false);
      setCanDelete(false);
      return;
    }

    try {
      const updateAuth = isAuthorized("Meeting", "updateMeeting");
      setCanUpdate(updateAuth);
    } catch (promise) {
      if (promise instanceof Promise) {
        promise
          .then(() => {
            try {
              const updateAuth = isAuthorized("Meeting", "updateMeeting");
              setCanUpdate(updateAuth);
            } catch {
              setCanUpdate(false);
            }
          })
          .catch(() => {
            setCanUpdate(false);
          });
      } else {
        setCanUpdate(false);
      }
    }

    try {
      const deleteAuth = isAuthorized("Meeting", "deleteMeeting");
      setCanDelete(deleteAuth);
    } catch (promise) {
      if (promise instanceof Promise) {
        promise
          .then(() => {
            try {
              const deleteAuth = isAuthorized("Meeting", "deleteMeeting");
              setCanDelete(deleteAuth);
            } catch {
              setCanDelete(false);
            }
          })
          .catch(() => {
            setCanDelete(false);
          });
      } else {
        setCanDelete(false);
      }
    }
  }, [organizationId, isAuthorized]);

  usePageTitle(meeting.name);

  const hasAnyAction = canUpdate || canDelete;

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
            'This will permanently delete the meeting "%s". This action cannot be undone.',
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
              {canUpdate && (
                <DropdownItem
                  onClick={() => updateMinutesDialogRef.current?.open()}
                  icon={IconPencil}
                >
                  {__("Edit minutes")}
                </DropdownItem>
              )}
              {canDelete && (
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
