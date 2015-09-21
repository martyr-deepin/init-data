package main

import (
	"fmt"
	"math/rand"
	"time"
)

var cow1 = `
(Ͼ˳Ͽ)..!!! 很快就没有问题了, 加油干呀小伙子
`

var cow2 = `
我(#‵′)靠 竟然没有错误
        \   ^__^
         \  (oo)\_______
            (__)\       )\/\
                ||----w |
                ||     ||
加上 -upload=true 开始上传吧
`

var cows = []string{
	`
还有%d个问题, 今天不解决就
          ▄︻̷̿┻̿═━一
`,
	`
还有%d个问题, 今天不解决就
      ︻デ┳═ー*----*	
`,
	`
还有%d个问题, 今天不解决就
                   ()==[:::::::::::::>
`,
}

func ShowCow(n int) {
	rand.Seed(time.Now().Unix())
	if n == 0 {
		fmt.Println(cow2)
	} else if n < 10 {
		fmt.Println(cow1)
	} else {
		fmt.Printf(cows[rand.Int31n(int32(len(cows)))], n)
	}
}
