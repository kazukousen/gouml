Automatically generate PlantUML from Go Code.  

example (self-reference):  
![self-ref](self-ref.png)

Note that the interface of this library is still ALPHA level quality.  
Breaking changes will be introduced frequently.  

## Getting Started

```sh
$ go get -u github.com/kazukousen/gouml/cmd/gouml
```

Run `gouml init` (or `gouml i`) . This will parse `.go` files and generate the plantUML file.  

### Directory-base
If you requires point to base directory, you can use `-d` flag.  

```console
$ gouml i -d path/to/package/
```

### File-base
If you requires parse one `.go` file. you can use `-f` flag.

```console
$ gouml i -f path/to/package/foo.go -f path/to/package/bar.go
```

## License

Copyright (c) 2019-present [Kazuki Nitta](https://github.com/kazukousen)

Licensed under [MIT License](./LICENSE)
