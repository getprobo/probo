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

import { ButtonLink } from "../Button/ButtonLink";

import { Table, type TableProps } from "./Table";
import { TableBody } from "./TableBody";
import { TableCell } from "./TableCell";
import { TableColumnHeaderCell } from "./TableColumnHeaderCell";
import { TableHeader } from "./TableHeader";
import { TableLink } from "./TableLink";
import { TableRow } from "./TableRow";
import { TableRowHeaderCell } from "./TableRowHeaderCell";
import { TableSkeleton } from "./TableSkeleton";

const rows = [
  { name: "Danilo Sousa", email: "danilo@example.com", group: "Developer" },
  { name: "Zahra Ambessa", email: "zahra@example.com", group: "Admin" },
  { name: "Jasper Eriksson", email: "jasper@example.com", group: "Developer" },
] as const;

const sizes = [1, 2, 3] as const;
const variants = ["surface", "ghost"] as const;

function SampleTable(props: Pick<TableProps, "size" | "variant">) {
  return (
    <Table {...props}>
      <TableHeader>
        <TableRow>
          <TableColumnHeaderCell>Full name</TableColumnHeaderCell>
          <TableColumnHeaderCell>Email</TableColumnHeaderCell>
          <TableColumnHeaderCell>Group</TableColumnHeaderCell>
        </TableRow>
      </TableHeader>
      <TableBody>
        {rows.map(entry => (
          <TableRow key={entry.email}>
            <TableRowHeaderCell>{entry.name}</TableRowHeaderCell>
            <TableCell>{entry.email}</TableCell>
            <TableCell>{entry.group}</TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}

export default {
  title: "v2/Table",
  component: Table,
  args: {
    size: 2,
    variant: "ghost",
  },
  render: args => (
    <div className="w-xl">
      <SampleTable size={args.size} variant={args.variant} />
    </div>
  ),
} satisfies Meta<typeof Table>;

type Story = StoryObj<typeof Table>;

export const Playground: Story = {};

export const Sizes: Story = {
  render: () => (
    <div className="flex w-xl flex-col gap-6">
      {sizes.map(size => (
        <SampleTable key={size} size={size} variant="surface" />
      ))}
    </div>
  ),
};

export const Variants: Story = {
  render: () => (
    <div className="flex w-xl flex-col gap-6">
      {variants.map(variant => (
        <SampleTable key={variant} variant={variant} />
      ))}
    </div>
  ),
};

export const Skeleton: Story = {
  render: () => (
    <div className="w-xl">
      <TableSkeleton variant="surface" count={3} columns={3} />
    </div>
  ),
};

// Title TableLink stretches across the row; the trailing ButtonLink sits
// above the overlay via TableCell interactive.
export const InteractiveRows: Story = {
  render: () => (
    <div className="w-xl">
      <Table variant="surface">
        <TableHeader>
          <TableRow>
            <TableColumnHeaderCell>Full name</TableColumnHeaderCell>
            <TableColumnHeaderCell>Email</TableColumnHeaderCell>
            <TableColumnHeaderCell>Group</TableColumnHeaderCell>
            <TableColumnHeaderCell />
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map(entry => (
            <TableRow key={entry.email} align="center" interactive>
              <TableRowHeaderCell>
                <TableLink to={`/${entry.email}`}>{entry.name}</TableLink>
              </TableRowHeaderCell>
              <TableCell>{entry.email}</TableCell>
              <TableCell>{entry.group}</TableCell>
              <TableCell interactive justify="end">
                <ButtonLink
                  to={`/${entry.email}`}
                  size={2}
                  variant="soft"
                  color="neutral"
                  highContrast
                >
                  Open
                </ButtonLink>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  ),
};
