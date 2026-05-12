// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import { tv } from "tailwind-variants";

export const tableRowMenuVariants = tv({
  slots: {
    trigger: [
      "z-20 flex flex-col items-center justify-center",
      "rounded text-txt-tertiary bg-subtle hover:bg-border-solid cursor-grab",
      "px-0.5 w-2",
    ],
    menu: ["rounded-lg border border-border-mid bg-level-0 p-1 shadow-mid z-30"],
  },
});
