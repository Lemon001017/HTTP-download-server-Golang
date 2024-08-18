# HTTP-download-server

## Run locally

## Go

### Download project

```shell
git clone https://github.com/Lemon001017/HTTP-download-server.git
cd server
```

### Install dependencies

```shell
go mod tidy
go mod download
```

### Unit Test

```shell
go test ./...
```

### Run

```shell
cd cmd
go run  .  -dsn file:dev.db
```

## Web

### Install the vscode plugin

Click -> [Live Server](https://marketplace.visualstudio.com/items?itemName=ritwickdey.LiveServer) or Search `Live Server` in vscode

### Run

```shell
cd ui
cd public/index.html
```

Right click the mouse to select -> `Open with Live Server`
