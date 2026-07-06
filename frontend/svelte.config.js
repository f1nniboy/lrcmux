import adapterNode from "@sveltejs/adapter-node";
import adapterCloudflare from "@sveltejs/adapter-cloudflare";
import { vitePreprocess } from "@sveltejs/vite-plugin-svelte";

/** @type {import('@sveltejs/kit').Config} */
const config = {
  preprocess: vitePreprocess(),
  kit: {
    adapter: process.env.CF_PAGES ? adapterCloudflare() : adapterNode(),
    alias: {
      $components: "src/components",
    },
  },
};

export default config;
