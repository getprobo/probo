import type { Preview } from "@storybook/react";

import "../src/theme.css";
import "./preview.css";
import { BrowserRouter } from "react-router";
import { useEffect } from "react";

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
