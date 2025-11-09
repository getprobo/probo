import { useOutletContext } from "react-router";
import type { MeetingDetailPageMeetingFragment$data } from "../__generated__/MeetingDetailPageMeetingFragment.graphql";
import { Markdown } from "@probo/ui";

export default function MeetingMinutesTab() {
  const { meeting } = useOutletContext<{
    meeting: MeetingDetailPageMeetingFragment$data;
  }>();
  
  return (
    <div>
      {meeting.minutes ? (
        <Markdown content={meeting.minutes} />
      ) : (
        <div className="text-txt-tertiary text-sm">
          No minutes recorded yet. Click "Edit minutes" to add meeting minutes.
        </div>
      )}
    </div>
  );
}

