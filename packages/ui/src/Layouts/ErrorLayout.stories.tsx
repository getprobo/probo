// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
