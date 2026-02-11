import type { Preview } from "@storybook/react";
import { useEffect } from "react";
import { BrowserRouter } from "react-router";

import "../src/theme.css";
import "./preview.css";

const preview: Preview = {
  parameters: {
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/i,
      },
    },
  },
  decorators: [
    (Story) => {
      useEffect(() => {
        document.body.classList.add("bg-level-0");
      }, []);
      return (
        <BrowserRouter>
          <div className="text-txt-primary">
            <Story />
          </div>
        </BrowserRouter>
      );
    },
  ],
};

export default preview;
