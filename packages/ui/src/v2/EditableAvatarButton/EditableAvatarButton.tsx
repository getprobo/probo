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

import { CameraIcon, PencilSimpleIcon } from "@phosphor-icons/react";

import { Avatar } from "../Avatar/Avatar";

import { editableAvatarButton } from "./variants";

export interface EditableAvatarButtonProps {
  fullName: string;
  email: string;
  src?: string | null;
  label: string;
  size?: 1 | 2 | 3;
  radius?: "small" | "full";
  onClick: () => void;
}

export function EditableAvatarButton({
  fullName,
  email,
  src,
  label,
  size = 1,
  radius = "small",
  onClick,
}: EditableAvatarButtonProps) {
  const slots = editableAvatarButton({ radius });

  return (
    <button
      type="button"
      className={slots.root()}
      onClick={onClick}
      aria-label={label}
    >
      <span className={slots.frame()}>
        <Avatar
          name={fullName}
          email={email}
          src={src}
          size={size}
          radius={radius}
        />
        <span className={slots.overlay()} aria-hidden>
          <CameraIcon className="size-3" />
        </span>
      </span>
      <span className={slots.badge()} aria-hidden>
        <PencilSimpleIcon className="size-2.5" weight="bold" />
      </span>
    </button>
  );
}
