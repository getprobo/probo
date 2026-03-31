// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { formatDate, sprintf } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Avatar,
  Button,
  Card,
  DropdownItem,
  IconPlusLarge,
  IconTrashCan,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  useConfirm,
} from "@probo/ui";
import {
  type PreloadedQuery,
  useFragment,
  usePaginationFragment,
  usePreloadedQuery,
} from "react-relay";
import { Link } from "react-router";
import { graphql } from "relay-runtime";

import type { MeetingsPageDeleteMutation } from "#/__generated__/core/MeetingsPageDeleteMutation.graphql";
import type { MeetingsPageListFragment$key } from "#/__generated__/core/MeetingsPageListFragment.graphql";
import type { MeetingsPageQuery } from "#/__generated__/core/MeetingsPageQuery.graphql";
import type { MeetingsPageRowFragment$key } from "#/__generated__/core/MeetingsPageRowFragment.graphql";
import { SortableTable, SortableTh } from "#/components/SortableTable";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

import { CreateMeetingDialog } from "./dialogs/CreateMeetingDialog";

export const meetingsPageQuery = graphql`
  query MeetingsPageQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      id
      ... on Organization {
        canCreateMeeting: permission(action: "core:meeting:create")
      }
      ...MeetingsPageListFragment
    }
  }
`;

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
    canCreateMeeting: permission(action: "core:meeting:create")
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

const deleteMeetingMutation = graphql`
  mutation MeetingsPageDeleteMutation($input: DeleteMeetingInput!) {
    deleteMeeting(input: $input) {
      deletedMeetingId @deleteRecord
    }
  }
`;

function useDeleteMeetingMutation() {
  const { __ } = useTranslate();

  return useMutationWithToasts<MeetingsPageDeleteMutation>(
    deleteMeetingMutation,
    {
      successMessage: __("Meeting deleted successfully."),
      errorMessage: __("Failed to delete meeting"),
    },
  );
}

type Props = {
  queryRef: PreloadedQuery<MeetingsPageQuery>;
};

export default function MeetingsPage(props: Props) {
  const { __ } = useTranslate();
  const organization = usePreloadedQuery(
    meetingsPageQuery,
    props.queryRef,
  ).organization;

  // eslint-disable-next-line relay/generated-typescript-types
  const pagination = usePaginationFragment(
    meetingsFragment,
    organization as MeetingsPageListFragment$key,
  );

  const meetingNodes = pagination.data.meetings.edges
    .map(edge => edge.node)
    .filter(Boolean);
  const connectionId = pagination.data.meetings.__id;

  return (
    <div className="space-y-6">
      {pagination.data.canCreateMeeting && (
        <div className="flex justify-end">
          <CreateMeetingDialog connectionId={connectionId}>
            <Button icon={IconPlusLarge}>{__("Add meeting")}</Button>
          </CreateMeetingDialog>
        </div>
      )}
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
    <Tr to={`/organizations/${organizationId}/context/meetings/${meeting.id}`}>
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
