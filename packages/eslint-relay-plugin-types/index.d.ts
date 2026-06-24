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
