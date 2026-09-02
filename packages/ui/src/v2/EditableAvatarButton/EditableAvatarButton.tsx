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

import { CameraIcon } from "@phosphor-icons/react";
import { avatarFallback } from "@probo/helpers";
import type { ReactNode } from "react";

import { Avatar } from "../Avatar/Avatar";

import { editableAvatarButton } from "./variants";

export interface EditableAvatarButtonProps {
  fullName: string;
  src?: string | null;
  fallback?: ReactNode;
  label: string;
  size?: 1 | 2 | 3;
  radius?: "small" | "full";
  onClick: () => void;
}

export function EditableAvatarButton({
  fullName,
  src,
  fallback,
  label,
  size = 1,
  radius = "small",
  onClick,
}: EditableAvatarButtonProps) {
  const slots = editableAvatarButton();

  return (
    <button
      type="button"
      className={slots.root()}
      onClick={onClick}
      aria-label={label}
    >
      <Avatar
        size={size}
        variant="soft"
        color="gold"
        radius={radius}
        src={src ?? undefined}
        alt={fullName}
        fallback={fallback ?? avatarFallback(fullName)}
      />
      <span className={slots.overlay()}>
        <CameraIcon className="size-3" />
      </span>
    </button>
  );
}
