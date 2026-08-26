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

import {
  Button,
  Dropdown,
  DropdownItem,
  IconChevronDown,
} from "@probo/ui";

import type { ConnectMethod } from "../_lib/connectMethods";

export interface ConnectMethodAction {
  id: ConnectMethod;
  label: string;
  onSelect: () => void;
}

interface ConnectMethodSplitButtonProps {
  actions: ReadonlyArray<ConnectMethodAction>;
  chooseAnotherMethodLabel: string;
}

export function ConnectMethodSplitButton({
  actions,
  chooseAnotherMethodLabel,
}: ConnectMethodSplitButtonProps) {
  const [preferredAction, ...alternativeActions] = actions;
  if (!preferredAction) {
    return null;
  }
  if (alternativeActions.length === 0) {
    return (
      <Button type="button" variant="primary" onClick={preferredAction.onSelect}>
        {preferredAction.label}
      </Button>
    );
  }

  return (
    <div className="flex items-center">
      <Button
        type="button"
        variant="primary"
        className="rounded-r-none"
        onClick={preferredAction.onSelect}
      >
        {preferredAction.label}
      </Button>
      <Dropdown
        className="min-w-40"
        toggle={(
          <Button
            type="button"
            variant="primary"
            icon={IconChevronDown}
            aria-label={chooseAnotherMethodLabel}
            className="rounded-l-none border-l border-white/20"
          />
        )}
      >
        {alternativeActions.map(action => (
          <DropdownItem key={action.id} onSelect={action.onSelect}>
            {action.label}
          </DropdownItem>
        ))}
      </Dropdown>
    </div>
  );
}
