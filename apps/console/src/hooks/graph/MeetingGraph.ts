import { useTranslate } from "@probo/i18n";
import { graphql } from "relay-runtime";

import type { MeetingGraphDeleteMutation } from "#/__generated__/core/MeetingGraphDeleteMutation.graphql";
import type { MeetingGraphUpdateMutation } from "#/__generated__/core/MeetingGraphUpdateMutation.graphql";

import { useMutationWithToasts } from "../useMutationWithToasts";

export const meetingsQuery = graphql`
  query MeetingGraphListQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      id
      ... on Organization {
        canCreateMeeting: permission(action: "core:meeting:create")
      }
      ...MeetingsPageListFragment
    }
  }
`;

export const meetingNodeQuery = graphql`
  query MeetingGraphNodeQuery($meetingId: ID!) {
    node(id: $meetingId) {
      ...MeetingDetailPageMeetingFragment
    }
  }
`;

export const MeetingsConnectionKey = "MeetingsListQuery_meetings";

const deleteMeetingMutation = graphql`
  mutation MeetingGraphDeleteMutation($input: DeleteMeetingInput!) {
    deleteMeeting(input: $input) {
      deletedMeetingId @deleteRecord
    }
  }
`;

export function useDeleteMeetingMutation() {
  const { __ } = useTranslate();

  return useMutationWithToasts<MeetingGraphDeleteMutation>(
    deleteMeetingMutation,
    {
      successMessage: __("Meeting deleted successfully."),
      errorMessage: __("Failed to delete meeting"),
    },
  );
}

const updateMeetingMutation = graphql`
  mutation MeetingGraphUpdateMutation($input: UpdateMeetingInput!) {
    updateMeeting(input: $input) {
      meeting {
        id
        name
        date
        minutes
        attendees {
          id
          fullName
        }
      }
    }
  }
`;

export function useUpdateMeetingMutation() {
  const { __ } = useTranslate();

  return useMutationWithToasts<MeetingGraphUpdateMutation>(
    updateMeetingMutation,
    {
      successMessage: __("Meeting updated successfully."),
      errorMessage: __("Failed to update meeting"),
    },
  );
}

const createMeetingMutation = graphql`
  mutation MeetingGraphCreateMutation(
    $input: CreateMeetingInput!
    $connections: [ID!]!
  ) {
    createMeeting(input: $input) {
      meetingEdge @prependEdge(connections: $connections) {
        node {
          id
          name
          date
          minutes
          attendees {
            id
            fullName
          }
        }
      }
    }
  }
`;

export function useCreateMeetingMutation() {
  const { __ } = useTranslate();

  return useMutationWithToasts(createMeetingMutation, {
    successMessage: __("Meeting created successfully."),
    errorMessage: __("Failed to create meeting"),
  });
}
