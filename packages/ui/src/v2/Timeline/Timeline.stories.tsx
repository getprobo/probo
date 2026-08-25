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

import { CheckCircleIcon, EyeIcon, WarningCircleIcon } from "@phosphor-icons/react";

import { Text } from "../typography/Text";

import { Timeline } from "./Timeline";
import { TimelineContent } from "./TimelineContent";
import { TimelineItem } from "./TimelineItem";
import { TimelineMarker } from "./TimelineMarker";

export default {
  title: "v2/Timeline",
  component: Timeline,
};

export function Default() {
  return (
    <Timeline>
      <TimelineItem>
        <TimelineMarker>
          <EyeIcon />
        </TimelineMarker>
        <TimelineContent>
          <Text size={1} highContrast>Viewed the document</Text>
          <Text size={1} color="faint">25 Aug 2026, 21:04:12</Text>
        </TimelineContent>
      </TimelineItem>
      <TimelineItem>
        <TimelineMarker>
          <CheckCircleIcon />
        </TimelineMarker>
        <TimelineContent>
          <Text size={1} highContrast>Agreed to sign electronically</Text>
          <Text size={1} color="faint">25 Aug 2026, 21:04:18</Text>
        </TimelineContent>
      </TimelineItem>
      <TimelineItem>
        <TimelineMarker color="red">
          <WarningCircleIcon />
        </TimelineMarker>
        <TimelineContent>
          <Text size={1} highContrast>Processing failed</Text>
          <Text size={1} color="faint">25 Aug 2026, 21:04:40</Text>
        </TimelineContent>
      </TimelineItem>
    </Timeline>
  );
}
