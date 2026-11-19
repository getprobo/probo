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

export { ProboElement } from "./base.js";
export type { ProboState, ProboRootElement, ConsentDraft } from "./base.js";
export { ProboBanner } from "./banner.js";
export {
  ProboAcceptButton,
  ProboAcknowledgeButton,
  ProboCustomizeButton,
  ProboRejectButton,
} from "./buttons.js";
export { ProboCategory } from "./category.js";
export { ProboCategoryList } from "./category-list.js";
export { ProboCategoryToggle } from "./category-toggle.js";
export { ProboCookieBannerRoot } from "./cookie-banner-root.js";
export { ProboCookie, ProboCookieList } from "./cookie-list.js";
export { ProboPreferencePanel, ProboSaveButton } from "./preference-panel.js";
export { ProboPrivacyChoices } from "./privacy-choices.js";
export { ProboSettingsLink } from "./settings-link.js";
export { registerHeadlessComponents } from "./register.js";
