# HTTP-download-server

## Run locally

## Go

### 1. Download project

```shell
git clone https://github.com/Lemon001017/HTTP-download-server.git
```

### 2. Install dependencies

```shell
cd server
go mod tidy
go mod download
```

### 3. Unit Test

```shell
go test ./...
```

### 4. Run

```shell
cd cmd
go run  .  -dsn file:dev.db
```

### Show api docs

```shell
http://localhost:8000/api/docs/
```

## Web

### 1. Install the vscode plugin

Click -> [Live Server](https://marketplace.visualstudio.com/items?itemName=ritwickdey.LiveServer) or Search `Live 
Server` in vscode

### 2. Modify settings.json (Required)

1. enter into Settings
2. Search `live server settings ignore files`
3. Click `Edit in settings.json`
4. Add the following:

```
"liveServer.settings.ignoreFiles": [
    "**/server/**/*",
    "**/*.go",
    "**/*.db",
]
```

5. Restart vscode

### 3. Run

```shell
cd ui
cd public/index.html
```

Right click the mouse to select -> `Open with Live Server`
