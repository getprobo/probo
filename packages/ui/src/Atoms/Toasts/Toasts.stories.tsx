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
