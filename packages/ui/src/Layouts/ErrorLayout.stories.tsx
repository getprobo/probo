// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

import { Button } from "../Atoms/Button/Button";

import {
  ErrorDetailMessage,
  ErrorDetails,
  ErrorLayout,
} from "./ErrorLayout";

export default {
  title: "Layouts/ErrorLayout",
  component: ErrorLayout,
  argTypes: {},
} satisfies Meta<typeof ErrorLayout>;

type Story = StoryObj<typeof ErrorLayout>;

export const Default: Story = {
  args: {
    showLogo: true,
    title: "Something went wrong",
    description: "We hit an unexpected error. Head back home to continue.",
    actions: <Button>Go home</Button>,
    children: (
      <ErrorDetails summary="Technical details">
        <ErrorDetailMessage>
          Relay: Missing @required value at path &apos;organization&apos; in
          &apos;ViewerMembershipLayoutQuery&apos;.
        </ErrorDetailMessage>
      </ErrorDetails>
    ),
  },
};

export const NotFound: Story = {
  args: {
    title: "Page not found",
    description: "The page you are looking for does not exist.",
    actions: <Button>Go home</Button>,
  },
};
