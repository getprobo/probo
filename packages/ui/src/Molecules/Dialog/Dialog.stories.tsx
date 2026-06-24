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
