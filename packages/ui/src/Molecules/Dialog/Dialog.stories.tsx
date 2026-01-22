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
