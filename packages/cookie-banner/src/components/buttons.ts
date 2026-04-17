// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

import { ProboElement } from "./base";
import type { ProboRootElement } from "./base";
import type { ProboCookieBannerRoot } from "./cookie-banner-root";

class ProboActionButton extends ProboElement {
  protected root: ProboRootElement | null = null;

  connectedCallback(): void {
    this.root = this.findAncestor<ProboCookieBannerRoot>("probo-cookie-banner-root");
    this.addEventListener("click", this.handleClick);
  }

  disconnectedCallback(): void {
    this.removeEventListener("click", this.handleClick);
  }

  protected handleClick = (_e: Event): void => {};
}

export class ProboAcceptButton extends ProboActionButton {
  protected handleClick = (): void => {
    if (!this.root) return;
    void this.root.client.acceptAll().then(() => {
      this.root!.setState("hidden");
      this.root!.dispatchEvent(
        new CustomEvent("probo-consent", {
          bubbles: true,
          composed: true,
          detail: { action: "ACCEPT_ALL" },
        }),
      );
    });
  };
}

export class ProboRejectButton extends ProboActionButton {
  protected handleClick = (): void => {
    if (!this.root) return;
    void this.root.client.rejectAll().then(() => {
      this.root!.setState("hidden");
      this.root!.dispatchEvent(
        new CustomEvent("probo-consent", {
          bubbles: true,
          composed: true,
          detail: { action: "REJECT_ALL" },
        }),
      );
    });
  };
}

export class ProboCustomizeButton extends ProboActionButton {
  protected handleClick = (): void => {
    if (!this.root) return;
    this.root.setState("panel");
  };
}
