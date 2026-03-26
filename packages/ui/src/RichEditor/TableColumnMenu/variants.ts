// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import { tv } from "tailwind-variants";

export const tableColumnMenuVariants = tv({
  slots: {
    trigger: [
      "z-10 flex items-center justify-center",
      "rounded text-txt-tertiary bg-subtle hover:bg-border-solid cursor-grab",
      "py-0.5 h-3",
    ],
    menu: ["rounded-lg border border-border-mid bg-level-0 p-1 shadow-md z-20"],
  },
});
