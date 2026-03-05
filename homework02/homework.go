package main
import (
	"fmt"
	"time"
	"sync"
	"math"
	"sync/atomic"
	// "strconv"
	// "strings"
	// "sort"
)

//指针1-题目 ：编写一个Go程序，定义一个函数，该函数接收一个整数指针作为参数，在函数内部将该指针指向的值增加10，然后在主函数中调用该函数并输出修改后的值。
//考察点 ：指针的使用、值传递与引用传递的区别。
func AddTen(num *int) {
	*num += 10
}
//指针2
//题目 ：实现一个函数，接收一个整数切片的指针，将切片中的每个元素乘以2。
//考察点 ：指针运算、切片操作。
func MultiplyByTwo(slice *[]int) {
	for i := 0; i < len(*slice); i++ {
		(*slice)[i] *= 2
	}
}

//goroutine1
//题目 ：编写一个程序，使用 go 关键字启动两个协程，一个协程打印从1到10的奇数，另一个协程打印从2到10的偶数。
//考察点 ： go 关键字的使用、协程的并发执行。
func PrintOddEven() {
	go func() {
		for i := 1; i <= 10; i += 2 {
			fmt.Println("奇数：", i)
			// time.Sleep(1000 * time.Millisecond)
		}
	}()

	go func() {
		for i := 2; i <= 10; i += 2 {
			fmt.Println("偶数：", i)
			// time.Sleep(1000 * time.Millisecond)
		}
	}()

	time.Sleep(time.Second)
}
//goroutine2
//题目 ：设计一个任务调度器，接收一组任务（可以用函数表示），并使用协程并发执行这些任务，同时统计每个任务的执行时间。
//考察点 ：协程原理、并发任务调度。

type Task struct {//任务结构体
	Name string
	Func func()
	Duration time.Duration
}
type TaskScheduler struct { //任务调度器结构体
	tasks []*Task
	wg sync.WaitGroup
	mu sync.Mutex
}
//添加任务
func (ts *TaskScheduler) AddTask(name string, taskFunc func()) {
	task := &Task{
		Name: name,
		Func: taskFunc,
	}
	ts.tasks = append(ts.tasks, task)
}
//运行任务
func (ts *TaskScheduler) Run() {
	for _, task := range ts.tasks {
		ts.wg.Add(1)
		go func(task *Task) {
			defer ts.wg.Done()
			startTime := time.Now()
			fmt.Println("Starting task:", task.Name)
			// time.Sleep(task.Duration)
			task.Func()
			duration := time.Since(startTime).Milliseconds()
			task.Duration = time.Duration(duration)*time.Millisecond

			ts.mu.Lock()
			fmt.Printf("Task %s executed in %v\n", task.Name, task.Duration)
			ts.mu.Unlock()

			fmt.Println("Finished task:", task.Name)
		}(task)
	}
	ts.wg.Wait()
	fmt.Println("All tasks completed")
}
//打印所有任务的统计执行信息
func (ts *TaskScheduler) PrintStats() {
	// ts.mu.Lock()
	// defer ts.mu.Unlock()
	fmt.Println("Task Durations:")
	for _, task := range ts.tasks {
		fmt.Printf("%s: %v\n", task.Name, task.Duration)
	}
	// fmt.Println("Total Duration:", ts.wg.Wait())
	// ts.wg.Wait()
	// fmt.Println("All tasks completed")
}

//面向对象1
//题目 ：定义一个 Shape 接口，包含 Area() 和 Perimeter() 两个方法。然后创建 Rectangle 和 Circle 结构体，实现 Shape 接口。在主函数中，创建这两个结构体的实例，并调用它们的 Area() 和 Perimeter() 方法。
//考察点 ：接口的定义与实现、面向对象编程风格。
type Shape interface {
	Area() float64
	Perimeter() float64
}
type Rectangle struct {
	width  float64
	height float64
}
func (r Rectangle) Area() float64 {
	return r.width * r.height
}
func (r Rectangle) Perimeter() float64 {
	return 2 * (r.width + r.height)
}
type Circle struct {
	radius float64
}
func (c Circle) Area() float64 {
	return math.Pi * c.radius * c.radius
}
func (c Circle) Perimeter() float64 {
	return 2 * math.Pi * c.radius
}




