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

export { ProboElement } from "./base";
export type { ProboState, ProboRootElement, ConsentDraft } from "./base";
export { ProboBanner } from "./banner";
export {
  ProboAcceptButton,
  ProboCustomizeButton,
  ProboRejectButton,
} from "./buttons";
export { ProboCategory } from "./category";
export { ProboCategoryList } from "./category-list";
export { ProboCategoryToggle } from "./category-toggle";
export { ProboCookieBannerRoot } from "./cookie-banner-root";
export { ProboCookie, ProboCookieList } from "./cookie-list";
export { ProboPreferencePanel, ProboSaveButton } from "./preference-panel";
export { ProboSettingsButton } from "./settings-button";
export { registerComponents } from "./register";
