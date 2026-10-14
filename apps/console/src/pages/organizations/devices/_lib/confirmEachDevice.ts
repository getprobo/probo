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

import type { MutationFeedback } from "#/lib/relay/useMutation";

interface ConfirmEachDeviceOptions {
  devices: readonly { id: string }[];
  successMessage: string;
  confirmDevice: (deviceId: string, feedback: MutationFeedback) => Promise<unknown>;
  onPartial: (completed: number, total: number) => void;
}

// Runs one mutation per device. Resolves true only when every device succeeds.
export async function confirmEachDevice(options: ConfirmEachDeviceOptions): Promise<boolean> {
  const { devices, successMessage, confirmDevice, onPartial } = options;
  let completed = 0;

  try {
    for (const [index, device] of devices.entries()) {
      const last = index === devices.length - 1;
      await confirmDevice(device.id, {
        // Empty string skips the toast; omitting it uses the hook default per device.
        successMessage: last ? successMessage : "",
      });
      completed += 1;
    }
    return true;
  } catch {
    if (completed > 0) {
      onPartial(completed, devices.length);
    }
    return false;
  }
}
