# sol

[![Go Reference](https://pkg.go.dev/badge/github.com/wantnotshould/sol.svg)](https://pkg.go.dev/github.com/wantnotshould/sol)
[![Go version](https://img.shields.io/github/go-mod/go-version/wantnotshould/sol)](https://github.com/wantnotshould/sol)
[![GitHub license](https://img.shields.io/github/license/wantnotshould/sol)](https://github.com/wantnotshould/sol/blob/main/LICENSE)
[![GitHub stars](https://img.shields.io/github/stars/wantnotshould/sol?style=social)](https://github.com/wantnotshould/sol/stargazers)

**A go web framework.** 🌌

## Install

```bash
go get -u github.com/wantnotshould/sol
```

## Quick Start

```go
package main

import (
	"fmt"

	"github.com/wantnotshould/sol"
)

func main() {
	sl := sol.New()

	sl.GET("/", func(c *sol.Context) {
		fmt.Fprintln(c.Writer, "Hello, world!")
	})

	sl.Run()
}
```


## License

This project is under the MIT License. See the [LICENSE](LICENSE) file for the full license text.