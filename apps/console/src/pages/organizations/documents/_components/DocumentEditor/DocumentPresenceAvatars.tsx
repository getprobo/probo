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

import { groupPresencesByIdentity, type PresencePerson } from "@probo/collaboration";
import {
  presenceColor,
  type RichEditorCollaborationHandle,
  type RichEditorPresence,
} from "@probo/ui/src/RichEditor/presence";
import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";
import { Separator } from "@probo/ui/src/v2/Separator/Separator";
import { Tooltip } from "@probo/ui/src/v2/Tooltip/Tooltip";
import { TooltipPopup } from "@probo/ui/src/v2/Tooltip/TooltipPopup";
import { TooltipTrigger } from "@probo/ui/src/v2/Tooltip/TooltipTrigger";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";

import { documentPresence } from "../../variants";

const visibleLimit = 5;

interface DocumentPresenceAvatarsProps {
  collaborationHandle: RichEditorCollaborationHandle;
}

// The people connected to the document, one avatar per person across their
// tabs. Each avatar ring matches the color of that person's caret. The
// separator before the mode select comes with the stack, so it never shows
// alone.
export function DocumentPresenceAvatars({ collaborationHandle }: DocumentPresenceAvatarsProps) {
  const { t } = useTranslation("organizations/documents");
  const [presences, setPresences] = useState<RichEditorPresence[]>([]);
  const { stack, avatarTrigger, overflow, separator } = documentPresence();

  useEffect(() => {
    return collaborationHandle.onPresence?.(setPresences);
  }, [collaborationHandle]);

  const people = groupPresencesByIdentity(presences);
  if (people.length === 0) {
    return null;
  }

  const visible = people.slice(0, visibleLimit);
  const hidden = people.slice(visibleLimit);
  const labelOf = (person: PresencePerson) => {
    const name = person.displayName?.trim() || t("details.presence.anonymous");
    return person.active ? name : t("details.presence.inactive", { name });
  };

  return (
    <>
      <div className={stack()} role="group" aria-label={t("details.presence.label")}>
        {visible.map((person) => {
          const label = labelOf(person);
          return (
            <Tooltip key={person.identityID}>
              <TooltipTrigger
                className={avatarTrigger({ active: person.active })}
                style={person.active
                  ? { boxShadow: `0 0 0 2px ${presenceColor(person.identityID)}` }
                  : undefined}
                aria-label={label}
              >
                <Avatar
                  name={person.displayName?.trim() || undefined}
                  src={person.avatarURL}
                  size={1}
                  radius="full"
                  alt={label}
                />
              </TooltipTrigger>
              <TooltipPopup>{label}</TooltipPopup>
            </Tooltip>
          );
        })}
        {hidden.length > 0 && (
          <Tooltip>
            <TooltipTrigger
              className={overflow()}
              aria-label={t("details.presence.overflow", { count: hidden.length })}
            >
              {`+${hidden.length}`}
            </TooltipTrigger>
            <TooltipPopup>
              <ul>
                {hidden.map(person => (
                  <li key={person.identityID}>
                    <Text size={1}>{labelOf(person)}</Text>
                  </li>
                ))}
              </ul>
            </TooltipPopup>
          </Tooltip>
        )}
      </div>
      <Separator orientation="vertical" className={separator()} />
    </>
  );
}
