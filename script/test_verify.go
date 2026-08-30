package main

import (
	"fmt"
	"go-zero-gorm-code/internal/util"
)

func main() {
	hash := "DpwH0D8PW6gHktJ2oTgE/8SE3op6atizUDGIXiDeqAugAlfLe7goyhA7kpONjOmA"
	fmt.Printf("Hash: %s\n", hash)
	fmt.Printf("Hash length: %d\n", len(hash))

	ok := util.VerifyPassword(hash, "123456")
	fmt.Printf("Verify '123456': %v\n", ok)

	// 测试生成新密码
	newHash, err := util.HashPassword("123456")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("\nNew Hash: %s\n", newHash)
	fmt.Printf("New Hash length: %d\n", len(newHash))

	ok2 := util.VerifyPassword(newHash, "123456")
	fmt.Printf("Verify with new hash: %v\n", ok2)
}
