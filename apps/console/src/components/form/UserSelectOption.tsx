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

import { Avatar } from "@probo/ui/src/v2/Avatar/Avatar";

interface UserSelectOptionProps {
  fullName: string;
  emailAddress?: string | null;
  avatarUrl?: string | null;
  // Matches the select trigger. Size 1 is a 24px control, so the avatar is
  // 16px; size 2 is 32px, so the avatar stays 24px.
  size?: 1 | 2;
}

export function UserSelectOption({
  fullName,
  emailAddress,
  avatarUrl,
  size = 2,
}: UserSelectOptionProps) {
  return (
    <span className={size === 1
      ? "flex min-w-0 items-center gap-1"
      : "flex min-w-0 items-center gap-2"}
    >
      <Avatar
        size={1}
        radius="full"
        name={fullName}
        email={emailAddress ?? undefined}
        src={avatarUrl}
        alt=""
        // Avatar size 1 is 24px. The important utility wins over that class
        // so the compact trigger can inset the picture.
        className={size === 1 ? "size-4!" : undefined}
      />
      <span className="truncate">{fullName}</span>
    </span>
  );
}
