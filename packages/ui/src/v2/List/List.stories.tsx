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

import type { Meta, StoryObj } from "@storybook/react";

import { Badge } from "../Badge/Badge";
import { Text } from "../typography/Text";

import { List } from "./List";
import { ListItem } from "./ListItem";
import { ListItemContent } from "./ListItemContent";
import { ListSkeleton } from "./ListSkeleton";

const entries = [
  { title: "Information Security Policy", meta: "Policy" },
  { title: "Access Control Procedure", meta: "Procedure" },
  { title: "Vendor Risk Assessment", meta: "Report" },
] as const;

export default {
  title: "v2/List",
  component: List,
  render: () => (
    <div className="w-96">
      <List>
        {entries.map(entry => (
          <ListItem key={entry.title}>
            <ListItemContent>
              <Text size={2} weight="medium" color="neutral" highContrast>
                {entry.title}
              </Text>
              <Text size={1} color="gold">
                {entry.meta}
              </Text>
            </ListItemContent>
            <Badge color="neutral" variant="soft">
              View
            </Badge>
          </ListItem>
        ))}
      </List>
    </div>
  ),
} satisfies Meta<typeof List>;

type Story = StoryObj<typeof List>;

export const Playground: Story = {};

export const Skeleton: Story = {
  render: () => (
    <div className="w-96">
      <ListSkeleton count={3} />
    </div>
  ),
};
