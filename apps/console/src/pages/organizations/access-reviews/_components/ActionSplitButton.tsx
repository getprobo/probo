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

export interface ActionSplitButtonAction {
  id: string;
  label: string;
  href?: string;
  onSelect?: () => void;
}

interface ActionSplitButtonProps {
  actions: ReadonlyArray<ActionSplitButtonAction>;
  chooseAnotherMethodLabel: string;
}

export function ActionSplitButton({
  actions,
  chooseAnotherMethodLabel,
}: ActionSplitButtonProps) {
  const [preferredAction, ...alternativeActions] = actions;
  if (!preferredAction) {
    return null;
  }

  const splitClassName
    = alternativeActions.length > 0 ? "rounded-r-none" : undefined;
  const preferredButton = preferredAction.href
    ? (
        <Button
          type="button"
          variant="primary"
          className={splitClassName}
          asChild
        >
          <a
            href={preferredAction.href}
            target="_blank"
            rel="noopener noreferrer"
          >
            {preferredAction.label}
          </a>
        </Button>
      )
    : (
        <Button
          type="button"
          variant="primary"
          className={splitClassName}
          onClick={preferredAction.onSelect}
        >
          {preferredAction.label}
        </Button>
      );

  if (alternativeActions.length === 0) {
    return preferredButton;
  }

  return (
    <div className="flex items-center">
      {preferredButton}
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
          action.href
            ? (
                <DropdownItem key={action.id} asChild>
                  <a
                    href={action.href}
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    {action.label}
                  </a>
                </DropdownItem>
              )
            : (
                <DropdownItem key={action.id} onSelect={action.onSelect}>
                  {action.label}
                </DropdownItem>
              )
        ))}
      </Dropdown>
    </div>
  );
}
