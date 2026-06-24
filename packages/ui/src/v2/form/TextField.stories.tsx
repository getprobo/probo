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

export function Variants() {
  return (
    <div className="flex w-60 flex-col gap-3">
      <TextField variant="classic" icon={<MagnifyingGlassIcon />} placeholder="Classic" />
      <TextField variant="surface" icon={<MagnifyingGlassIcon />} placeholder="Surface" />
      <TextField variant="soft" icon={<MagnifyingGlassIcon />} placeholder="Soft" />
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
      <span className="text-2 text-sand-11">
        Value:
        {value || "(empty)"}
      </span>
    </div>
  );
}

export function Skeleton() {
  return <TextFieldSkeleton />;
}
