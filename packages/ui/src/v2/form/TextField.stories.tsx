// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { MagnifyingGlassIcon } from "@phosphor-icons/react";
import { useState } from "react";

import { TextField } from "./TextField";
import { TextFieldSkeleton } from "./TextFieldSkeleton";

export default {
  title: "v2/form/TextField",
  component: TextField,
};

export function Default() {
  return (
    <div className="w-60">
      <TextField placeholder="Search..." />
    </div>
  );
}

export function WithIcon() {
  return (
    <div className="w-60">
      <TextField icon={<MagnifyingGlassIcon />} placeholder="Search..." />
    </div>
  );
}

export function Controlled() {
  const [value, setValue] = useState("");
  return (
    <div className="flex flex-col gap-3">
      <div className="w-60">
        <TextField value={value} onValueChange={setValue} placeholder="Type here" />
      </div>
      <span className="text-2 text-sand-11">Value: {value || "(empty)"}</span>
    </div>
  );
}

export function Skeleton() {
  return <TextFieldSkeleton />;
}
