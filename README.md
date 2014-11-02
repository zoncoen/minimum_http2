# minimum http2

HTTP2 [最速実装](https://speakerdeck.com/syucream/2-zui-su-shi-zhuang-v3)

## example

最初に example/server.go を実行し、別のターミナルで example/client.go を実行する

- Server
    - Usage: `[DEBUG=1] go run example/server.go {PORT}`

```
$ DEBUG=1 go run example/server.go 5050
2014/11/03 08:09:14 Recv connection preface
2014/11/03 08:09:14 Send SettingsFrame UNSET 0
2014/11/03 08:09:14 Recv SettingsFrame UNSET 0
2014/11/03 08:09:14 Send SettingsFrame ACK 0
2014/11/03 08:09:14 Recv SettingsFrame ACK 0
2014/11/03 08:09:14 create new stream 1
2014/11/03 08:09:14 Recv HeadersFrame END_HEADERS 1
2014/11/03 08:09:14 Send HeadersFrame END_HEADERS 1
2014/11/03 08:09:14 Send DataFrame END_STREAM 1
2014/11/03 08:09:14 Recv GoAwayFrame UNSET 0
2014/11/03 08:09:14 Got EOF
```

- Client
    - Usage: `[DEBUG=1] go run example/client.go {ADDR} {PORT}`


```
$ DEBUG=1 go run example/client.go 127.0.0.1 5050
2014/11/03 08:09:14 Send connection preface
2014/11/03 08:09:14 Send SettingsFrame UNSET 0
2014/11/03 08:09:14 Recv SettingsFrame UNSET 0
2014/11/03 08:09:14 Send SettingsFrame ACK 0
2014/11/03 08:09:14 Recv SettingsFrame ACK 0
2014/11/03 08:09:14 Send HeadersFrame UNSET 1
2014/11/03 08:09:14 Recv HeadersFrame END_HEADERS 1
2014/11/03 08:09:14 Recv DataFrame END_STREAM 1
Hello HTTP/2!
```
