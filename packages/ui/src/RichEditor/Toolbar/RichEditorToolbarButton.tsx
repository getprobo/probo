// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { IconButton, type IconButtonProps } from "../../v2/IconButton/IconButton";
import { Tooltip } from "../../v2/Tooltip/Tooltip";
import { TooltipPopup } from "../../v2/Tooltip/TooltipPopup";
import { TooltipTrigger } from "../../v2/Tooltip/TooltipTrigger";

type RichEditorToolbarButtonProps = Omit<IconButtonProps, "aria-label" | "aria-pressed"> & {
  label: string;
  pressed?: boolean;
};

// An icon-only toolbar control. The tooltip names it, and a pressed control
// shows a filled background.
export function RichEditorToolbarButton(props: RichEditorToolbarButtonProps) {
  const { label, pressed, children, ...buttonProps } = props;

  return (
    <Tooltip>
      <TooltipTrigger
        render={(
          <IconButton
            size={2}
            variant={pressed ? "soft" : "ghost"}
            color="neutral"
            highContrast
            aria-label={label}
            aria-pressed={pressed}
            {...buttonProps}
          />
        )}
      >
        {children}
      </TooltipTrigger>
      <TooltipPopup>{label}</TooltipPopup>
    </Tooltip>
  );
}
