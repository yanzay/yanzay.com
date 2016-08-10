+++
date = "2016-08-10T12:34:23+03:00"
draft = false
title = "Реализация аналога netcat на go"
tags = ["go", "network", "netcat"]
slug = "go_netcat"
+++

В этой статье мы рассмотрим основы работы с сетью в go и реализуем простейший аналог известной утилиты netcat.

Для начала определимся с необходимым функционалом. Дабы не усложнять, остановимся на таких фичах:

- только TCP
- сервер принимает соединения и выводит на stdout всё что ему присылает клиент
- клиент соединяется с сервером и просто посылает ему весь stdin
<!--more-->

## Сервер
Для управления опциями командной строки, будем использовать пакет flag из стандартной библиотеки. Итак, нам нужен флаг, указывающий что это сервер (`-l`), на каком хосте слушать (`-h`) и на каком порте (`-p`):

```go
var (
  listen = flag.Bool("l", false, "Listen")
  host   = flag.String("h", "localhost", "Host")
  // если оставить дефолтный порт 0, то сервер запустится на случайном порту
  port   = flag.Int("p", 0, "Port")
)
```

В функции main в первую очередь распарсим эти флаги и запустим сервер если передан флаг `-l`:

```go
func main() {
  flag.Parse()
  if *listen {
    startServer()
    return
  }
}
```

И тепер собственно реализуем сам сервер:

```go
func startServer() {
  // вычисляем адрес сервера по хосту и порту
  addr := fmt.Sprintf("%s:%d", *host, *port)
  // запускаем TCP сервер
  listener, err := net.Listen("tcp", addr)

  if err != nil {
    // если не удалось запустить сервер,
    // то тут мы ничего не можем поделать, паникуем
    panic(err)
  }

  log.Printf("Listening for connections on %s", listener.Addr().String())

  for {
    conn, err := listener.Accept()
    if err != nil {
      log.Printf("Error accepting connection from client: %s", err)
    } else {
      go processClient(conn)
    }
  }
}
```

Обратите внимание на цикл for и что происходит внутри. Мы в бесконечном цикле принимаем соединения от клиентов, но обрабатываем эти соединения в отдельных горутинах `go processClient(conn)`, то есть после успешного подключения клиента, сервер возвращается на `listener.Accept()` и готов принимать следующего.

Дело за малым, реализовать основную логику нашего сервера, а именно обработку соединения клиента:

```go
func processClient(conn net.Conn) {
  _, err := io.Copy(os.Stdout, conn)
  if err != nil {
    fmt.Println(err)
  }
  conn.Close()
}
```

Вот и весь фокус! Всё что нам нужно сделать -- это перенаправить поток данных от клиента в наш `os.Stdout`, что мы успешно реализовали с помощью `io.Copy`. Эта функция будет читать из `conn` до тех пор, пока не встретит End Of File (EOF), либо пока не произойдёт какая-то ошибка. Ну и, конечно же, не забываем закрыть соединение после себя.

Наш сервер готов, проверим его с помощью штатного netcat, чтоб убедиться что всё работает как надо:

```bash
$ # запускаем сервер
$ go run main.go -l -p 1408

$ # в отдельном терминале подключимся штатным netcat к нашему серверу
$ nc localhost 1408
```

![Netcat Server](/images/netcat_server.png)

## Клиент

Настало время заняться клиентом, добавим в main:

```go
func main() {
  flag.Parse()
  if *listen {
    startServer()
    return
  }
  if len(flag.Args()) < 2 {
    fmt.Println("Hostname and port required")
    return
  }
  serverHost := flag.Arg(0)
  serverPort := flag.Arg(1)
  startClient(fmt.Sprintf("%s:%s", serverHost, serverPort))
}
```

Здесь мы проверяем, переданы ли нужные нам аргументы и запускаем функцию `startClient` с адресом сервера.

Теперь реализуем логику клиента:

```go
func startClient(addr string) {
  conn, err := net.Dial("tcp", addr)
  if err != nil {
    fmt.Printf("Can't connect to server: %s\n", err)
    return
  }
  _, err = io.Copy(conn, os.Stdin)
  if err != nil {
    fmt.Printf("Connection error: %s\n", err)
  }
}
```

Здесь мы для начала дозваниваемся до TCP сервера с помощью `net.Dial`, а далее используем уже знакомый нам `io.Copy` для перенаправления потока из stdin в conn. Готово!

Проверяем всё вместе:

```bash
$ # запускаем сервер
$ go run main.go -l -p 1408

$ # в отдельном терминале запускаем клиент
$ go run main.go localhost 1408
```

![Go Netcat](/images/netcat_client.png)

Вот так, простейший аналог netcat готов, полный код:

```go
package main

import (
  "flag"
  "fmt"
  "io"
  "log"
  "net"
  "os"
)

var (
  listen = flag.Bool("l", false, "Listen")
  host   = flag.String("h", "localhost", "Host")
  port   = flag.Int("p", 0, "Port")
)

func main() {
  flag.Parse()
  if *listen {
    startServer()
    return
  }
  if len(flag.Args()) < 2 {
    fmt.Println("Hostname and port required")
    return
  }
  serverHost := flag.Arg(0)
  serverPort := flag.Arg(1)
  startClient(fmt.Sprintf("%s:%s", serverHost, serverPort))
}

func startServer() {
  addr := fmt.Sprintf("%s:%d", *host, *port)
  listener, err := net.Listen("tcp", addr)

  if err != nil {
    panic(err)
  }

  log.Printf("Listening for connections on %s", listener.Addr().String())

  for {
    conn, err := listener.Accept()
    if err != nil {
      log.Printf("Error accepting connection from client: %s", err)
    } else {
      go processClient(conn)
    }
  }
}

func processClient(conn net.Conn) {
  _, err := io.Copy(os.Stdout, conn)
  if err != nil {
    fmt.Println(err)
  }
  conn.Close()
}

func startClient(addr string) {
  conn, err := net.Dial("tcp", addr)
  if err != nil {
    fmt.Printf("Can't connect to server: %s\n", err)
    return
  }
  _, err = io.Copy(conn, os.Stdin)
  if err != nil {
    fmt.Printf("Connection error: %s\n", err)
  }
}
```

В [следующей статье]({{< relref "go_netcat_2.md" >}}) мы реализуем удалённый запуск программ через наш netcat и получим удалённый shell.
