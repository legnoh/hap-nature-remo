# hap-nature-remo

[![Static Badge](https://img.shields.io/badge/homebrew-legnoh%2Fetc%2Fhap--nature--remo-orange?logo=apple)](https://github.com/legnoh/homebrew-etc/blob/main/Formula/hap-nature-remo.rb)
[![Static Badge](https://img.shields.io/badge/image-ghcr.io%2Flegnoh%2Fhap--nature--remo-blue?logo=github)](https://github.com/legnoh/hap-nature-remo/pkgs/container/hap-nature-remo)

This app provides homekit virtual devices defined by [Nature Remo](https://nature.global/nature-remo/).

## Usage

Install, init, and start. That's it !

All configs are provided from `~/.hap-nature-remo/config.yml` file.  
Create a configuration file with the following command and edit it.

- Config sample: [`sample/configs.yml`](./cmd/sample/configs.yml).

### macOS

```sh
# install
brew install legnoh/etc/hap-nature-remo

# init & edit
hap-nature-remo init
vi ~/.hap-nature-remo/config.yml

# start
brew services start hap-nature-remo
```

### Docker

> **Warning**
> This app does not work when running in Docker for Mac or Docker for Windows due to [this](https://github.com/docker/for-mac/issues/68) and [this](https://github.com/docker/for-win/issues/543).

```sh
# pull
docker pull ghcr.io/legnoh/hap-nature-remo

# init
docker run \
    -v .:/root/.hap-nature-remo/ \
    ghcr.io/legnoh/hap-nature-remo init

# edit
vi config.yml

# start
docker run \
    --network host \
    -v "./config.yml:/root/.hap-nature-remo/config.yml" \
    ghcr.io/legnoh/hap-nature-remo
```
