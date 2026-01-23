import { objectKeys } from "@probo/helpers";
import type { Meta, StoryObj } from "@storybook/react";

import * as icons from "./Atoms/Icons";

const meta = {
  title: "Foundations",
} satisfies Meta;

export default meta;
type Story = StoryObj<typeof meta>;

export const Colors: Story = {
  render() {
    const colors = [
      ["Surface", ["level-0", "level-1", "level-2", "level-3"]],
      [
        "Text",
        [
          "txt-primary",
          "txt-invert",
          "txt-secondary",
          "txt-tertiary",
          "txt-quaternary",
          "txt-disabled",
          "txt-accent",
          "txt-danger",
          "txt-success",
          "txt-info",
          "txt-warning",
        ],
      ],
      [
        "Background",
        [
          "bg-primary",
          "bg-invert",
          "bg-secondary",
          "bg-tertiary",
          "bg-subtle",
          "bg-highlight",
          "bg-active",
          "bg-accent",
          "bg-danger",
          "bg-success",
          "bg-info",
          "bg-warning",
        ],
      ],
      [
        "Border",
        [
          "border-solid",
          "border-low",
          "border-mid",
          "border-strong",
          "border-active",
          "border-accent",
          "border-danger",
          "border-success",
          "border-warning",
          "border-info",
        ],
      ],
      [
        "Hover",
        [
          "bg-primary-hover",
          "bg-invert-hover",
          "bg-secondary-hover",
          "bg-tertiary-hover",
          "bg-subtle-hover",
          "bg-highlight-hover",
          "bg-active-hover",
          "bg-accent-hover",
          "bg-danger-hover",
          "bg-success-hover",
          "bg-info-hover",
          "bg-warning-hover",
        ],
      ],
      [
        "Pressed",
        [
          "bg-primary-pressed",
          "bg-invert-pressed",
          "bg-secondary-pressed",
          "bg-tertiary-pressed",
          "bg-subtle-pressed",
          "bg-highlight-pressed",
          "bg-active-pressed",
          "bg-accent-pressed",
          "bg-danger-pressed",
          "bg-success-pressed",
          "bg-info-pressed",
          "bg-warning-pressed",
        ],
      ],
    ] as [string, string[]][];
    return (
      <div className="divide-dashed divide-y divide-border-low">
        {colors.map(([name, variants]) => (
          <div key={name} className="space-y-10 py-10">
            <h2 className="text-xl">{name}</h2>
            <div className="flex flex-wrap gap-8">
              {variants.map(variant => (
                <div
                  className="space-y-3"
                  key={variant}
                  style={{ width: 200 }}
                >
                  <div
                    className="rounded-lg size-10 border border-border-low"
                    style={{
                      backgroundColor: `var(--color-${variant})`,
                    }}
                  />
                  <div className="text-sm font-medium">
                    {variant}
                  </div>
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>
    );
  },
};

export const Typography: Story = {
  render() {
    const textVariants = [
      "text-5xl",
      "text-4xl",
      "text-3xl",
      "text-2xl",
      "text-xl",
      "text-xl-medium",
      "text-lg",
      "text-lg-bold",
      "text-base",
      "text-base-link",
      "text-sm",
      "text-xs",
      "text-xxs",
    ];
    return (
      <div className="divide-dashed divide-y divide-border-low">
        {textVariants.map(variant => (
          <div className="flex items-center py-4" key={variant}>
            <div className="opacity-50 w-50 flex-none">
              {variant}
            </div>
            <div className={variant}>
              This is a text
              {variant}
            </div>
          </div>
        ))}
      </div>
    );
  },
};

export const Icons: Story = {
  render() {
    return (
      <div className="flex flex-wrap gap-4">
        {objectKeys(icons).map((k) => {
          /* eslint import-x/namespace: ['error', { allowComputed: true }] */
          const IconComponent = icons[k];
          return (
            <button
              title={k}
              key={k}
              className="flex flex-col items-center gap-2 w-[50px] overflow-hidden"
            >
              <IconComponent />
              <div className="text-center text-xs text-txt-secondary truncate w-[50px]">
                {k.replace("Icon", "")}
              </div>
            </button>
          );
        })}
      </div>
    );
  },
};
