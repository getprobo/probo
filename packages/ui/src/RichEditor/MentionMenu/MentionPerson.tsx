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

import { Avatar } from "../../v2/Avatar/Avatar";
import { Text } from "../../v2/typography/Text";
import type { MentionItem } from "../MentionExtension";

import { mentionMenu } from "./variants";

export function MentionPerson({
  item,
  highlighted = false,
}: {
  item: MentionItem;
  highlighted?: boolean;
}) {
  const { body, label, description } = mentionMenu();
  const color = highlighted ? "current" : "neutral";

  return (
    <>
      <Avatar
        size={1}
        radius="full"
        name={item.label}
        email={item.description ?? undefined}
        src={item.avatarUrl}
        alt=""
      />
      <span className={body()}>
        <Text
          size={2}
          weight="medium"
          highContrast={!highlighted}
          color={color}
          className={label()}
        >
          {item.label}
        </Text>
        {item.description && (
          <Text size={1} color={color} className={description()}>
            {item.description}
          </Text>
        )}
      </span>
    </>
  );
}
