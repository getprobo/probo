import { formatDate, sprintf } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Avatar,
  Button,
  Card,
  DropdownItem,
  IconCheckmark1,
  IconCrossLargeX,
  IconPencil,
  IconPlusLarge,
  IconTrashCan,
  Markdown,
  PageHeader,
  Tbody,
  Td,
  Textarea,
  Th,
  Thead,
  Tr,
  useConfirm,
} from "@probo/ui";
import { useRef, useState } from "react";
import {
  type PreloadedQuery,
  useFragment,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";
import { Link } from "react-router";
import { graphql } from "relay-runtime";

import type { MeetingGraphListQuery } from "#/__generated__/core/MeetingGraphListQuery.graphql";
import type { MeetingsPage_UpdateSummaryMutation } from "#/__generated__/core/MeetingsPage_UpdateSummaryMutation.graphql";
import type { MeetingsPageListFragment$key } from "#/__generated__/core/MeetingsPageListFragment.graphql";
import type { MeetingsPageRowFragment$key } from "#/__generated__/core/MeetingsPageRowFragment.graphql";
import { SortableTable, SortableTh } from "#/components/SortableTable";
import {
  meetingsQuery,
  useDeleteMeetingMutation,
} from "#/hooks/graph/MeetingGraph";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

import { CreateMeetingDialog } from "./dialogs/CreateMeetingDialog";

const meetingsFragment = graphql`
  fragment MeetingsPageListFragment on Organization
  @refetchable(queryName: "MeetingsListQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: {
      type: "MeetingOrder"
      defaultValue: { field: DATE, direction: DESC }
    }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    id
    context {
      summary
    }
    meetings(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "MeetingsListQuery_meetings") {
      __id
      edges {
        node {
          id
          ...MeetingsPageRowFragment
        }
      }
    }
  }
`;

type Props = {
  queryRef: PreloadedQuery<MeetingGraphListQuery>;
};

export default function MeetingsPage(props: Props) {
  const { __ } = useTranslate();
  const organization = usePreloadedQuery(
    meetingsQuery,
    props.queryRef,
  ).organization;

  const pagination = usePaginationFragment(
    meetingsFragment,
    organization as MeetingsPageListFragment$key,
  );

  const meetingNodes = pagination.data.meetings.edges
    .map(edge => edge.node)
    .filter(Boolean);
  const connectionId = pagination.data.meetings.__id;

  usePageTitle(__("Meetings"));

  const [isEditing, setIsEditing] = useState(false);
  const summary = pagination.data.context?.summary || "";
  const [summaryText, setSummaryText] = useState(summary);
  // Local state to track the displayed summary (updated immediately on save)
  const [displayedSummary, setDisplayedSummary] = useState(summary);
  // Track if we just saved to prevent useEffect from overwriting our update
  const justSavedRef = useRef(false);

  const updateSummaryMutation = graphql`
    mutation MeetingsPage_UpdateSummaryMutation(
      $input: UpdateOrganizationContextInput!
    ) {
      updateOrganizationContext(input: $input) {
        context {
          organizationId
          summary
        }
      }
    }
  `;

  const [updateSummary, isUpdating]
    = useMutationWithToasts<MeetingsPage_UpdateSummaryMutation>(
      updateSummaryMutation,
      {
        successMessage: __("Summary updated successfully"),
        errorMessage: __("Failed to update summary"),
      },
    );

  const handleSave = async () => {
    const valueToSave = summaryText.trim();
    const previousValue = pagination.data.context?.summary || "";
    setDisplayedSummary(valueToSave);
    justSavedRef.current = true;

    const valueToSend = valueToSave.length > 0 ? valueToSave : "";

    await updateSummary({
      variables: {
        input: {
          organizationId: organization.id,
          summary: valueToSend || null,
        },
      },
      onError: () => {
        // Roll back optimistic update on error
        setDisplayedSummary(previousValue);
        justSavedRef.current = false;
      },
      onCompleted: (_, errors) => {
        if (errors?.length) {
          // Roll back optimistic update on GraphQL error
          setDisplayedSummary(previousValue);
          justSavedRef.current = false;
        }

        setIsEditing(false);
      },
    });
  };

  const handleCancel = () => {
    setSummaryText(pagination.data.context?.summary || "");
    setIsEditing(false);
  };

  return (
    <div className="space-y-6">
      <Card padded>
        {isEditing
          ? (
              <div className="space-y-4">
                <Textarea
                  value={summaryText}
                  onChange={e => setSummaryText(e.target.value)}
                  autogrow
                  className="min-h-32 font-mono text-sm"
                  placeholder={__("Enter meetings summary in markdown format")}
                />
                <div className="flex gap-2 justify-end">
                  <Button
                    variant="secondary"
                    icon={IconCrossLargeX}
                    onClick={handleCancel}
                    disabled={isUpdating}
                  >
                    {__("Cancel")}
                  </Button>
                  <Button
                    icon={IconCheckmark1}
                    onClick={() => void handleSave()}
                    disabled={isUpdating}
                  >
                    {__("Save")}
                  </Button>
                </div>
              </div>
            )
          : (
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <h3 className="text-sm font-semibold text-txt-secondary">
                    {__("Summary")}
                  </h3>
                  {organization.canCreateMeeting && (
                    <Button
                      variant="quaternary"
                      icon={IconPencil}
                      onClick={() => setIsEditing(true)}
                    >
                      {__("Edit")}
                    </Button>
                  )}
                </div>
                <div className="w-full">
                  {displayedSummary
                    ? (
                        <div className="prose prose-sm max-w-none w-full [&_.prose]:max-w-none">
                          <Markdown content={displayedSummary} />
                        </div>
                      )
                    : (
                        <div className="text-txt-tertiary text-sm italic">
                          {__("No summary yet. Click Edit to add one.")}
                        </div>
                      )}
                </div>
              </div>
            )}
      </Card>
      <PageHeader
        title={__("Meetings")}
        description={__(
          "Track and manage your organization's meetings and their minutes.",
        )}
      >
        {organization.canCreateMeeting && (
          <CreateMeetingDialog connectionId={connectionId}>
            <Button icon={IconPlusLarge}>{__("Add meeting")}</Button>
          </CreateMeetingDialog>
        )}
      </PageHeader>
      {meetingNodes.length > 0
        ? (
            <SortableTable {...pagination}>
              <Thead>
                <Tr>
                  <SortableTh field="DATE" className="w-40">
                    {__("Date")}
                  </SortableTh>
                  <SortableTh field="NAME" className="min-w-0">
                    {__("Meeting name")}
                  </SortableTh>
                  <Th className="w-60">{__("Attendees")}</Th>
                  <Th className="w-18"></Th>
                </Tr>
              </Thead>
              <Tbody>
                {meetingNodes.map(meeting => (
                  <MeetingRow
                    key={meeting.id}
                    meeting={meeting}
                    organizationId={organization.id}
                  />
                ))}
              </Tbody>
            </SortableTable>
          )
        : (
            <Card padded>
              <div className="text-center py-12">
                <h3 className="text-lg font-semibold mb-2">
                  {__("No meetings yet")}
                </h3>
                <p className="text-txt-tertiary mb-4">
                  {__("Create your first meeting to get started.")}
                </p>
              </div>
            </Card>
          )}
    </div>
  );
}

const rowFragment = graphql`
  fragment MeetingsPageRowFragment on Meeting {
    id
    name
    date
    attendees {
      id
      fullName
    }
    canDelete: permission(action: "core:meeting:delete")
  }
`;

function MeetingRow({
  meeting: meetingKey,
  organizationId,
}: {
  meeting: MeetingsPageRowFragment$key;
  organizationId: string;
}) {
  const meeting = useFragment<MeetingsPageRowFragment$key>(
    rowFragment,
    meetingKey,
  );
  const { __ } = useTranslate();
  const [deleteMeeting] = useDeleteMeetingMutation();
  const confirm = useConfirm();
  const handleDelete = () => {
    confirm(
      () =>
        deleteMeeting({
          variables: {
            input: { meetingId: meeting.id },
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
    <Tr to={`/organizations/${organizationId}/meetings/${meeting.id}`}>
      <Td className="w-40">{formatDate(meeting.date)}</Td>
      <Td className="min-w-0">
        <div className="flex gap-4 items-center">{meeting.name}</div>
      </Td>
      <Td className="w-60">
        {meeting.attendees && meeting.attendees.length > 0
          ? (
              <div className="flex gap-2 items-center flex-wrap">
                {meeting.attendees.map(attendee => (
                  <div key={attendee.id} className="flex gap-2 items-center">
                    <Avatar name={attendee.fullName ?? ""} />
                    <Link
                      to={`/organizations/${organizationId}/people/${attendee.id}`}
                      onClick={(e) => {
                        e.stopPropagation();
                      }}
                      className="text-sm hover:underline"
                    >
                      {attendee.fullName}
                    </Link>
                  </div>
                ))}
              </div>
            )
          : (
              <span className="text-txt-tertiary text-sm">
                {__("No attendees")}
              </span>
            )}
      </Td>
      {meeting.canDelete && (
        <Td noLink width={50} className="text-end w-18">
          <ActionDropdown>
            <DropdownItem
              variant="danger"
              icon={IconTrashCan}
              onClick={handleDelete}
            >
              {__("Delete")}
            </DropdownItem>
          </ActionDropdown>
        </Td>
      )}
    </Tr>
  );
}
