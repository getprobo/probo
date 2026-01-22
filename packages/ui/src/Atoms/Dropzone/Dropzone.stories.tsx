import type { Meta, StoryObj } from "@storybook/react";

import { Dropzone } from "./Dropzone";

export default {
  title: "Atoms/Dropzone",
  component: Dropzone,
  argTypes: {},
} satisfies Meta<typeof Dropzone>;

type Story = StoryObj<typeof Dropzone>;

export const Default: Story = {
  args: {
    description: "Maximum 500 MB file size",
    isUploading: false,
    disabled: false,
  },
};

export const Disabled: Story = {
  args: {
    ...Default.args,
    disabled: true,
  },
};

export const Uploading: Story = {
  args: {
    ...Default.args,
    isUploading: true,
  },
};
