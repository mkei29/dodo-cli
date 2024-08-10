この章では実際にドキュメントをアップロードするまでの一連の流れを説明します。

事前にサインアップとプロジェクトの作成を完了させている必要があります。
まだこれらの作業を完了していない場合には、以下のドキュメントに従い先にサインアップとプロジェクトの作成を完了させてください。

https://document.do.dodo-doc.com/install

# ドキュメントのアップロードまでの流れ
ドキュメントをアップロードするまでの大まかな流れは以下の通りです。

* プロジェクト画面から新しいAPI Keyを発行して環境変数に設定する
* `dodo-cli init`コマンドで設定ファイルを生成する。
* `dodo-cli upload`コマンドでドキュメントをアップロードする

では具体的な作業を深堀りしましょう。

# API Keyの作成
API Keyはdodo-cliを実行したユーザーが適切な権限を持っていることを確認するために必要になります。
まずはブラウザでdodoにログインして新しいAPI Keyを発行しましょう。

API Keyは各プロジェクトの画面から発行することができます。
ダッシュボードからアップロードしたいプロジェクトをクリックしてプロジェクト画面を開いてください。

https://www.dodo-doc.com/dashboard

プロジェクト画面の右上にある`New API Key`ボタンを押すと、新しいAPI Keyが発行されます。
画面上部に発行されたAPI Keyが表示されるのでコピーして保管してください。

!important
API Keyは一度しか表示されず、画面を閉じると二度と確認することができません。

!important
API Keyはインターネットなどに公開しないでください。
API Keyが流出するとドキュメントの内容を改ざんされる可能性があります。

# 設定ファイルの作成
次にdodo用の設定ファイルを作成します。
`dodo-cli init`コマンドを使うことで簡単に設定ファイルの雛形を作成することができます。

まずはgitレポジトリのルートに移動して、`dodo-cli`コマンドを実行してください。
対話形式で何個か質問されるので回答してください。
質問への回答が終わると新しく`.dodo.yaml`という設定ファイルが生成されます。

```yaml
version: 1
project:
  name: testdoc
  version: 1
  description: test description
pages:
  - markdown: README.md
    path: "/README"
    title: "README"
  ## Create the directory and place all markdown files in the docs
  #- directory: "Directory"
  #  children:
  #    - match: "docs/*.md"
```

デフォルトではREADME.mdがトップページになるように設定されています。
必要に応じて[設定ファイルの仕様](https://document.do.dodo-doc.com/yaml_specification)を参照してpagesフィールドを書き換えてください。
設定ファイルの準備ができたら最後にアップロードしましょう。

# ドキュメントのアップロード
ドキュメントをアップロードするためには、環境変数`DOOO_API_KEY`に最初に取得したAPI Keyをセットする必要があります。
以下のコマンドを実行してAPI Keyを環境変数にセットしてください。

```bash
export DODO_API_KEY="<最初に取得したAPI Key>"
```

!note
ローカル環境から継続的にアップロードする場合にはdirenvなどを利用すると便利です。

これでアップロードするための準備が整いました。
以下のコマンドを実行して、実際にドキュメントをアップロードしてみましょう。

```bash
dodo-cli upload
```

成功すれば`successfully uploaded`というログメッセージが出力されます。
またドキュメントもURLが表示されるので、ブラウザでそのURLを開いて確認してみましょう。

もう一回アップロードしたい場合には、`dodo-cli upload`をもう一度実行するだけです。
簡単ですね。

# Next Steps
アップロードの基本はこれでおわりです。
より詳しい使い方を知りたい場合には以下のリンクを参考にしてください。

https://document.do.dodo-doc.com/yaml_specification

https://document.do.dodo-doc.com/ci