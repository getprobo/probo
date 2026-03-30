// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

import { useEffect } from "react";
import { useQueryLoader } from "react-relay";
import { useParams } from "react-router";

import type { MeetingDetailPageQuery } from "#/__generated__/core/MeetingDetailPageQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import MeetingDetailPage, { meetingDetailPageQuery } from "./MeetingDetailPage";

function MeetingDetailPageQueryLoader() {
  const { meetingId } = useParams<{ meetingId: string }>();
  const [queryRef, loadQuery] = useQueryLoader<MeetingDetailPageQuery>(meetingDetailPageQuery);

  useEffect(() => {
    if (meetingId) {
      loadQuery({ meetingId });
    }
  }, [meetingId, loadQuery]);

  if (!queryRef) return <PageSkeleton />;

  return <MeetingDetailPage queryRef={queryRef} />;
}

export default function MeetingDetailPageLoader() {
  return (
    <CoreRelayProvider>
      <MeetingDetailPageQueryLoader />
    </CoreRelayProvider>
  );
}
