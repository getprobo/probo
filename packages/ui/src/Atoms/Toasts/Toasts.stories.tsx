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
