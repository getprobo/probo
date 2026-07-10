// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import type { Meta, StoryObj } from "@storybook/react";

import { Button } from "../Button/Button";

import { ErrorState } from "./ErrorState";

const actions = (
  <>
    <Button size={2} color="neutral" highContrast>Back to trust center</Button>
    <Button size={2} variant="soft" color="neutral">Contact support</Button>
  </>
);

export default {
  title: "v2/ErrorState",
  component: ErrorState,
  args: {
    code: "404",
    title: "Page not found",
    description: "The page you're looking for doesn't exist or may have been moved.",
    actions,
  },
} satisfies Meta<typeof ErrorState>;

type Story = StoryObj<typeof ErrorState>;

export const Playground: Story = {};

export const NotFound: Story = {
  args: {
    code: "404",
    title: "Page not found",
    description: "The page you're looking for doesn't exist or may have been moved.",
  },
};

export const Forbidden: Story = {
  args: {
    code: "403",
    title: "Access denied",
    description: "You don't have permission to view this page.",
  },
};

export const ServerError: Story = {
  args: {
    code: "500",
    title: "Something went wrong",
    description: "We hit an unexpected error. Please try again later.",
  },
};

export const WithoutCode: Story = {
  args: {
    code: undefined,
    title: "Something went wrong",
    description: "We hit an unexpected error. Please try again later.",
  },
};
