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

import { CopyIcon, WarningIcon } from "@phosphor-icons/react";

import { IconButton } from "../IconButton/IconButton";
import { Text } from "../typography/Text";

import { errorNotice } from "./variants";

export interface ErrorNoticeProps {
  messages: readonly string[];
  copyLabel: string;
  onCopied?: (message: string) => void;
  className?: string;
}

// Error tied to the entity a card is showing. Each message can be copied so a
// user can share it. Labels and the copied callback come from the caller.
export function ErrorNotice({
  messages,
  copyLabel,
  onCopied,
  className,
}: ErrorNoticeProps) {
  const { root, header, icon, row, message: messageSlot } = errorNotice();

  if (messages.length === 0) {
    return null;
  }

  async function copyMessage(text: string) {
    try {
      await navigator.clipboard.writeText(text);
    } catch {
      return;
    }
    onCopied?.(text);
  }

  return (
    <div className={root({ className })}>
      <div className={header()}>
        <WarningIcon className={icon()} aria-hidden />
      </div>
      {messages.map(text => (
        <div key={text} className={row()}>
          <Text size={2} color="neutral" className={messageSlot()}>
            {text}
          </Text>
          <IconButton
            size={1}
            variant="soft"
            color="neutral"
            aria-label={copyLabel}
            onClick={() => {
              void copyMessage(text);
            }}
          >
            <CopyIcon />
          </IconButton>
        </div>
      ))}
    </div>
  );
}
