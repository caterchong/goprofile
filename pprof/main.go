package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"
)

var (
	vec []string
)

func fib(n int) int {
	str := "hello"
	vec = append(vec, str)
	if n <= 1 {
		return n
	}
	return fib(n-1) + fib(n-2) + len(str)
}

func main() {
	//start block profile
	runtime.SetBlockProfileRate(1)

	runtime.SetMutexProfileFraction(1)

	mutexFile, err := os.Create("mutex.prof")
	if err != nil {
		panic(err)
	}
	defer mutexFile.Close()

	blockFile, err := os.Create("block.prof")
	if err != nil {
		panic(err)
	}
	defer blockFile.Close()

	// 创建CPU profile文件
	f, err := os.Create("cpu.prof")
	if err != nil {
		fmt.Println("Could not create CPU profile:", err)
		return
	}
	defer f.Close()

	// 创建 allocs profile 文件
	allocsFile, err := os.Create("allocs.prof")
	if err != nil {
		panic(err)
	}
	defer allocsFile.Close()

	// 创建 goroutine profile 文件
	routineFile, err := os.Create("goroutine.prof")
	if err != nil {
		panic(err)
	}
	defer routineFile.Close()

	// 开始CPU profile
	if err := pprof.StartCPUProfile(f); err != nil {
		fmt.Println("Could not start CPU profile:", err)
		return
	}

	fib(33) // 计算斐波那契数列第40个数

	defer pprof.StopCPUProfile()

	// 模拟一些工作负载
	var wg sync.WaitGroup
	workload(&wg)

	// 把当前存活的goroutine信息记录，根据堆栈做聚合
	pprof.Lookup("goroutine").WriteTo(routineFile, 0)

	wg.Wait()

	// 创建内存 profile文件
	memFile, err := os.Create("mem.prof")
	if err != nil {
		fmt.Println("Could not create memory profile:", err)
		return
	}
	defer memFile.Close()

	// 执行内存 profile
	vec = nil
	runtime.GC() // 获取内存分配情况之前触发GC

	// runtime.MemProfileRate, 通过这个控制采样，获得堆栈，这个是必须打开的，没有关闭的选项, 默认512K
	fmt.Printf("MemProfileRate: %d\n", runtime.MemProfileRate)
	if err := pprof.WriteHeapProfile(memFile); err != nil {
		fmt.Println("Could not write memory profile:", err)
		return
	}

	//alloc的采样间隔也是runtime.MemProfileRate
	pprof.Lookup("allocs").WriteTo(allocsFile, 0)

	// 注意delay和contentions之间的区别， delay是延迟的时间， conteionts是争用的次数
	pprof.Lookup("block").WriteTo(blockFile, 0)

	// mutex是block的子集
	pprof.Lookup("mutex").WriteTo(mutexFile, 0)
}

// 模拟工作负载函数
func workload(wg *sync.WaitGroup) {
	var mu sync.Mutex
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			mu.Lock()
			defer wg.Done()
			fmt.Printf("Worker %d is doing some work...\n", n)
			mu.Unlock()

			time.Sleep(1 * time.Second)
		}(i)
	}
}
