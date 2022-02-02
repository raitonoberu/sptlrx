## Building sptlrx

Make sure you have [Go 1.17+](https://go.dev/) installed.

### Clone the repository

```sh
git clone https://github.com/raitonoberu/sptlrx
cd sptlrx
```

### Fetch dependencies

```sh
go get
```

### Build it

```sh
go build -ldflags '-w -s'
```

### Run it

```sh
./sptlrx
```
