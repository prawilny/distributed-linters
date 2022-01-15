```
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -installsuffix machine_spawner/main.go -o machine_spawner
```

```
cp spawner machine_spawner/
podman build machine_spawner/ -t spawner
```
