import { browser } from "@probo/eslint-config";

export default [
    { ignores: ["dist", "eslint.config.mjs", "*.config.{js,mjs,ts}"] },
    ...browser(["./tsconfig.app.json", "./tsconfig.node.json"], import.meta.dirname),
];
