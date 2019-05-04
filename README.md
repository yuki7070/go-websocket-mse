
# WebSocket-MSE

*勉強用です。実用的なコードではありません。


websocketを使用してwebmコンテナの動画をバイナリで送信し、JS側でコンテナをパースしてMSEで再生

### Server

websocketのwriteの部分を回りくどい書き方していますが,io.Writerインターフェイスを満たすやつを渡すことで動的に書き込み対象を増やすことが出来ます。

*今回はwebmコンテナを書き込んでいるので初期化セグメントの部分がないと途中から書き込んでも再生できない


webmコンテナのパーサーをサーバサイドで実装してclusterごとに書き込むのが良さそう

### 参考
- https://qiita.com/tomoyukilabs/items/57ba8a982ab372611669
