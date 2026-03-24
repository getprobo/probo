// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import type { PropsWithChildren } from "react";
import { tv } from "tailwind-variants";

const menuButtonVariants = tv({
  base: [
    "flex items-center gap-2 w-full",
    "px-2 py-1.5 text-sm rounded-sm bg-level-0 hover:bg-subtle cursor-pointer",
  ],
  variants: {
    active: {
      true: ["bg-active"],
    },
  },
});

type MenuButtonProps = {
  active?: boolean;
  onClick: () => void;
};

export function MenuButton({ children, active, onClick }: PropsWithChildren<MenuButtonProps>) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={menuButtonVariants({ active })}
    >
      {children}
    </button>
  );
}
