# validator


```go
package main

import (
	"fmt"
	"github.com/wantnotshould/sol/validator"
)

func main() {
	// 设置语言为中文
	validator.SetLanguage(validator.ZH)
	fmt.Println(validator.GetMessage("required", nil)) // 此字段是必填的
	fmt.Println(validator.GetMessage("min", 18)) // 此字段必须至少为 18

	validator.SetLanguage(validator.EN)
	fmt.Println(validator.GetMessage("required", nil)) // This field is required
	fmt.Println(validator.GetMessage("min", 18)) // This field must be at least 18

	fmt.Println(validator.GetMessage("nonexistent", nil)) // Invalid validation rule
}
```