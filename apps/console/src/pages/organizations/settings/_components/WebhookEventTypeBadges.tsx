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
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Popover } from "@probo/ui/src/v2/Popover/Popover";
import { PopoverPopup } from "@probo/ui/src/v2/Popover/PopoverPopup";
import { PopoverTrigger } from "@probo/ui/src/v2/Popover/PopoverTrigger";
import { useTranslation } from "react-i18next";

import { webhookEventTypeLabel } from "../_lib/webhookEventTypes";
import { webhookSubscriptionListItem } from "../variants";

const VISIBLE_EVENT_COUNT = 5;

interface WebhookEventTypeBadgesProps {
  selectedEvents: readonly string[];
}

export function WebhookEventTypeBadges({
  selectedEvents,
}: WebhookEventTypeBadgesProps) {
  const { t } = useTranslation();
  const { events, eventsAll, badge, action } = webhookSubscriptionListItem();
  const visibleEvents = selectedEvents.slice(0, VISIBLE_EVENT_COUNT);
  const hiddenEvents = selectedEvents.slice(VISIBLE_EVENT_COUNT);

  return (
    <div className={events()}>
      {visibleEvents.map(event => (
        <Badge key={event} variant="soft" color="neutral" className={badge()}>
          {webhookEventTypeLabel(event)}
        </Badge>
      ))}
      {hiddenEvents.length > 0 && (
        <Popover>
          <PopoverTrigger
            render={(
              <Button variant="ghost" color="neutral" size={1} className={action()}>
                {t("webhooksSettingsPage.seeAll")}
              </Button>
            )}
          />
          <PopoverPopup className={eventsAll()}>
            {selectedEvents.map(event => (
              <Badge key={event} variant="soft" color="neutral" className={badge()}>
                {webhookEventTypeLabel(event)}
              </Badge>
            ))}
          </PopoverPopup>
        </Popover>
      )}
    </div>
  );
}
