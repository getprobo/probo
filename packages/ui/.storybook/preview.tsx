import type { Preview } from "@storybook/react";

import "../src/theme.css";
import "./preview.css";
import { BrowserRouter } from "react-router";

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
        (Story, { parameters }) => {
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
