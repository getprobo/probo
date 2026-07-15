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

import { useState } from "react";

import { Button } from "../Button/Button";
import { Text } from "../typography/Text";

import { Dialog } from "./Dialog";
import { DialogBody } from "./DialogBody";
import { DialogClose } from "./DialogClose";
import { DialogDescription } from "./DialogDescription";
import { DialogFooter } from "./DialogFooter";
import { DialogHeader } from "./DialogHeader";
import { DialogPopup } from "./DialogPopup";
import { DialogSkeleton } from "./DialogSkeleton";
import { DialogTitle } from "./DialogTitle";
import { DialogTrigger } from "./DialogTrigger";

export default {
  title: "v2/Dialog",
  component: Dialog,
};

export function Default() {
  return (
    <Dialog>
      <DialogTrigger render={<Button variant="solid" color="neutral" highContrast>Open dialog</Button>} />
      <DialogPopup className="max-w-lg">
        <DialogHeader>
          <DialogTitle>Sign in to continue</DialogTitle>
          <DialogDescription>
            Access protected resources or submit a request.
          </DialogDescription>
        </DialogHeader>
        <DialogBody>
          <Text size={2} color="neutral">
            Dialog body content lives here — a form, a message, or any other
            slot the surrounding page composes.
          </Text>
        </DialogBody>
        <DialogFooter>
          <DialogClose render={<Button variant="soft" color="neutral" highContrast>Cancel</Button>} />
          <DialogClose render={<Button variant="solid" color="neutral" highContrast>Confirm</Button>} />
        </DialogFooter>
      </DialogPopup>
    </Dialog>
  );
}

export function Controlled() {
  const [open, setOpen] = useState(false);

  return (
    <div className="flex flex-col items-start gap-3">
      <Button variant="soft" color="neutral" highContrast onClick={() => setOpen(true)}>
        Open controlled dialog
      </Button>
      <Text size={1} color="faint">{open ? "open" : "closed"}</Text>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogPopup className="max-w-lg">
          <DialogHeader>
            <DialogTitle>Delete third party</DialogTitle>
            <DialogDescription>
              This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="soft" color="neutral" highContrast onClick={() => setOpen(false)}>
              Cancel
            </Button>
            <Button variant="solid" color="red" highContrast onClick={() => setOpen(false)}>
              Delete
            </Button>
          </DialogFooter>
        </DialogPopup>
      </Dialog>
    </div>
  );
}

// Shown over a simulated backdrop, matching how the dialog actually appears —
// the frame's elevation only reads correctly against the dimmed overlay.
export function Skeleton() {
  return (
    <div className="flex min-h-[420px] items-center justify-center rounded-4 bg-sand-12/40 p-8">
      <DialogSkeleton />
    </div>
  );
}
