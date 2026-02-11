import type { StorybookConfig } from "@storybook/react-vite";

const config: StorybookConfig = {
  stories: ["../src/**/*.stories.@(js|jsx|mjs|ts|tsx)"],
  addons: [
    import.meta.resolve("@chromatic-com/storybook"),
    import.meta.resolve("@storybook/addon-vitest"),
  ],
  framework: {
    name: "@storybook/react-vite",
    options: {},
  },
};
export default config;
