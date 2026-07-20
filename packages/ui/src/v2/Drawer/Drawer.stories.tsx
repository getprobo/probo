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

import { Drawer } from "./Drawer";
import { DrawerBody } from "./DrawerBody";
import { DrawerClose } from "./DrawerClose";
import { DrawerDescription } from "./DrawerDescription";
import { DrawerFooter } from "./DrawerFooter";
import { DrawerHeader } from "./DrawerHeader";
import { DrawerPopup } from "./DrawerPopup";
import { DrawerSkeleton } from "./DrawerSkeleton";
import { DrawerTitle } from "./DrawerTitle";
import { DrawerTrigger } from "./DrawerTrigger";

export default {
  title: "v2/Drawer",
  component: Drawer,
};

export function Default() {
  return (
    <Drawer swipeDirection="right">
      <DrawerTrigger render={<Button variant="solid" color="neutral" highContrast>Open drawer</Button>} />
      <DrawerPopup side="right">
        <DrawerHeader>
          <DrawerTitle>Menu</DrawerTitle>
          <DrawerClose render={<Button variant="soft" color="neutral" highContrast>Close</Button>} />
        </DrawerHeader>
        <DrawerBody>
          <Text size={2} color="neutral">
            Drawer body content — navigation links, filters, or any other slot
            the surrounding page composes.
          </Text>
        </DrawerBody>
        <DrawerFooter>
          <DrawerClose render={<Button variant="solid" color="neutral" highContrast className="w-full">Done</Button>} />
        </DrawerFooter>
      </DrawerPopup>
    </Drawer>
  );
}

export function BottomSheet() {
  return (
    <Drawer swipeDirection="down">
      <DrawerTrigger render={<Button variant="soft" color="neutral" highContrast>Open bottom sheet</Button>} />
      <DrawerPopup side="bottom">
        <DrawerHeader>
          <DrawerTitle>Options</DrawerTitle>
          <DrawerClose render={<Button variant="soft" color="neutral" highContrast>Close</Button>} />
        </DrawerHeader>
        <DrawerBody>
          <DrawerDescription>
            Swipe down to dismiss.
          </DrawerDescription>
        </DrawerBody>
        <DrawerFooter>
          <DrawerClose render={<Button variant="solid" color="neutral" highContrast className="w-full">Confirm</Button>} />
        </DrawerFooter>
      </DrawerPopup>
    </Drawer>
  );
}

export function Controlled() {
  const [open, setOpen] = useState(false);

  return (
    <div className="flex flex-col items-start gap-3">
      <Button variant="soft" color="neutral" highContrast onClick={() => setOpen(true)}>
        Open controlled drawer
      </Button>
      <Text size={1} color="faint">{open ? "open" : "closed"}</Text>

      <Drawer open={open} onOpenChange={setOpen} swipeDirection="right">
        <DrawerPopup side="right">
          <DrawerHeader>
            <DrawerTitle>Account</DrawerTitle>
          </DrawerHeader>
          <DrawerBody>
            <Text size={2} color="neutral">
              Controlled the same way as Base UI — `open` / `onOpenChange`.
            </Text>
          </DrawerBody>
          <DrawerFooter>
            <Button variant="solid" color="neutral" highContrast className="w-full" onClick={() => setOpen(false)}>
              Close
            </Button>
          </DrawerFooter>
        </DrawerPopup>
      </Drawer>
    </div>
  );
}

export function Skeleton() {
  return (
    <div className="flex h-80 justify-end border border-sand-6 bg-sand-2">
      <DrawerSkeleton />
    </div>
  );
}
