// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

export function displayValue(
  value: string | null | undefined,
  pendingLabel: string,
) {
  return value && value.length > 0 ? value : pendingLabel;
}

export function stateVariant(
  state: string,
): "success" | "danger" | "warning" | "info" {
  switch (state) {
    case "ACTIVE":
      return "success";
    case "REVOKED":
      return "danger";
    default:
      return "warning";
  }
}

type Translator = (key: string, options?: Record<string, unknown>) => string;

/** What a posture check observed, as returned by the DevicePostureValue type. */
export interface PostureValue {
  readonly kind: string;
  readonly text?: string | null;
  readonly number?: number | null;
}

export function postureValueLabel(t: Translator, value: PostureValue): string {
  switch (value.kind) {
    case "ON":
      return t("devices.postures.values.on");
    case "OFF":
      return t("devices.postures.values.off");
    case "IMMEDIATE":
      return t("devices.postures.values.immediate");
    case "SECONDS":
      return t("devices.postures.values.seconds", {
        seconds: value.number ?? 0,
      });
    case "MIN_PASSWORD_LENGTH":
      return t("devices.postures.values.minPasswordLength", {
        length: value.number ?? 0,
      });
    case "CONFIGURED":
      return t("devices.postures.values.configured");
    case "NONE":
      return t("devices.postures.values.none");
    case "TEXT":
      return value.text || t("devices.postures.values.unknown");
    default:
      return t("devices.postures.values.unknown");
  }
}

/**
 * Badge variant for a posture value. Kinds carrying a measurement render as
 * plain text instead: whether a 15 second delay or an 8 character minimum is
 * acceptable is a ruleset decision, not something the observation can say.
 */
export function postureValueVariant(
  kind: string,
  checkKey?: string,
): "success" | "danger" | "info" | undefined {
  if (checkKey === "REMOTE_LOGIN") {
    // Remote login reports reachability, so On is the exposed state.
    switch (kind) {
      case "ON":
        return "danger";
      case "OFF":
        return "info";
      default:
        return undefined;
    }
  }

  switch (kind) {
    case "ON":
    case "IMMEDIATE":
      return "success";
    case "OFF":
    case "NONE":
      return "danger";
    default:
      return undefined;
  }
}
