# dxregress

Regression tester for BlocknetDX.

# Commands

The tool looks for a `.dxregress.yml` config at `$HOME/.dxregress.yml`.
Use `--config=` to load configurations of various trading setups.

# Creating localenv

```
dxregress localenv up /path/to/codebase
dxregress localenv down /path/to/codebase
```

# Setup

* Docker CE is required: https://www.docker.com/community-edition#/download
* Docker Compose is required: https://docs.docker.com/compose/install/

`dxregress` utilizes glide package management https://glide.sh
To pull vendor source:

```
go get github.com/Masterminds/glide
cd /path/to/dxregress
glide up
```

# Copyright

```
// Copyright © 2017 The Blocknet Developers
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
```
