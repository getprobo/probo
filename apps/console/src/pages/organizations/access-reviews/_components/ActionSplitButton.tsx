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
import { Link } from "react-router";

export interface ActionSplitButtonAction {
  id: string;
  label: string;
  href?: string;
  to?: string;
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
  const preferredButton = (
    <ActionButton action={preferredAction} className={splitClassName} />
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
          <ActionMenuItem key={action.id} action={action} />
        ))}
      </Dropdown>
    </div>
  );
}

function ActionButton({
  action,
  className,
}: {
  action: ActionSplitButtonAction;
  className?: string;
}) {
  if (action.to) {
    return (
      <Button variant="primary" className={className} asChild>
        <Link to={action.to}>{action.label}</Link>
      </Button>
    );
  }
  if (action.href) {
    return (
      <Button variant="primary" className={className} asChild>
        <a
          href={action.href}
          target="_blank"
          rel="noopener noreferrer"
        >
          {action.label}
        </a>
      </Button>
    );
  }
  return (
    <Button
      type="button"
      variant="primary"
      className={className}
      onClick={action.onSelect}
    >
      {action.label}
    </Button>
  );
}

function ActionMenuItem({ action }: { action: ActionSplitButtonAction }) {
  if (action.to) {
    return (
      <DropdownItem asChild>
        <Link to={action.to}>{action.label}</Link>
      </DropdownItem>
    );
  }
  if (action.href) {
    return (
      <DropdownItem asChild>
        <a
          href={action.href}
          target="_blank"
          rel="noopener noreferrer"
        >
          {action.label}
        </a>
      </DropdownItem>
    );
  }
  return (
    <DropdownItem onSelect={action.onSelect}>
      {action.label}
    </DropdownItem>
  );
}
