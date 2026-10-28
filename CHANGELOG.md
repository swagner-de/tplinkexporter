# Changelog

## [0.5.0](https://github.com/swagner-de/tplinkexporter/compare/v0.4.0...v0.5.0) (2026-10-28)


### Features

* collect port speed/flow-control and switch info metrics ([4de5e7d](https://github.com/swagner-de/tplinkexporter/commit/4de5e7ddfe519fdb5cfba6dfc9bf611030f40eac))

## [0.4.0](https://github.com/swagner-de/tplinkexporter/compare/v0.3.0...v0.4.0) (2026-10-26)


### Features

* **chart:** add Helm chart for the exporter ([87b3aa2](https://github.com/swagner-de/tplinkexporter/commit/87b3aa2804ccb720f0548d25dd1235896ea4b480))
* **chart:** per-switch password via inline value or existing Secret ([e510b3e](https://github.com/swagner-de/tplinkexporter/commit/e510b3efdcd9db3af511dae2cecb41f5296226d1))
* **config:** add YAML config-file support with multi-switch schema ([d7c04aa](https://github.com/swagner-de/tplinkexporter/commit/d7c04aa58884e0630cf60f8702422f4a6b7ce46f))
* **config:** support per-switch password_file as an alternative to password ([1432ed1](https://github.com/swagner-de/tplinkexporter/commit/1432ed11525c6b57f235548a4e4b23ff774705b9))


### Bug Fixes

* **client:** add cookie jar, request timeout and robust error handling ([cb80315](https://github.com/swagner-de/tplinkexporter/commit/cb803154c3725142d0b160b79da406fae0d87353))
* **collector:** add scrape-health metric and multi-switch support ([d56a03b](https://github.com/swagner-de/tplinkexporter/commit/d56a03b518da0000d3f0061a8b39c8303c39e264))
* **deps:** rename module to swagner-de and bump dependencies ([62128e6](https://github.com/swagner-de/tplinkexporter/commit/62128e6835edd024661e138e197880ee7a3d2e68))

## [0.3.0](https://github.com/swagner-de/tplinkexporter/compare/v0.2.0...v0.3.0) (2026-10-26)


### Features

* **chart:** add Helm chart for the exporter ([7c3aa65](https://github.com/swagner-de/tplinkexporter/commit/7c3aa65b4f86081b98e1dc9203e443c6953d6418))
* **config:** support per-switch password_file as an alternative to password ([9fabb86](https://github.com/swagner-de/tplinkexporter/commit/9fabb86eee01c3a4bab14431b9c71783f6ad5387))

## [0.2.0](https://github.com/swagner-de/tplinkexporter/compare/v0.1.0...v0.2.0) (2026-10-26)


### Features

* **config:** add YAML config-file support with multi-switch schema ([8d0c736](https://github.com/swagner-de/tplinkexporter/commit/8d0c736cfffc74a8d899c312de4d2f65ce70a91f))


### Bug Fixes

* **client:** add cookie jar, request timeout and robust error handling ([3d75e7a](https://github.com/swagner-de/tplinkexporter/commit/3d75e7a064f36961000405c37cde85f468233399))
* **collector:** add scrape-health metric and multi-switch support ([a4203f9](https://github.com/swagner-de/tplinkexporter/commit/a4203f9af9a633f0deb15b3e91f07e38f1cfb920))
* **deps:** rename module to swagner-de and bump dependencies ([c648d51](https://github.com/swagner-de/tplinkexporter/commit/c648d51fb196e0b7ef729419c7147a11f7baacd5))
