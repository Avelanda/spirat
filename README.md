# Software Package Information Retrieval and Analysis Tool
Software Package Information Retrieval and Analysis Tool (abbreviated as SPIRAT) is a command-line tool that outputs information about packages installed on the operating system.

# Usage
Executing the `spirat` command will output all installed packages on the OS in SPDX format to a file named `spirat.json`.

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

Using the `spirat -diff` command, you can output the difference between the previously generated SPDX format JSON file
and the currently installed package information.

```shell
$ spirat
$ apt install vim
$ spirat -diff spirat.json
$ ls
spirat_diff.json  spirat.json
```

# Supported Package Formats
- deb
- rpm
- npm
