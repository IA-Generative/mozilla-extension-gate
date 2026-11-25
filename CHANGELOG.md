# Changelog

## [1.5.2](https://github.com/IA-Generative/mozilla-extension-gate/compare/v1.5.1...v1.5.2) (2026-11-25)


### Bug Fixes

* filter out draft and prerelease GitHub releases in ListReleases ([1b09181](https://github.com/IA-Generative/mozilla-extension-gate/commit/1b091810eed4e78ffc42fdca0d9384d024df6f81))

## [1.5.1](https://github.com/IA-Generative/mozilla-extension-gate/compare/v1.5.0...v1.5.1) (2026-11-19)


### Bug Fixes

* update alias configuration reference in configmap.yaml ([5d48407](https://github.com/IA-Generative/mozilla-extension-gate/commit/5d48407ec74e6f6081d687b7b7223fdb3ec44322))
* update GetLatestArtifact to return version alongside artifact data ([f436c48](https://github.com/IA-Generative/mozilla-extension-gate/commit/f436c481c3cf1d516039c84c492d0d2924d482e4))

## [1.5.0](https://github.com/IA-Generative/mozilla-extension-gate/compare/v1.4.1...v1.5.0) (2026-11-19)


### Features

* add extension alias support and implement latest artifact retrieval ([7671595](https://github.com/IA-Generative/mozilla-extension-gate/commit/767159517e8d3d3057a8ab409b2d1294bd9359f0))


### Bug Fixes

* add error logging for JSON marshaling in CreateArtifactManifest method ([e8af6a6](https://github.com/IA-Generative/mozilla-extension-gate/commit/e8af6a6f03320db85da50e669031d8c6958fb328))
* adjust storageClassName in PVC configuration ([a5758dd](https://github.com/IA-Generative/mozilla-extension-gate/commit/a5758dd4de0685c41dffac3f4b8eb065aedc0e94))
* convert release ID to string using strconv.FormatInt ([b96f0a3](https://github.com/IA-Generative/mozilla-extension-gate/commit/b96f0a337d1b5ccdfb2ee7e55ae9aac0cc24ab90))
* handle errors when closing response bodies in various providers ([b157c7e](https://github.com/IA-Generative/mozilla-extension-gate/commit/b157c7e8ae8aa22394a4413eee4207bcc79c7c20))
* handle errors when creating directories and syncing manifest service ([bde0717](https://github.com/IA-Generative/mozilla-extension-gate/commit/bde07174d57a25277139476279a7b5f242f63070))
* log errors when closing response bodies in ListReleaseAssets and FetchReleaseAsset ([fe414cf](https://github.com/IA-Generative/mozilla-extension-gate/commit/fe414cf6cda2b6af9e11792d99ac33c1d9c77724))
* remove unnecessary labels from PersistentVolumeClaim metadata ([f11d507](https://github.com/IA-Generative/mozilla-extension-gate/commit/f11d50757c7801f8a115fb051a6073505c891ffc))
* remove unused regex and string sanitization functions from API package ([e8af6a6](https://github.com/IA-Generative/mozilla-extension-gate/commit/e8af6a6f03320db85da50e669031d8c6958fb328))

## [1.4.1](https://github.com/IA-Generative/mozilla-extension-gate/compare/v1.4.0...v1.4.1) (2026-11-18)


### Bug Fixes

* correct indentation for storageClassName in PVC configuration ([3715f5b](https://github.com/IA-Generative/mozilla-extension-gate/commit/3715f5bae9bd1786c19a45f066c8e374f78beb0c))

## [1.4.0](https://github.com/IA-Generative/mozilla-extension-gate/compare/v1.3.0...v1.4.0) (2026-11-18)


### Features

* add CreateBucketIfNotExists option to S3 configuration and implement bucket creation logic ([b2964e8](https://github.com/IA-Generative/mozilla-extension-gate/commit/b2964e8a26e6979b1c4948e9cff7fd9a6bbe82b4))
* implement file storage provider and update configuration for file-based storage ([cbc70e3](https://github.com/IA-Generative/mozilla-extension-gate/commit/cbc70e37e0ecf306dc1a4f1428e9c2d74d3c92d8))


### Bug Fixes

* reorganize middleware and path handling in API router ([717fede](https://github.com/IA-Generative/mozilla-extension-gate/commit/717fedea8e75a15d412abb45d228d12d456da80d))
* update init container to use dynamic image and simplify CA certificate handling ([3cd0807](https://github.com/IA-Generative/mozilla-extension-gate/commit/3cd0807d72221d71b5047bb6021bd2d178d417a1))
* update URL assignment for release assets in conversion function ([d50b20d](https://github.com/IA-Generative/mozilla-extension-gate/commit/d50b20d685de9aa0538f649bad7be70d079459d0))

## [1.3.0](https://github.com/IA-Generative/mozilla-extension-gate/compare/v1.2.0...v1.3.0) (2026-11-18)


### Features

* add TLS configuration options for S3 and GitHub clients ([0d9b91c](https://github.com/IA-Generative/mozilla-extension-gate/commit/0d9b91cffaf5330ee4a10d4e4e12c3b905596ab3))


### Bug Fixes

* update BASE_PATH to empty string and adjust health check paths in deployment ([8b94ec2](https://github.com/IA-Generative/mozilla-extension-gate/commit/8b94ec21cec4adfffc3e1b390cb8cac0cbf060ff))

## [1.2.0](https://github.com/IA-Generative/mozilla-extension-gate/compare/v1.1.0...v1.2.0) (2026-11-18)


### Features

* add support for extra CA certificates and HTTP proxy settings in deployment ([3e001b5](https://github.com/IA-Generative/mozilla-extension-gate/commit/3e001b5e3c32c3dc55c72e8a6367d992d126c04d))


### Bug Fixes

* update health check paths for liveness and readiness probes ([7e77c39](https://github.com/IA-Generative/mozilla-extension-gate/commit/7e77c3943eb8071cd4decc339c2605d5f4aaabac))

## [1.1.0](https://github.com/IA-Generative/mozilla-extension-gate/compare/v1.0.0...v1.1.0) (2026-11-17)


### Features

* init prroject ([c868883](https://github.com/IA-Generative/mozilla-extension-gate/commit/c868883ce55b721a726864657fc1e1137c40b78f))
* update image repository and add CI/CD workflows ([43dcb15](https://github.com/IA-Generative/mozilla-extension-gate/commit/43dcb154c4552076031a2b335b4f6abed26c202a))


### Bug Fixes

* explicitly specify release-please config and manifest file paths ([2a0a27b](https://github.com/IA-Generative/mozilla-extension-gate/commit/2a0a27b070791beef7b27be89f35705b21ff5de6))
