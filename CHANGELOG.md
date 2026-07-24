# Changelog

## [1.7.0](https://github.com/f1nniboy/lrcmux/compare/v1.6.0...v1.7.0) (2026-07-24)


### Features

* **api/docs:** add provider ID column to providers table ([31863ff](https://github.com/f1nniboy/lrcmux/commit/31863ff389bf4355878a5d11f110f44eb62cb616))
* **isrc:** do multiple queries using both deezer's search operators and ([cb08fe6](https://github.com/f1nniboy/lrcmux/commit/cb08fe6d1eaaf1dfead18ca35062f1e1196347fe))


### Bug fixes

* **format/json:** don't leave out start timestamp for lines if it's 0 ([31863ff](https://github.com/f1nniboy/lrcmux/commit/31863ff389bf4355878a5d11f110f44eb62cb616))
* **isrc:** don't rely on deezer's search operators ([6ce81e4](https://github.com/f1nniboy/lrcmux/commit/6ce81e42d14c2ae48bcd24a4f341294974558092))
* **isrc:** duration shouldn't outweigh title+artist ([720a032](https://github.com/f1nniboy/lrcmux/commit/720a032b9dbff9f5a0b9d3ec6fb2ca2294db42c1))
* **isrc:** require minimum score, title score must be above 0 ([cd97030](https://github.com/f1nniboy/lrcmux/commit/cd97030a81389b5f482b0c0c7b870ad6ec9c04a5))
* **providers/kugou:** don't unconditionally strip first line ([4518534](https://github.com/f1nniboy/lrcmux/commit/4518534a763adc863470a47577f2a62eefe4f350))


### Refactors

* **api:** remove provider.hide option, always expose source ([31863ff](https://github.com/f1nniboy/lrcmux/commit/31863ff389bf4355878a5d11f110f44eb62cb616))
* **cache:** rename NotFound/Found to Miss/Hit, extract decode ([31863ff](https://github.com/f1nniboy/lrcmux/commit/31863ff389bf4355878a5d11f110f44eb62cb616))
* **frontend:** add icons, links and descriptions to app list ([e2faeaa](https://github.com/f1nniboy/lrcmux/commit/e2faeaab819a1f8e2670422a1fc7530c06b4b61c))
* **frontend:** new landing page with trending grid and app guides ([d8b83a5](https://github.com/f1nniboy/lrcmux/commit/d8b83a5bae31416942427c1f1ec1a391abaf0bb8))
* **providers/musixmatch:** scope rate-limit retry to get(), not ([1871632](https://github.com/f1nniboy/lrcmux/commit/1871632fb1e512ebb1524e26473ebf9e5fe33ab2))

## [1.6.0](https://github.com/f1nniboy/lrcmux/compare/v1.5.0...v1.6.0) (2026-07-14)


### Features

* **api:** add sources query param to include or exclude providers ([8c740f0](https://github.com/f1nniboy/lrcmux/commit/8c740f021ac8ffc0d087c2f05d4682b3c0f492d6))
* **frontend:** render source attribution as a clickable link when URL ([1b76cd1](https://github.com/f1nniboy/lrcmux/commit/1b76cd1b4d6cf861f816da94c434c3078919b072))
* **normalize:** add Match helper, use it across providers ([869a1e5](https://github.com/f1nniboy/lrcmux/commit/869a1e5c7b3a4c927a71729403895fa2c1f8df05))
* **normalize:** extend patterns, clean up funcs ([904e939](https://github.com/f1nniboy/lrcmux/commit/904e93935a497a74c8f6ba3c2dbd842004d398e2))
* **providers/netease:** add NetEase provider ([80205db](https://github.com/f1nniboy/lrcmux/commit/80205db4076221e58b0432bec46a2d5151fb5bcb))


### Bug fixes

* **isrc/deezer:** improve duration scoring with wider window ([8c740f0](https://github.com/f1nniboy/lrcmux/commit/8c740f021ac8ffc0d087c2f05d4682b3c0f492d6))
* **isrc:** remove primary artist fallback ([c6851f4](https://github.com/f1nniboy/lrcmux/commit/c6851f48fb3a5f0ae95f19954447fe268ff973c9))
* **lyrics:** preserve blank lines in ParsePlain as section separators ([b17f16f](https://github.com/f1nniboy/lrcmux/commit/b17f16f709eac0937d45a0353e3b3dab6949b81a))
* **normalize:** make String idempotent ([69b6ad9](https://github.com/f1nniboy/lrcmux/commit/69b6ad9534cafe8b1f8d27576a34f36d805f0f1e))
* **normalize:** strip brackets from artist string before splitting ([67fd990](https://github.com/f1nniboy/lrcmux/commit/67fd9904ff5616f11c346ba4d53aa7754e76ceb5))
* preserve empty-text lines as instrumental markers for LRC ([8c740f0](https://github.com/f1nniboy/lrcmux/commit/8c740f021ac8ffc0d087c2f05d4682b3c0f492d6))
* **providers/netease:** strip section markers ([28749ec](https://github.com/f1nniboy/lrcmux/commit/28749ecfe24154a61794c61bdd1a6808d615b738))


### Refactors

* **cache/redis:** remove unused logger ([c6851f4](https://github.com/f1nniboy/lrcmux/commit/c6851f48fb3a5f0ae95f19954447fe268ff973c9))
* **isrc:** clean up search result scoring in lookup ([f7a19b1](https://github.com/f1nniboy/lrcmux/commit/f7a19b1ae8e0da8bdba13c44d1a64e5cfb4bc938))
* **isrc:** remove track-to-ISRC cache layer ([8c740f0](https://github.com/f1nniboy/lrcmux/commit/8c740f021ac8ffc0d087c2f05d4682b3c0f492d6))
* **lyrics:** extract shared (e)LRC parser ([80205db](https://github.com/f1nniboy/lrcmux/commit/80205db4076221e58b0432bec46a2d5151fb5bcb))
* **normalize:** strip video suffixes and prod credits in Title ([8c740f0](https://github.com/f1nniboy/lrcmux/commit/8c740f021ac8ffc0d087c2f05d4682b3c0f492d6))
* **orchestrator:** split code into more logical files, expand ([df1de41](https://github.com/f1nniboy/lrcmux/commit/df1de415e96136d7936a3bcbf25895fa69a30150))
* **providers/genius:** always drop empty lines ([c6851f4](https://github.com/f1nniboy/lrcmux/commit/c6851f48fb3a5f0ae95f19954447fe268ff973c9))
* **providers/netease:** combine uncensor and halfWidth into one ([df1de41](https://github.com/f1nniboy/lrcmux/commit/df1de415e96136d7936a3bcbf25895fa69a30150))
* **providers/ytmusic:** always drop gap/empty lines ([c6851f4](https://github.com/f1nniboy/lrcmux/commit/c6851f48fb3a5f0ae95f19954447fe268ff973c9))
* **providers:** add URL to providers for easier source attribution ([1b76cd1](https://github.com/f1nniboy/lrcmux/commit/1b76cd1b4d6cf861f816da94c434c3078919b072))

## [1.5.0](https://github.com/f1nniboy/lrcmux/compare/v1.4.0...v1.5.0) (2026-07-10)


### Features

* **cache:** per-sync-level TTL by sync quality ([dd83afd](https://github.com/f1nniboy/lrcmux/commit/dd83afdf99269ffb3657a5db8e8ea6b45413ad25))
* **format:** add lyricsfile encoder, Extension() on all encoders ([5bbbf1f](https://github.com/f1nniboy/lrcmux/commit/5bbbf1fc0a1f7abc49ef8484c9847cdfbfb33bae))
* **frontend:** add /api/search server-side proxy for Deezer search ([bf79d1c](https://github.com/f1nniboy/lrcmux/commit/bf79d1cf4daa2fb1deb9a8144de4aac58c2afaa7))
* **frontend:** add navigation progress bar to layout ([bf79d1c](https://github.com/f1nniboy/lrcmux/commit/bf79d1cf4daa2fb1deb9a8144de4aac58c2afaa7))
* **frontend:** fetch trending client-side via proxy ([e20c9d8](https://github.com/f1nniboy/lrcmux/commit/e20c9d8e3775838ba532377861d78f53e41d8340))
* **frontend:** migrate to full SSR ([bf79d1c](https://github.com/f1nniboy/lrcmux/commit/bf79d1cf4daa2fb1deb9a8144de4aac58c2afaa7))
* **frontend:** rework landing page, add /apps page, SSR for lyrics page ([3d8de52](https://github.com/f1nniboy/lrcmux/commit/3d8de52f763a33142455e733f0aca6637134ed58))
* **orchestrator:** make fan-out tiered by sync level, simplify pick ([26ebec8](https://github.com/f1nniboy/lrcmux/commit/26ebec8eac51bbce35facd3b0ca36566c39840af))


### Bug fixes

* **api:** pass error message through instead of hardcoded strings ([26ebec8](https://github.com/f1nniboy/lrcmux/commit/26ebec8eac51bbce35facd3b0ca36566c39840af))
* **breaker:** use GetMany instead of per-key Get in states ([26ebec8](https://github.com/f1nniboy/lrcmux/commit/26ebec8eac51bbce35facd3b0ca36566c39840af))
* **format/lrc:** only emit headers for word-level sync ([e20c9d8](https://github.com/f1nniboy/lrcmux/commit/e20c9d8e3775838ba532377861d78f53e41d8340))
* **format/lrc:** remove metadata headers from output ([51b5ee7](https://github.com/f1nniboy/lrcmux/commit/51b5ee7c30a1c8820e1e30e33aaedcbf6f3d9da0))
* **frontend:** disable link preloading, was triggering on mobile touch ([bf79d1c](https://github.com/f1nniboy/lrcmux/commit/bf79d1cf4daa2fb1deb9a8144de4aac58c2afaa7))
* **frontend:** forward User-Agent through lyrics API requests ([bf79d1c](https://github.com/f1nniboy/lrcmux/commit/bf79d1cf4daa2fb1deb9a8144de4aac58c2afaa7))
* **providers/kugou:** reject censored results at provider level ([26ebec8](https://github.com/f1nniboy/lrcmux/commit/26ebec8eac51bbce35facd3b0ca36566c39840af))
* **providers/kugou:** trim trailing space from last word in line ([329a310](https://github.com/f1nniboy/lrcmux/commit/329a310b6acf08daee4b8bbd48dfa51e430e77f7))


### Refactors

* **frontend:** simplify lyrics page ([bf79d1c](https://github.com/f1nniboy/lrcmux/commit/bf79d1cf4daa2fb1deb9a8144de4aac58c2afaa7))

## [1.4.0](https://github.com/f1nniboy/lrcmux/compare/v1.3.0...v1.4.0) (2026-07-06)


### Features

* **api/compat:** add LRCLIB search endpoint, clean up KPOE response ([5ddf0ff](https://github.com/f1nniboy/lrcmux/commit/5ddf0ffcef1e4ad8566a114e76b4c761e8d93e34))
* **api:** set Cache-Control header, exclude internal IPs from metrics ([5e7b1a6](https://github.com/f1nniboy/lrcmux/commit/5e7b1a6b3eab24ed3721f6fd0403b75198b6f327))
* **frontend:** add setup guide for LRCGET ([5ddf0ff](https://github.com/f1nniboy/lrcmux/commit/5ddf0ffcef1e4ad8566a114e76b4c761e8d93e34))
* **frontend:** migrate from Fly.io to Cloudflare Workers ([4436752](https://github.com/f1nniboy/lrcmux/commit/4436752f1ddbf04eb74d77a5b8cdf75ab4cebbdc))
* **ratelimit:** replace sliding window with token bucket ([1f9731f](https://github.com/f1nniboy/lrcmux/commit/1f9731f38f8875c5d3ed66f1e4add22935d2349f))
* **ratelimit:** sliding window, and penalty box with exponential ([468fe14](https://github.com/f1nniboy/lrcmux/commit/468fe1400dad4f1e955cb555ae59ac2dec062a2c))


### Bug fixes

* **api:** redirect non-CF traffic ([ee589cc](https://github.com/f1nniboy/lrcmux/commit/ee589cc936896d4a9248b0db5021ef45342a7f39))
* **api:** set exposed CORS headers ([1f9731f](https://github.com/f1nniboy/lrcmux/commit/1f9731f38f8875c5d3ed66f1e4add22935d2349f))
* **api:** use sentry's own panic recover func to preserve stack ([2fd23ea](https://github.com/f1nniboy/lrcmux/commit/2fd23eaeaf4f8f8ec288e5a8e7fb50515b2c692c))
* **frontend:** add invisible pre element to determine height for ([f83ab68](https://github.com/f1nniboy/lrcmux/commit/f83ab68aec2bca599472d8a0c9d4f62802d79a30))
* **frontend:** show Retry-After seconds on 429 ([1f9731f](https://github.com/f1nniboy/lrcmux/commit/1f9731f38f8875c5d3ed66f1e4add22935d2349f))


### Documentation

* update deploy guide ([ee589cc](https://github.com/f1nniboy/lrcmux/commit/ee589cc936896d4a9248b0db5021ef45342a7f39))


### Refactors

* **api:** pass config to Server instead of individual fields ([177b722](https://github.com/f1nniboy/lrcmux/commit/177b722018de2ce7d9691327bd222247711b40aa))
* **orchestrator:** simplify breaker and outcome classification ([42e9e6a](https://github.com/f1nniboy/lrcmux/commit/42e9e6a8681e84b2fc1f1b4efcbd91acb71a3601))
* **providers:** replace registry/factory with struct embedding ([6484adc](https://github.com/f1nniboy/lrcmux/commit/6484adcd4f0e49ba7870ee6dd39a3e0ea1b73cb5))

## [1.3.0](https://github.com/f1nniboy/lrcmux/compare/v1.2.0...v1.3.0) (2026-07-05)


### Features

* **frontend:** cache artist top tracks in module scope to avoid refetch ([3820926](https://github.com/f1nniboy/lrcmux/commit/3820926fd47c0bec1acb998dde285dab26c777f5))
* **frontend:** replace lyrics container with textarea for Ctrl+A select ([b258bea](https://github.com/f1nniboy/lrcmux/commit/b258beab564bf5fd60b525b61c0041d8b460449c))
* **metrics:** add HTTP request latency histogram ([8c13ec9](https://github.com/f1nniboy/lrcmux/commit/8c13ec9ad99306e82c3ed98184e21a31fd2313c2))


### Bug fixes

* **api:** reduce graceful shutdown timeout ([67f42e6](https://github.com/f1nniboy/lrcmux/commit/67f42e69da8199511f5e3fe3cd23c22fbf315976))
* **format/lrc:** write header tags for all sync levels, not only word ([3476fcb](https://github.com/f1nniboy/lrcmux/commit/3476fcbe60b391ba4961b247cca08945c35587cd))
* **providers/genius:** strip role annotation lines from lyrics ([202b3c4](https://github.com/f1nniboy/lrcmux/commit/202b3c46c86fc11c3d9c7ba1ed114f467294b8f2))
* **proxy:** sticky proxy selection per provider call ([d87b17d](https://github.com/f1nniboy/lrcmux/commit/d87b17d1f3ff901ebc68e5673ddbf67e6c6422c0))


### Documentation

* **api:** add section about setting user agent ([30e5e17](https://github.com/f1nniboy/lrcmux/commit/30e5e178df1481a086e09ea9754d8762a1e0d866))


### Refactors

* **config:** migrate to go-toml/v2, add tool to generate example ([269405d](https://github.com/f1nniboy/lrcmux/commit/269405d38ec0bd03c1e601ac94887510eee48cd5))
* **providers/musixmatch:** swap token pool for signed URLs ([9444adc](https://github.com/f1nniboy/lrcmux/commit/9444adca77cb45dd8392403d6addbf5f4c0f38d8))

## [1.2.0](https://github.com/f1nniboy/lrcmux/compare/v1.1.0...v1.2.0) (2026-07-03)


### Features

* **frontend:** landing page redesign, compat guides, misc fixes ([952c487](https://github.com/f1nniboy/lrcmux/commit/952c4875a6f566da7235d3fee65c2e3b5287075d))


### Bug fixes

* **frontend:** extend lyrics panel to fill viewport, misc clean up ([051ca71](https://github.com/f1nniboy/lrcmux/commit/051ca71070da2122b96433f3da39dcee3a6a8c40))
* **orchestrator:** include sync level in singleflight group key ([8aeff87](https://github.com/f1nniboy/lrcmux/commit/8aeff8798ddd72abdeff4e8253c98fc3aeb0488e))
* **orchestrator:** penalize censored results in provider ranking ([ec9f528](https://github.com/f1nniboy/lrcmux/commit/ec9f5289a27c52e96113c8d434d9d6cd49fe3b27))

## [1.1.0](https://github.com/f1nniboy/lrcmux/compare/lrcmux-v1.0.0...lrcmux-v1.1.0) (2026-07-03)


### Features

* **frontend:** landing page redesign, compat guides, misc fixes ([952c487](https://github.com/f1nniboy/lrcmux/commit/952c4875a6f566da7235d3fee65c2e3b5287075d))


### Bug fixes

* **frontend:** extend lyrics panel to fill viewport, misc clean up ([051ca71](https://github.com/f1nniboy/lrcmux/commit/051ca71070da2122b96433f3da39dcee3a6a8c40))
* **orchestrator:** include sync level in singleflight group key ([8aeff87](https://github.com/f1nniboy/lrcmux/commit/8aeff8798ddd72abdeff4e8253c98fc3aeb0488e))
* **orchestrator:** penalize censored results in provider ranking ([ec9f528](https://github.com/f1nniboy/lrcmux/commit/ec9f5289a27c52e96113c8d434d9d6cd49fe3b27))
