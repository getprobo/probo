import { useTranslate } from "@probo/i18n";
import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Spinner,
  Textarea,
  useDialogRef,
} from "@probo/ui";
import { forwardRef, useImperativeHandle } from "react";
import { z } from "zod";

import type { MeetingDetailPageMeetingFragment$data } from "#/__generated__/core/MeetingDetailPageMeetingFragment.graphql";
import { useUpdateMeetingMutation } from "#/hooks/graph/MeetingGraph";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

type Props = {
  meeting: MeetingDetailPageMeetingFragment$data;
};

export type UpdateMeetingMinutesDialogRef = {
  open: () => void;
};

const minutesSchema = z.object({
  minutes: z.string(),
});

export const UpdateMeetingMinutesDialog = forwardRef<
  UpdateMeetingMinutesDialogRef,
  Props
>(function UpdateMeetingMinutesDialog({ meeting }, ref) {
  const { __ } = useTranslate();
  const dialogRef = useDialogRef();
  const [updateMeeting, isUpdating] = useUpdateMeetingMutation();
  const { handleSubmit, register, reset } = useFormWithSchema(minutesSchema, {
    defaultValues: {
      minutes: meeting.minutes || "",
    },
  });

  useImperativeHandle(ref, () => ({
    open: () => {
      reset({
        minutes: meeting.minutes || "",
      });
      dialogRef.current?.open();
    },
  }));

  const onSubmit = async (data: z.infer<typeof minutesSchema>) => {
    await updateMeeting({
      variables: {
        input: {
          meetingId: meeting.id,
          minutes: data.minutes,
        },
      },
      onSuccess: () => {
        dialogRef.current?.close();
      },
    });
  };

  return (
    <Dialog
      ref={dialogRef}
      title={<Breadcrumb items={[__("Meetings"), __("Edit minutes")]} />}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent>
          <Textarea
            id="minutes"
            variant="ghost"
            autogrow
            placeholder={__("Add meeting minutes")}
            aria-label={__("Minutes")}
            className="p-6"
            {...register("minutes")}
          />
        </DialogContent>
        <DialogFooter>
          <Button disabled={isUpdating} type="submit">
            {isUpdating && <Spinner />}
            {__("Update minutes")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
});
