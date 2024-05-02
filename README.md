
## Summary of uploading

* PageConfigをPageに変換する
* Pageに基づいて必要なファイルをarchiveする
* 最後にアップロードする

## Issues

なんとなくの課題感
* 基本的にyamlの記載はシンプルにしたい。
* なのでmatchとfilepathでの差は可能な限り無くしたい。= titleやdescriptionをyamlに記載したくない。

* matchにはexcludeとfilepathを記載した際の挙動を明確にしたい。
* pathを指定できるのはよくなさそう。pathnameという名前にしてスラッシュを禁止したい。


## 実装におけるメモ
* pathは親ディレクトリのパスと結合したものが最終的なパスになるが、結合処理はサーバー側で行う。
* pathに使える文字は[a-zA-Z0-9-_]に限定する。