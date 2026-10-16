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
  presenceColor,
  type RichEditorCollaborationHandle,
  type RichEditorPresence,
} from "@probo/ui";
import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";
import { Tooltip } from "@probo/ui/src/v2/Tooltip/Tooltip";
import { TooltipPopup } from "@probo/ui/src/v2/Tooltip/TooltipPopup";
import { TooltipTrigger } from "@probo/ui/src/v2/Tooltip/TooltipTrigger";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { tv } from "tailwind-variants/lite";

import {
  type DocumentPresencePerson,
  groupPresencesByIdentity,
} from "../_lib/documentPresence";

const visibleLimit = 5;

const facepile = tv({
  slots: {
    root: "flex items-center",
    stack: "flex items-center -space-x-2",
    avatarWrap: "rounded-full bg-sand-1 p-px",
    overflow:
      "inline-flex size-8 items-center justify-center rounded-full bg-sand-3 text-1 font-medium text-sand-11 ring-2 ring-sand-1",
  },
  variants: {
    active: {
      true: { avatarWrap: "" },
      false: { avatarWrap: "opacity-50 grayscale" },
    },
  },
});

type DocumentPresenceProps = {
  collaborationHandle: RichEditorCollaborationHandle;
};

export function DocumentPresence({ collaborationHandle }: DocumentPresenceProps) {
  const { t } = useTranslation();
  const [presences, setPresences] = useState<RichEditorPresence[]>([]);
  const slots = facepile();

  useEffect(() => {
    return collaborationHandle.onPresence?.(setPresences);
  }, [collaborationHandle]);

  const people = groupPresencesByIdentity(presences);
  if (people.length === 0) return null;

  const visible = people.slice(0, visibleLimit);
  const overflow = people.slice(visibleLimit);

  return (
    <div className={slots.root()} aria-label={t("documentDescriptionPage.presence.label")}>
      <div className={slots.stack()}>
        {visible.map(person => {
          const name = person.displayName?.trim()
            || t("documentDescriptionPage.presence.anonymous");
          const label = person.active
            ? name
            : t("documentDescriptionPage.presence.inactive", { name });
          return (
            <PresenceAvatar
              key={person.identityID}
              person={person}
              label={label}
            />
          );
        })}
        {overflow.length > 0 && (
          <Tooltip>
            <TooltipTrigger
              className={slots.overflow()}
              aria-label={t("documentDescriptionPage.presence.overflow", {
                count: overflow.length,
              })}
            >
              +{overflow.length}
            </TooltipTrigger>
            <TooltipPopup>
              <ul className="flex flex-col gap-1">
                {overflow.map(person => (
                  <li key={person.identityID}>
                    <Text size={1}>
                      {person.active
                        ? (person.displayName?.trim()
                          || t("documentDescriptionPage.presence.anonymous"))
                        : t("documentDescriptionPage.presence.inactive", {
                            name: person.displayName?.trim()
                              || t("documentDescriptionPage.presence.anonymous"),
                          })}
                    </Text>
                  </li>
                ))}
              </ul>
            </TooltipPopup>
          </Tooltip>
        )}
      </div>
    </div>
  );
}

function PresenceAvatar({
  person,
  label,
}: {
  person: DocumentPresencePerson;
  label: string;
}) {
  const slots = facepile({ active: person.active });
  const ring = person.active
    ? { boxShadow: `0 0 0 2px ${presenceColor(person.identityID)}` }
    : undefined;

  return (
    <Tooltip>
      <TooltipTrigger
        className={slots.avatarWrap()}
        style={ring}
        aria-label={label}
      >
        <Avatar
          name={person.displayName?.trim() || undefined}
          src={person.avatarURL}
          size={2}
          radius="full"
          alt={label}
        />
      </TooltipTrigger>
      <TooltipPopup>{label}</TooltipPopup>
    </Tooltip>
  );
}
