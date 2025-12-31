import { useTranslate } from "@probo/i18n";
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  Input,
  Spinner,
  useDialogRef,
} from "@probo/ui";
import { graphql } from "relay-runtime";
import { useMutation } from "react-relay";
import { z } from "zod";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import type { CreateMeetingDialogCreateMutation } from "/__generated__/core/CreateMeetingDialogCreateMutation.graphql";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { PeopleMultiSelectField } from "/components/form/PeopleMultiSelectField";
import { formatDatetime } from "@probo/helpers";

const createMeetingMutation = graphql`
  mutation CreateMeetingDialogCreateMutation(
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
          canDelete: permission(action: "core:meeting:delete")
        }
      }
    }
  }
`;

type Props = {
  children: React.ReactElement;
  connectionId: string;
};

const meetingSchema = z.object({
  name: z.string().min(1, "Meeting name is required"),
  date: z.string().min(1, "Date is required"),
  attendeeIds: z.array(z.string()).optional(),
});

export function CreateMeetingDialog({ children, connectionId }: Props) {
  const { __ } = useTranslate();
  const dialogRef = useDialogRef();
  const organizationId = useOrganizationId();
  const [createMeeting, isCreating] =
    useMutation<CreateMeetingDialogCreateMutation>(createMeetingMutation);
  const { handleSubmit, register, control } = useFormWithSchema(
    meetingSchema,
    {},
  );

  const onSubmit = handleSubmit((data) => {
    createMeeting({
      variables: {
        input: {
          organizationId,
          name: data.name,
          date: formatDatetime(data.date)!,
          attendeeIds: data.attendeeIds || null,
        },
        connections: [connectionId],
      },
      onCompleted: () => {
        dialogRef.current?.close();
      },
    });
  });

  return (
    <Dialog ref={dialogRef} trigger={children} title={__("Create meeting")}>
      <form onSubmit={onSubmit}>
        <DialogContent padded className="space-y-4">
          <Field label={__("Meeting name")} required>
            <Input
              {...register("name")}
              placeholder={__("Enter meeting name")}
              autoFocus
            />
          </Field>
          <Field label={__("Date")} required>
            <Input
              {...register("date")}
              type="date"
              placeholder={__("Select date")}
            />
          </Field>
          <PeopleMultiSelectField
            name="attendeeIds"
            control={control}
            organizationId={organizationId}
            label={__("Attendees")}
          />
        </DialogContent>
        <DialogFooter>
          <Button disabled={isCreating} type="submit">
            {isCreating && <Spinner />}
            {__("Create meeting")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
