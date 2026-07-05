# Changelog

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
