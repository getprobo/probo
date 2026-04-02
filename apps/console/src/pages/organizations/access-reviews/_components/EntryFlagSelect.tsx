// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { formatError, type GraphQLError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Badge, Checkbox, useToast } from "@probo/ui";
import * as Popover from "@radix-ui/react-popover";
import { useRef, useState } from "react";
import { useMutation } from "react-relay";
import { graphql } from "relay-runtime";

import type { AccessEntryFlag, EntryFlagSelectMutation } from "#/__generated__/core/EntryFlagSelectMutation.graphql";

import { flagBadgeVariant, flagGroups, flagLabel } from "./accessReviewHelpers";

const mutation = graphql`
  mutation EntryFlagSelectMutation($input: FlagAccessEntryInput!) {
    flagAccessEntry(input: $input) {
      accessEntry {
        id
        flags
        flagReasons
      }
    }
  }
`;

type Props = {
  entryId: string;
  currentFlags: readonly AccessEntryFlag[];
};

export function EntryFlagSelect({ entryId, currentFlags }: Props) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const [open, setOpen] = useState(false);
  const [localFlags, setLocalFlags] = useState<AccessEntryFlag[]>([...currentFlags]);
  const openedWithRef = useRef<readonly AccessEntryFlag[]>(currentFlags);
  const [flagEntry] = useMutation<EntryFlagSelectMutation>(mutation);

  const toggleFlag = (flagValue: AccessEntryFlag) => {
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
              accessEntryId: entryId,
              flags: localFlags,
            },
          },
          onCompleted(_, errors) {
            if (errors?.length) {
              toast({
                title: __("Error"),
                description: formatError(
                  __("Failed to flag entry"),
                  errors as GraphQLError[],
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
                error as GraphQLError,
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
