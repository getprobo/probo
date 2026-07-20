// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { Badge } from "@probo/ui/src/v2/Badge/Badge";
import { ListItem } from "@probo/ui/src/v2/List/ListItem";
import { ListItemContent } from "@probo/ui/src/v2/List/ListItemContent";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import { formatRelativeTime } from "#/lib/datetime/relativeTime";

import {
  formatRightsRequestReference,
  getRightsRequestStatusBadge,
  getRightsRequestTypeIcon,
} from "../_lib/rightsRequest";

import type { RightsRequestListItem_rightsRequest$key } from "./__generated__/RightsRequestListItem_rightsRequest.graphql";
import { rightsRequestList } from "./variants";

const rightsRequestListItemFragment = graphql`
  fragment RightsRequestListItem_rightsRequest on RightsRequest {
    id
    requestType
    requestState
    actionTaken
    createdAt
  }
`;

interface RightsRequestListItemProps {
  rightsRequestKey: RightsRequestListItem_rightsRequest$key;
}

// A single data request row: type icon + title, reference and optional response
// message, with the submitted time and a status badge on the trailing edge.
export function RightsRequestListItem({ rightsRequestKey }: RightsRequestListItemProps) {
  const { t, i18n } = useTranslation("requests");
  const request = useFragment(rightsRequestListItemFragment, rightsRequestKey);

  const badge = getRightsRequestStatusBadge(request.requestState);
  const reference = formatRightsRequestReference(request.id, request.createdAt);
  const { icon, subline, trailing } = rightsRequestList();

  return (
    <ListItem className="max-sm:flex-col max-sm:items-stretch max-sm:gap-2">
      <div className="flex min-w-0 flex-1 items-center gap-4">
        <span className={icon()}>
          {getRightsRequestTypeIcon(request.requestType)}
        </span>
        <ListItemContent>
          <Text size={2} weight="medium" color="neutral" highContrast className="truncate">
            {t(`types.${request.requestType}`)}
          </Text>
          <div className={subline()}>
            <Text size={1} color="gold">
              {reference}
            </Text>
            {request.actionTaken != null && request.actionTaken !== "" && (
              <Text size={1} color="faint" className="truncate">
                {`· ${request.actionTaken}`}
              </Text>
            )}
          </div>
        </ListItemContent>
      </div>
      <div className={trailing()}>
        <Text size={1} color="faint">
          {formatRelativeTime(request.createdAt, i18n.language)}
        </Text>
        <Badge color={badge.color} variant={badge.variant}>
          {t(`status.${request.requestState}`)}
        </Badge>
      </div>
    </ListItem>
  );
}
