## Getting Started

```sh
$ go get -u github.com/kazukousen/gouml/cmd/gouml
```

Run `gouml init` . This will parse `.go` files and generate the plantUML file.  
(Option) If you requires point to base directory, you can use `-d` flag.  

```sh
$ gouml init -d example/model/
```

## Example

from `example/model/` directory.  

![sample image](/example.png)

and self-reference :tada:  

![self-ref image](self-ref.png)
