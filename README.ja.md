# Software Package Information Retrieval and Analysis Tool
Software Package Information Retrieval and Analysis Tool (以下、SPIRAT と記載) は OS にインストールされているパッケージの情報を出力するコマンドラインツールです。

# 使い方
`spirat` コマンドを実行すると OS にインストールされているすべてのパッケージが `spirat.json` に SPDX 形式で出力されます。

```shell
$ spirat
$ jq '.' spirat.json | head -n30
{
  "spdxVersion": "SPDX-2.3",
  "dataLicense": "CC0-1.0",
  "SPDXID": "SPDXRef-DOCUMENT",
  "name": "spirat-generated-document",
  "documentNamespace": "",
  "creationInfo": {
    "licenseListVersion": "",
    "creators": [
      "Tool: spirat"
    ],
    "created": "2023-12-05T12:05:31+09:00"
  },
  "packages": [
    {
      "name": "libmagic1",
      "SPDXID": "SPDXRef-Package-libmagic1-5.41-3ubuntu0.1",
      "versionInfo": "5.41-3ubuntu0.1",
      "downloadLocation": "",
      "homepage": "https://www.darwinsys.com/file/",
      "sourceInfo": "http://archive.ubuntu.com/ubuntu jammy-updates/main amd64 Packages",
      "licenseDeclared": "BSD-2-Clause-regents and MIT-Old-Style-with-legal-disclaimer-2 and BSD-2-Clause-alike and public-domain and BSD-2-Clause-netbsd",
      "copyrightText": "",
      "externalRefs": [
        {
          "referenceCategory": "PACKAGE-MANAGER",
          "referenceType": "purl",
          "referenceLocator": "pkg:deb/ubuntu/libmagic1@5.41-3ubuntu0.1"
        }
      ]
```

`spirat -diff` コマンドで生成済みの SPDX 形式 JSON ファイルと今インストールされている情報の差分を出力することができます。

```shell
$ spirat
$ apt install vim
$ spirat -diff spirat.json
$ ls
spirat_diff.json  spirat.json
```

# 対応パッケージ
- deb
- rpm
- npm
