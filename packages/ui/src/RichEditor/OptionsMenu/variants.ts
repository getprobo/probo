// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import { tv } from "tailwind-variants";

export const optionsMenuVariants = tv({
  slots: {
    trigger: [
      "z-20 flex size-6 items-center justify-center",
      "rounded text-txt-tertiary hover:bg-subtle hover:text-txt-primary cursor-grab",
    ],
    menu: ["rounded-lg border border-border-mid bg-level-0 p-1 shadow-md z-30"],
  },
});
