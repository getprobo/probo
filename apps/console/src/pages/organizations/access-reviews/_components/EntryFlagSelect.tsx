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

import { formatError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Badge, Checkbox, useToast } from "@probo/ui";
import * as Popover from "@radix-ui/react-popover";
import { useRef, useState } from "react";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { AccessReviewEntryFlag, EntryFlagSelectMutation } from "#/__generated__/core/EntryFlagSelectMutation.graphql";

import { flagBadgeVariant, flagGroups, flagLabel } from "./accessReviewHelpers";

const mutation = graphql`
  mutation EntryFlagSelectMutation($input: FlagAccessReviewEntryInput!) {
    flagAccessReviewEntry(input: $input) {
      accessReviewEntry {
        id
        flags
        flagReasons
      }
    }
  }
`;

type Props = {
  entryId: string;
  currentFlags: readonly AccessReviewEntryFlag[];
};

export function EntryFlagSelect({ entryId, currentFlags }: Props) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const [open, setOpen] = useState(false);
  const [localFlags, setLocalFlags] = useState<AccessReviewEntryFlag[]>([...currentFlags]);
  const openedWithRef = useRef<readonly AccessReviewEntryFlag[]>(currentFlags);
  const [flagEntry] = useMutation<EntryFlagSelectMutation>(mutation);

  const toggleFlag = (flagValue: AccessReviewEntryFlag) => {
    setLocalFlags(prev =>
      prev.includes(flagValue)
        ? prev.filter(f => f !== flagValue)
        : [...prev, flagValue],
    );
  };

  const handleOpenChange = (nextOpen: boolean) => {
    if (nextOpen) {
      openedWithRef.current = currentFlags;
      setLocalFlags([...currentFlags]);
    }

    if (!nextOpen) {
      // Submit only if flags changed since popover opened
      const changed
        = localFlags.length !== openedWithRef.current.length
          || localFlags.some(f => !openedWithRef.current.includes(f));

      if (changed) {
        flagEntry({
          variables: {
            input: {
              accessReviewEntryId: entryId,
              flags: localFlags,
            },
          },
          onCompleted(_, errors) {
            if (errors?.length) {
              toast({
                title: __("Error"),
                description: formatError(
                  __("Failed to flag entry"),
                  errors,
                ),
                variant: "error",
              });
            }
          },
          onError(error) {
            toast({
              title: __("Error"),
              description: formatError(
                __("Failed to flag entry"),
                error,
              ),
              variant: "error",
            });
          },
        });
      }
    }

    setOpen(nextOpen);
  };

  const displayFlags = open ? localFlags : [...currentFlags];

  return (
    <Popover.Root open={open} onOpenChange={handleOpenChange}>
      <Popover.Trigger asChild>
        <button
          type="button"
          className="flex items-center gap-1 text-sm cursor-pointer"
        >
          {displayFlags.length === 0
            ? (
                <span className="text-txt-tertiary">--</span>
              )
            : (
                <div className="flex flex-wrap gap-1">
                  {displayFlags.map(f => (
                    <Badge key={f} variant={flagBadgeVariant(f)} size="sm">
                      {flagLabel(f)}
                    </Badge>
                  ))}
                </div>
              )}
        </button>
      </Popover.Trigger>
      <Popover.Portal>
        <Popover.Content
          sideOffset={5}
          className="z-100 w-64 rounded-[10px] bg-level-1 p-2 shadow-mid animate-in fade-in slide-in-from-top-2"
        >
          {flagGroups.map(group => (
            <div key={group.label} className="mb-2 last:mb-0">
              <div className="px-2 py-1 text-xs font-semibold text-txt-tertiary uppercase tracking-wider">
                {__(group.label)}
              </div>
              {group.flags.map(flag => (
                <label
                  key={flag.value}
                  className="flex items-center gap-2 px-2 py-1.5 rounded cursor-pointer hover:bg-tertiary-hover"
                >
                  <Checkbox
                    checked={localFlags.includes(flag.value)}
                    onChange={() => toggleFlag(flag.value)}
                  />
                  <span className="text-sm text-txt-primary">{__(flag.label)}</span>
                </label>
              ))}
            </div>
          ))}
        </Popover.Content>
      </Popover.Portal>
    </Popover.Root>
  );
}
