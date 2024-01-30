
## How to write `.dodo.yaml`
dodoは`.dodo.yaml`ファイルが配置されているディレクトリをドキュメント用のディレクトリとして認識します。


```yaml
pages:
  - path: "./index.md"
  - path: "./index.md"
```

pagesを指定した場合には, pagesに列挙した順番に従ってレイアウトが構成されます。
pagesを指定した上で列挙されなかったファイルはレイアウトに追加されません。
同じファイルを複数指定した場合にはエラーが発生します。


## Client Logics
Call /upload_archive
This endpoint returns the id for uploaded archive