//面向对象2
//题目 ：使用组合的方式创建一个 Person 结构体，包含 Name 和 Age 字段，再创建一个 Employee 结构体，组合 Person 结构体并添加 EmployeeID 字段。为 Employee 结构体实现一个 PrintInfo() 方法，输出员工的信息。
//考察点 ：组合的使用、方法接收者。
type Person struct {
	Name string
	Age  int
}
type Employee struct {
	Person
	EmployeeID int
}
func (e Employee) PrintInfo() {
	fmt.Printf("Name: %s, Age: %d, EmployeeID: %d\n", e.Name, e.Age, e.EmployeeID)
}



//channel1
//题目 ：编写一个程序，使用通道实现两个协程之间的通信。一个协程生成从1到10的整数，并将这些整数发送到通道中，另一个协程从通道中接收这些整数并打印出来。
//考察点 ：通道的基本使用、协程间通信。
func channelCommunity(){
	ch := make(chan int)
	go func() {
		defer close(ch)
		for i := 1; i <= 10; i++ {
			fmt.Println("Sent:", i)
			ch <- i
		}
	}()
	go func() {
		for num := range ch {
			fmt.Println("Received:", num)
		}
	}()

	time.Sleep(time.Second)
}

//channel2
//题目 ：实现一个带有缓冲的通道，生产者协程向通道中发送100个整数，消费者协程从通道中接收这些整数并打印。
//考察点 ：通道的缓冲机制。
func channelBufferCommunity() {
	ch := make(chan int, 10)
	// defer close(ch)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer close(ch)
		defer wg.Done()
		for i := 1; i <= 100; i++ {
			fmt.Println("Sent:", i)
			ch <- i
			// time.Sleep(1000 * time.Millisecond)
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for v := range ch {
			fmt.Println("Received1:",v)
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for v := range ch {
			fmt.Println("Received2:",v)
		}
	}()
	// time.Sleep(300000 * time.Millisecond)
	wg.Wait()
	fmt.Println("Done")
}


//锁机制1
//题目 ：编写一个程序，使用 sync.Mutex 来保护一个共享的计数器。启动10个协程，每个协程对计数器进行1000次递增操作，最后输出计数器的值。
//考察点 ： sync.Mutex 的使用、并发数据安全。
type CounterSafe struct {
	mu sync.Mutex
	count int
}
func (c *CounterSafe) Inc() {
	c.mu.Lock()
	c.count++
	c.mu.Unlock()
}
func (c *CounterSafe) Get() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}

//锁机制2
//题目 ：使用原子操作（ sync/atomic 包）实现一个无锁的计数器。启动10个协程，每个协程对计数器进行1000次递增操作，最后输出计数器的值。
//考察点 ：原子操作、并发数据安全。
func AutoAdd(){
	var count int32
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < 1000; j++ {
				atomic.AddInt32(&count, 1)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	fmt.Println("总数：",count)
}
func main() {
	/*num := 5
	AddTen(&num)
	fmt.Println(num)*/

	/*slice := []int{1, 2, 3, 4, 5}
	MultiplyByTwo(&slice)
	fmt.Println(slice)*/
	// PrintOddEven()


	/*scheduler := &TaskScheduler{}
	//创建任务
	scheduler.AddTask("Task 1", func() {
		fmt.Println("Task 1")
		time.Sleep(2 * time.Second)
	})
	scheduler.AddTask("Task 2", func() {
		fmt.Println("Task 2")
		time.Sleep(1 * time.Second)
	})
	scheduler.AddTask("Task 3", func() {
		fmt.Println("Task 3")
		time.Sleep(3 * time.Second)
	})
	//运行任务
	scheduler.Run()
	scheduler.PrintStats()*/

	/*rect := Rectangle{height: 5, width: 10}
	Area := rect.Area()
	fmt.Println("Rectangle Area:", Area)
	Perimeter := rect.Perimeter()
	fmt.Println("Rectangle Perimeter:", Perimeter)
//
	circle := Circle{radius: 5}
	AreaC := circle.Area()
	fmt.Println("Circle Area:", AreaC)
	PerimeterC := circle.Perimeter()
	fmt.Println("Circle Perimeter:", PerimeterC)*/

	/*person := Person{Age: 20, Name: "John"}
	employee := Employee{EmployeeID: 1, Person: person}
	employee.PrintInfo()*/

	// channelCommunity()
	// channelBufferCommunity()

	/*counterSafe := &CounterSafe{}
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				counterSafe.Inc()
			}
		}()
	}
	wg.Wait()
	fmt.Println("Counter:", counterSafe.Get())*/
	AutoAdd()
}


