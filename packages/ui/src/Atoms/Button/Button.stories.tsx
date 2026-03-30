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

import { IconPlusLarge } from "../Icons";

import { Button } from "./Button";

export default {
  title: "Atoms/Button",
  component: Button,
  argTypes: {},
} satisfies Meta<typeof Button>;

type Story = StoryObj<typeof Button>;

const variants = [
  "primary",
  "secondary",
  "tertiary",
  "quaternary",
  "danger",
] as const;

export const Default: Story = {
  render() {
    return (
      <div className="space-y-4">
        {variants.map(variant => (
          <div className="flex gap-4" key={variant}>
            <Button
              key={variant}
              variant={variant}
              icon={IconPlusLarge}
              iconAfter={IconPlusLarge}
            >
              Button
            </Button>
            <Button
              key={variant}
              disabled
              variant={variant}
              icon={IconPlusLarge}
              iconAfter={IconPlusLarge}
            >
              Button
            </Button>
            <Button
              key={variant}
              variant={variant}
              icon={IconPlusLarge}
            />
          </div>
        ))}
      </div>
    );
  },
};
