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

declare module "eslint-plugin-relay" {
  import type { Linter, Rule } from "eslint";

  export const rules: {
    "graphql-syntax": Rule.RuleModule;
    "graphql-naming": Rule.RuleModule;
    "generated-typescript-types": Rule.RuleModule;
    "no-future-added-value": Rule.RuleModule;
    "unused-fields": Rule.RuleModule;
    "must-colocate-fragment-spreads": Rule.RuleModule;
    "function-required-argument": Rule.RuleModule;
    "hook-required-argument": Rule.RuleModule;
  };

  export const configs: {
    recommended: {
      rules: Linter.RulesRecord;
    };
    "ts-recommended": {
      rules: Linter.RulesRecord;
    };
    strict: {
      rules: Linter.RulesRecord;
    };
    "ts-strict": {
      rules: Linter.RulesRecord;
    };
  };
}
