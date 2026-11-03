import js from "@eslint/js";
import svelte from "eslint-plugin-svelte";
import tseslint from "typescript-eslint";
import globals from "globals";
import tailwind from "eslint-plugin-better-tailwindcss";

/** @type {import('eslint').Linter.Config[]} */
export default [
  js.configs.recommended,
  ...tseslint.configs.recommended,
  ...svelte.configs["flat/all"],
  // disables svelte formatting rules that conflict with prettier
  ...svelte.configs["flat/prettier"],
  {
    languageOptions: {
      globals: { ...globals.browser, ...globals.node },
    },
  },
  {
    files: ["**/*.svelte"],
    languageOptions: {
      parserOptions: { parser: tseslint.parser },
    },
    rules: {
      // no-navigation-without-resolve already enforces this, linter can't statically
      // see that resolve() prefixes with base so both rules can't be satisfied at once
      "svelte/no-navigation-without-base": "off",
      "svelte/no-goto-without-base": "off",
      // tailwind classes are generated at build time, not statically analysable
      "svelte/no-unused-class-name": "off",
      // ID selectors in component CSS are wrong (too specific), rule is misguided
      "svelte/consistent-selector-style": "off",
      // Spinner uses style directives for dynamic prop-driven dimensions
      "svelte/no-inline-styles": "off",
      // require lang="ts" rather than forbid it
      "svelte/block-lang": ["error", { script: ["ts"] }],
    },
  },
  {
    files: ["src/**/*.{svelte,ts}"],
    plugins: { "better-tailwindcss": tailwind },
    rules: {
      ...tailwind.configs["recommended-error"].rules,
      // prettier wraps classes
      "better-tailwindcss/enforce-consistent-line-wrapping": "off",
    },
    // the root size lets it turn pixel values like size-[22px] into scale classes like size-5.5
    settings: {
      "better-tailwindcss": { entryPoint: "src/app.css", rootFontSize: 16 },
    },
  },
  {
    ignores: [".svelte-kit/", "build/", ".wrangler/"],
  },
];
