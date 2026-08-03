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

import { ProboElement } from "./base";
import type { ProboRootElement } from "./base";
import type { ProboCookieBannerRoot } from "./cookie-banner-root";
import type { LayoutButtons } from "../types";

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

class ProboHideableButton extends ProboActionButton {
  protected layoutButton: keyof LayoutButtons = "accept_all";

  private onReady = (): void => {
    this.applyVisibility();
  };

  connectedCallback(): void {
    super.connectedCallback();
    if (this.root) {
      if (this.root.layout) {
        this.applyVisibility();
      } else {
        this.root.addEventListener("probo-ready", this.onReady, { once: true });
      }
    }
  }

  disconnectedCallback(): void {
    super.disconnectedCallback();
    if (this.root) {
      this.root.removeEventListener("probo-ready", this.onReady);
    }
  }

  private applyVisibility(): void {
    const layout = this.root?.layout;
    if (layout && !layout.buttons[this.layoutButton]) {
      this.hidden = true;
    }
  }
}

export class ProboAcceptButton extends ProboActionButton {
  protected handleClick = (): void => {
    if (!this.root) return;
    this.root.client.acceptAll();
    this.root.setState("hidden");
    this.root.dispatchEvent(
      new CustomEvent("probo-consent", {
        bubbles: true,
        composed: true,
        detail: { action: "ACCEPT_ALL" },
      }),
    );
  };
}

export class ProboRejectButton extends ProboHideableButton {
  protected layoutButton = "reject_all" as const;

  protected handleClick = (): void => {
    if (!this.root) return;
    this.root.client.rejectAll();
    this.root.setState("hidden");
    this.root.dispatchEvent(
      new CustomEvent("probo-consent", {
        bubbles: true,
        composed: true,
        detail: { action: "REJECT_ALL" },
      }),
    );
  };
}

export class ProboCustomizeButton extends ProboHideableButton {
  protected layoutButton = "customize" as const;

  protected handleClick = (): void => {
    if (!this.root) return;
    this.root.setState("panel");
  };
}
