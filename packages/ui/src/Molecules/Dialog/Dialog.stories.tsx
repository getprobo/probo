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

import { Button } from "../../Atoms/Button/Button";

import { Dialog, DialogContent, DialogFooter } from "./Dialog";

export default {
  title: "Atoms/Dialog",
  component: Dialog,
  argTypes: {},
} satisfies Meta<typeof Dialog>;

type Story = StoryObj<typeof Dialog>;

export const Default: Story = {
  render: (args) => {
    return (
      <Dialog
        {...args}
        trigger={<Button>Open dialog</Button>}
        title="Edit profile"
      >
        <DialogContent>
          <div>
            Lorem ipsum dolor sit amet consectetur adipisicing elit.
            Voluptate, voluptas. Doloribus incidunt cum laboriosam
            nulla magni soluta voluptatum omnis, sapiente minus
            corporis impedit explicabo vero praesentium, fugit
            possimus facilis rem.
          </div>
        </DialogContent>
        <DialogFooter>
          <Button>Save</Button>
        </DialogFooter>
      </Dialog>
    );
  },
};
