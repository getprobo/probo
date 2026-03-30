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

import { Button } from "../Button/Button";

import { Toast, Toasts, useToast } from "./Toasts";

export default {
  title: "Atoms/Toasts",
  component: Toasts,
  argTypes: {},
} satisfies Meta<typeof Toasts>;

type Story = StoryObj<typeof Toasts>;

export const Default: Story = {
  render: function Render() {
    const { toast } = useToast();
    return (
      <>
        <Button
          className="mb-4"
          onClick={() =>
            toast({
              title: "Title",
              description: "This is a short description",
            })}
        >
          Trigger a Toast
        </Button>

        <div className="space-y-4">
          {(["success", "error", "warning", "info"] as const).map(
            variant => (
              <Toast
                key={variant}
                id={variant}
                title="Title"
                description="This is a short description"
                variant={variant}
                onClose={() => {}}
              />
            ),
          )}
        </div>
        <Toasts />
      </>
    );
  },
};
