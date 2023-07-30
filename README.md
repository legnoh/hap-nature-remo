# hap-nature-remo

[![Static Badge](https://img.shields.io/badge/homebrew-legnoh%2Fetc%2Fhap--nature--remo-orange?logo=apple)](https://github.com/legnoh/homebrew-etc/blob/main/Formula/hap-nature-remo.rb)
[![Static Badge](https://img.shields.io/badge/image-ghcr.io%2Flegnoh%2Fhap--nature--remo-blue?logo=github)](https://github.com/legnoh/hap-nature-remo/pkgs/container/hap-nature-remo)

<img width="400" alt="IMG_6950" src="https://github.com/legnoh/hap-nature-remo/assets/706834/a4a4e9ae-43b1-4948-ada6-641b75a35efe">

[Nature Remo](https://nature.global/nature-remo/) で登録したデバイスを HomeKit 経由で操作できるようになるアプリケーションです。  
現在、以下の操作に対応しています。

- エアコン
- リモコン式ファン(扇風機・シーリングファン)
- 内蔵センサー各種(温度・湿度・照度・人感) 

デバイスを個別に設定する必要がなく、設定ファイルにアクセストークンを指定するだけで、  
対応しているアクセサリーが自動的に検出されて登録されるため、気軽にお使い頂けます。

## Usage

インストール後、設定ファイルを所定のディレクトリに置いて起動するだけで使えます。

全ての設定ファイルは `~/.hap-nature-remo/config.yml` ファイルに設定します。  
下記のコマンドで設定ファイルを生成して、事前に変更しておいてください。

- サンプル: [`sample/configs.yml`](./cmd/sample/configs.yml).

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
> 下記の問題のため、Docker Desktop for Mac/Windows では Docker での起動はできません。
> [macOS](https://github.com/docker/for-mac/issues/68) / [Windows](https://github.com/docker/for-win/issues/543).

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

## 各デバイスごとの説明

### エアコン

|HomeKit|NatureRemo|
|--|--|
|<img width="300" alt="IMG_6946" src="https://github.com/legnoh/hap-nature-remo/assets/706834/3764c5c8-3414-4548-a25d-afb6e433086b">|<img width="300" alt="IMG_6951" src="https://github.com/legnoh/hap-nature-remo/assets/706834/5c35d7bd-e956-4227-993c-465a376a9a08">|


- 登録されている全てのエアコンが自動的に解釈されて登録されます。
- 現在対応しているのは冷房・暖房の2つです。
  - 除湿モードは、HomeKit が対応していないため、実装予定はありません。
  - 自動モードは、HomeKit 上での "自動" の概念と、エアコン各社の "自動モード" の概念が異なっていることが多く、現時点では利用できません。
- NatureRemo Nano など、温度計のない NatureRemo デバイスを利用しており、他に温度計がついているデバイスを利用している場合、別のデバイスの温度計を現在温度として代用するようになっています。
- スウィングについては未実装ですが、縦側の首振りが可能であればそのうち対応する予定です。
- 風量については、HomeKit 側に "風量: 自動" の概念がなく、日本の多くのエアコンと合致しない可能性が高いことから、実装予定はありません。

### 扇風機

|HomeKit|NatureRemo|
|--|--|
|<img width="300" alt="IMG_6948" src="https://github.com/legnoh/hap-nature-remo/assets/706834/7c7bdbd7-c881-4679-8fb2-69e9da4c082f">|<img width="300" alt="IMG_6949" src="https://github.com/legnoh/hap-nature-remo/assets/706834/9a8e4aee-0907-4e2d-bdd2-f8834409e03b">|

- リモコンとして登録され、かつ「扇風機」のアイコンがついているものを自動的にファンと解釈して登録されます。
- 電源オンオフ・風量調整・回転方向の設定に対応しています。
- ステート（現在設定の保持）は正確にはできないため、何度か設定してHomeアプリと状態を同期してからお使いください。
- 風向き・風量については、以下の通りにボタンを必ず配置してください。
  - 風向き
    - 風向き切替が1つのみ(トグル型)の場合は、早送り・巻き戻しボタンのいずれかを風向き切替ボタンとして登録しておいてください。
    - 風向き切替が2つそれぞれある場合は、早送り・巻き戻しボタンを両方とも風向き切替ボタンとして登録しておいてください。
  - 風量
    - オフを "0" として、そこからレベル別に 1(弱) ~ 10(強) のアイコンで風量のボタンを登録しておいてください。
    - 設定された解釈レベルに応じて、Home アプリ上で強さの指定ができるようになります。

### センサー各種(温度・湿度・照度・人感)

<img width="300" alt="IMG_6947" src="https://github.com/legnoh/hap-nature-remo/assets/706834/342e4fb3-d261-4e60-881f-bb506087f29f">

- センサー情報が取れる NatureRemo デバイスがあった場合、自動的に解釈されて登録されます。
- 温度・湿度・照度・人感 の4つのセンサーに対応しています。
- ただし、NatureRemo 側が Webhook などのデータを動的に送り込む機能を持たないこともあり、センサーで感知 -> 即時何かを行う、といったリアルタイム用途には使えません。
  - 現在の数値をHomeアプリ上で確認する、といった使い方がメインになります。
  - リアルタイム要素のトリガーが欲しい場合は公式アプリをご利用ください。
- 人感センサーはやや特殊で、5分以内に動作を感知した場合にのみ反応します。


## 注意事項

- デバイスの名前は Nature Remo でつけたものがそのまま引き継がれます。
- [Nature Remo Cloud API の利用制限](https://developer.nature.global/#リクエスト制限) を回避するため、同じリソースの取得を5秒に1回に絞るようにしています。
  - 基本的にはほぼ意識せずに使えると思いますが、あまり頻繁にコントロールを行うと、この制限に達する可能性があります。
  - homebrew で起動した際は、 `/opt/homebrew/var/log/hap-nature-remo.log` などで実行ログが見えますので、 `429 Too many Requests` が起こっていたら実行頻度を落とすようにしてください。
- オートメーションなどで指定した際、同時に複数デバイスへのリクエストが発生し、処理が安定しなくなることを防ぐために、意図的に揺らぎを持たせています。
  - 命令を受けてから最大5秒のスリープをランダムに仕込んでいます。
  - そのため、少し処理が遅く感じるかもしれませんが、ここは仕様のためご理解ください。
- アップデート時など、動作がおかしくなったときは起動時に 一度 `--reset` オプションをつけて起動すると改善することがあります。  
  (ただし、Homeアプリ上に設定したブリッジは削除する必要があり、オートメーションなども再設定が必要になります)
  ```sh
  hap-nature-remo serve --reset
  ```
- macOS では、起動時に下記のようなネットワークアクセスの許可を求めるダイアログが表示されますので、"Allow"を押してください。  
(アップデートを行う度に出現しますので注意してください)  
<img width="260" alt="Screenshot 2023-07-30 at 19 51 35" src="https://github.com/legnoh/hap-nature-remo/assets/706834/6ac1347b-ef64-4f2e-9a51-831bb2be3807">

## 余談

- このアプリを作ったのは、 NatureRemo Nano の Matter 対応では、センサーがついていない Home アプリでの見た目が微妙だったことが発端です。
  - FYI: [Matterでやりたかったけどできなかったこと - Nature Engineering Blog](https://engineering.nature.global/entry/blog-fes-2023-matter-matters)
- 温度センサーつきの Matter 対応 NatureRemo が発売され、対応デバイスが広がり次第、このアプリはアーカイブにする予定です。
- 素晴らしい製品を世に送り続けている Nature Inc. に心から感謝します。
