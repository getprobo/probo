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
