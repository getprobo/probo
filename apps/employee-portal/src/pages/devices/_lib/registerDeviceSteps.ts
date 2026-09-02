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

export const AGENT_INSTALL_URL = "https://pro.bo/install";

export const REGISTER_DEVICE_STEPS = ["review", "download", "enroll"] as const;

export type RegisterDeviceStep = (typeof REGISTER_DEVICE_STEPS)[number];

export function parseRegisterDeviceStep(value: string | null): RegisterDeviceStep {
  if (value === "download" || value === "enroll") {
    return value;
  }

  return "review";
}

export function registerDeviceStepIndex(step: RegisterDeviceStep): number {
  return REGISTER_DEVICE_STEPS.indexOf(step);
}

export function maxRegisterDeviceStep(
  left: RegisterDeviceStep,
  right: RegisterDeviceStep,
): RegisterDeviceStep {
  return registerDeviceStepIndex(left) >= registerDeviceStepIndex(right) ? left : right;
}
