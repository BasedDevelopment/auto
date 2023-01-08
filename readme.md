# Auto
[![Go Report Card](https://goreportcard.com/badge/github.com/BasedDevelopment/auto)](https://goreportcard.com/report/github.com/BasedDevelopment/auto)
[![Build Status](https://github.com/BasedDevelopment/auto/actions/workflows/makefile.yml/badge.svg)](https://github.com/BasedDevelopment/auto/actions/)
[![CodeQL](https://github.com/BasedDevelopment/auto/workflows/CodeQL/badge.svg)](https://github.com/BasedDevelopment/auto/actions/workflows/codeql.yml)
[![License](https://img.shields.io/github/license/BasedDevelopment/eve?style=plastic)](https://github.com/BasedDevelopment/auto/blob/main/COPYING)


Auto is the agent that runs on hypervisors, it is responsible for communicating
with libvirt, setting up nftables, and preparing isos and disk images for use.

On first run, Auto will create a CSR (Certificate Signing Request), and send it
to eve to be signed. After the CSR is signed, Auto will save the certificate
and use that for future communications with eve.

## License

Copyright (C) 2022-2023  BNS Services LLC

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
