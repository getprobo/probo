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

import { XIcon } from "@phosphor-icons/react";
import { useHotkeys } from "@probo/hotkeys";
import type { ReactNode } from "react";

import { Button, type ButtonProps } from "../Button/Button";
import { shortcutAria } from "../shortcutAria";
import { ShortcutIconSwap } from "../ShortcutIconSwap/ShortcutIconSwap";
import { ShortcutLabelSwap } from "../ShortcutLabelSwap/ShortcutLabelSwap";
import { useModifierPressed } from "../useModifierPressed";

import { Dialog } from "./Dialog";
import { DialogBody } from "./DialogBody";
import { DialogClose } from "./DialogClose";
import { DialogDescription } from "./DialogDescription";
import { DialogFooter } from "./DialogFooter";
import { DialogHeader } from "./DialogHeader";
import { DialogPopup } from "./DialogPopup";
import { DialogTitle } from "./DialogTitle";
import { ShortcutHint } from "./ShortcutHint";

export interface ConfirmDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: ReactNode;
  // Accessible description. A block such as a callout is valid because the
  // description renders as a div.
  description: ReactNode;
  cancelLabel: ReactNode;
  confirmLabel: ReactNode;
  confirmIcon?: ReactNode;
  confirmColor?: ButtonProps["color"];
  loading?: boolean;
  confirmDisabled?: boolean;
  onConfirm: () => void;
}

// Confirmation dialog: title, description, and cancel/confirm actions.
// The caller owns the async work; `loading` only reflects it.
export function ConfirmDialog(props: ConfirmDialogProps) {
  const {
    open,
    onOpenChange,
    title,
    description,
    cancelLabel,
    confirmLabel,
    confirmIcon,
    confirmColor = "red",
    loading = false,
    confirmDisabled = false,
    onConfirm,
  } = props;
  const modifierPressed = useModifierPressed();
  const cancelAvailable = !loading;
  const confirmAvailable = !loading && !confirmDisabled;
  const showCancelHint = cancelAvailable && modifierPressed;
  const showConfirmHint = confirmAvailable && modifierPressed;

  useHotkeys("mod+.", () => {
    onOpenChange(false);
  }, {
    preventDefault: true,
    capture: true,
    enabled: open && cancelAvailable,
  });
  useHotkeys("mod+enter", () => {
    onConfirm();
  }, {
    preventDefault: true,
    capture: true,
    enabled: open && confirmAvailable,
  });

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogPopup>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
        </DialogHeader>
        <DialogBody>
          <DialogDescription render={<div />}>{description}</DialogDescription>
        </DialogBody>
        <DialogFooter>
          <DialogClose
            render={(
              <Button
                variant="soft"
                color="neutral"
                density="tight"
                disabled={loading}
                aria-keyshortcuts={cancelAvailable ? shortcutAria(".") : undefined}
                iconStart={(
                  <ShortcutIconSwap pressed={showCancelHint} icon={<XIcon />}>
                    <ShortcutHint shortcut="." variant="soft" color="neutral" />
                  </ShortcutIconSwap>
                )}
              >
                {cancelLabel}
              </Button>
            )}
          />
          <Button
            variant="solid"
            color={confirmColor}
            density="tight"
            iconStart={confirmIcon != null && (
              <ShortcutIconSwap pressed={showConfirmHint} icon={confirmIcon}>
                <ShortcutHint shortcut="Enter" variant="solid" color={confirmColor} />
              </ShortcutIconSwap>
            )}
            loading={loading}
            disabled={confirmDisabled}
            onClick={onConfirm}
            aria-keyshortcuts={confirmAvailable ? shortcutAria("Enter") : undefined}
          >
            {confirmIcon != null
              ? confirmLabel
              : (
                  <ShortcutLabelSwap
                    pressed={showConfirmHint}
                    hint={<ShortcutHint shortcut="Enter" variant="solid" color={confirmColor} />}
                  >
                    {confirmLabel}
                  </ShortcutLabelSwap>
                )}
          </Button>
        </DialogFooter>
      </DialogPopup>
    </Dialog>
  );
}
