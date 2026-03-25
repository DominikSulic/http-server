package main

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"sync"
	"time"
	"unicode"
)

// go mod tidy is nice, can be used instead of go get

/**
 * ctrl e -> file finder
 * ctrl r -> find next
 * ctrl d -> duplicate
 * ctrl k c -> comment a line
 * ctrl j -> shell
 * ctrl b -> close files
 * ctrl shift d -> debug panel
 * ctrl shift k -> delete line
 * f2 -> rename (doesnt seem to work via gopls?)
 */

/**
 * goroutines are lightweight threads managed by the Go runtime
 * go functionA(x,y) starts a new goroutine running functionA(x,y)
 * the evaluation of functionA, x and y happens in the current goroutine and the execution of functionA happens
 * in the new goroutine. They run in the same adrress space, access to shared memory has to be synced -> Sync package.
 *
 * Channels are a typed conduit through which you can send and receive values with <- (goroutines use them to pass values to each other through them)
 * ch <- a //send a to channel ch
 * v := <- ch // receive from ch and assign value to v. -> data flows in the direction of the arrow.
 * create channels before use:
 * ch := make(chan int)
 *
 * by default, sends and receives block untill the other side is ready -> this allows goroutines to sync without locks or condition variables.
 */
func main() {
	go printStuff("hi")
	printStuff("123")

	numbers := []int{7, 2, 8, -9, 4, 0}

	channel := make(chan int) // this int defines the type of variable you work with?
	go sum(numbers[:len(numbers)/2], channel)
	go sum(numbers[len(numbers)/2:], channel)

	x, y := <-channel, <-channel // they both receive from channel

	fmt.Println("Concurrent sum calculation:", x, y, x+y)

	/**
	 * you can make a buffered channel - '2' is the buffer length in the 'make' below
	 * Sends to a buffered channel block only when the buffer is full. Receives block when the buffer is empty.
	 */
	testChannel := make(chan int, 2)
	testChannel <- 1 // sends data
	testChannel <- 2
	// testChannel <- 3 this one throws an error since you put the buffer to 2.
	fmt.Println("Buffered channels: ")
	fmt.Println(<-testChannel) // 'receiving' data
	fmt.Println(<-testChannel)

	/**
	 * A sender can close a channel to indicate no more values will be sent. Receivers can test whether a channel has been closed:
	 * value, ok := <- ch
	 * ok is false if there are no more values to receive and the channel is closed. Only the sender should close a channel, never the receiver.
	 * Channels dont usually need closing -> only to tell the receiver there are no more values coming - i.e. to terminate a range loop.
	 */
	fmt.Println("Fibonacci:  ")
	fibonacciChannel := make(chan int, 10)
	// cap = capacity of the array
	go fibonacci(cap(fibonacciChannel), fibonacciChannel)
	for i := range fibonacciChannel {
		fmt.Println(i)
	}

	fmt.Println("Select and default: ")
	selectAndDefaultInGoRoutines()

	fmt.Println("Mutex: ")
	syncMutex()

	testingMoreStuff()
}

func printStuff(stuff string) {
	for i := 0; i < 5; i++ {

		time.Sleep(100 * time.Millisecond)
		fmt.Println(stuff)
	}
}

func sum(numbers []int, channel chan int) {
	sum := 0

	for _, number := range numbers {
		sum += number
	}

	channel <- sum
}

func fibonacci(n int, channel chan int) {
	x, y := 0, 1
	for i := 0; i < n; i++ {

		channel <- x
		x, y = y, x+y
	}
	close(channel)
}

func selectAndDefaultInGoRoutines() {
	startTime := time.Now()
	tick := time.Tick(100 * time.Millisecond)
	endTime := time.After(500 * time.Millisecond)
	elapsed := func() time.Duration {
		return time.Since(startTime).Round(time.Millisecond)
	}

	for {
		select {
		case <-tick:
			fmt.Printf("[%6s tick.\n", elapsed())

		case <-endTime:
			fmt.Printf("[%6s] timer ended!\n", elapsed())
			return

		default:
			fmt.Printf("[%6s        .\n", elapsed())
			time.Sleep(50 * time.Millisecond)
		}
	}
}

/**
 * If you want to make sure that only one goroutine can access a variable at a time to avoid conflicts, you can use mutex - mutual exclusion.
 * sync.Mutex provides 2 methods - Lock and Unlock. Surround a block of code in Lock and Unlock to be executed in mutex.
 * you can also use defer to ensure the mutex will be unlocked
 */
func syncMutex() {
	concurrentCounter := ConcurrentCounter{values: make(map[string]int)}

	for i := 0; i < 1000; i++ {
		go concurrentCounter.IncrementCounter("someKey")
	}

	time.Sleep(time.Second)
	fmt.Println(concurrentCounter.getCurrentValue("someKey"))
}

type ConcurrentCounter struct {
	mutex  sync.Mutex
	values map[string]int
}

/*
 * this is how you ensure concurrent processing will work well
 */
func (concurrentCounter *ConcurrentCounter) IncrementCounter(key string) {
	concurrentCounter.mutex.Lock()
	// Lock so only one goroutine at a time can access the map concurrentCounter.values
	concurrentCounter.values[key]++
	concurrentCounter.mutex.Unlock()
}

/*
 * this is how you ensure concurrent processing will work well
 */
func (concurrentCounter *ConcurrentCounter) getCurrentValue(key string) int {
	concurrentCounter.mutex.Lock()
	// Lock so only one goroutine at a time can access the map concurrentCounter.values
	defer concurrentCounter.mutex.Unlock()
	// defer ensures the mutex will be unlocked when this function returns - good practice to define after a lock
	// because it ensures you wont forget to unlock the mutex
	return concurrentCounter.values[key]
}

// ------------------------------------------------------------------------------------------------------------------------------------------------
// ------------------------------------------------------------------------------------------------------------------------------------------------
// ------------------------------------------------------------------------------------------------------------------------------------------------
// ------------------------------------------------------------------------------------------------------------------------------------------------
// ------------------------------------------------------------------------------------------------------------------------------------------------
// ------------------------------------------------------------------------------------------------------------------------------------------------
/**
 * https://antonz.org/go-concurrency/goroutines/
 */
func testingMoreStuff() {
	var waitGroup sync.WaitGroup // has a counter inside

	waitGroup.Add(2) // the counter is incremented by 2

	go func() {
		// ensures the goroutine decrements the counter by 1 before exiting, even if the func panics.
		// decoupled concurrency logic and business logic
		defer waitGroup.Done()
		saySomething(1, "something is being said!")
	}()

	go func() {
		defer waitGroup.Done()
		saySomething(2, "something else is being said!")
	}()

	/*
	 * goroutines are independent - func main wont wait for them. the main function is also a goroutine, but the other
	 * goroutines end when the main ends.
	 * Wait blocks the goroutine (main) untill the counter reaches 0 -> it waits for the 2 goroutines
	 */
	waitGroup.Wait()
	fmt.Println("done with normal sayings")

	/*
	 * WaitGroup.Go automagically increments the counter, runs a function in a goroutine and decrements the counter when its done.
	 */

	waitGroup.Go(func() {
		fmt.Println("More concurrent sayings")
	})

	waitGroup.Go(func() {
		fmt.Println("Even more concurrent sayings!")
	})

	waitGroup.Wait()
	fmt.Println("done with concurrent sayings")

	fmt.Printf("The above sentence contains: %d digits\n", countDigitsInWords("d0ne w1th c0ncurrent say1ngs"))

	messages := make(chan string)

	// the first goroutine (the one below) sends a message to the second (main) through the messages channel
	// when the sending goroutine writes a value to the channel, it blocks and waits for someone to receive that value, only then does it continue.
	go func() { messages <- "hi" }()

	message := <-messages
	fmt.Println(message)

	sentence := "Th1s sentence c0nta1ns 6 d1g1ts: %d\n"

	fmt.Printf(sentence, countDigitsInWordsUsingChannels(sentence))

	fmt.Printf("using a generator to count digits:\n")

	fmt.Printf("generated digits: %d\n", countDigitsInWordsUsingAGenerator(getRandomStringForGenerator))

	// Returning an output channel from a function and filling it within an internal goroutine is a common Go pattern
	fmt.Printf("generated digits using 2 goroutines: %d\n", countDigitsInWordsUsingGoroutines(getRandomStringForGenerator))
}

type pair struct {
	word  string
	count int
}

func countDigitsInWordsUsingGoroutines(getNextWord func() string) int {
	pending := sendWordsToBeCounted(getNextWord)
	counted := countWordsGoroutine(pending)

	counter := 0
	for {
		tmp := <-counted
		counter += tmp.count
		if tmp.word == "" {
			break
		}
	}

	return counter
}

func countWordsGoroutine(input chan string) chan pair {
	output := make(chan pair)

	go func() {
		for {
			word := <-input
			count := countDigitsInAWord(word)
			output <- pair{word, count}
			if word == "" {
				break
			}
		}
	}()

	return output
}

func sendWordsToBeCounted(getNextWord func() string) chan string {
	output := make(chan string)

	go func() {
		for {
			word := getNextWord()
			output <- word
			if word == "" {
				break
			}
		}
	}()

	return output
}

func getRandomStringForGenerator() string {
	if 1 == rand.Intn(5) {
		return ""
	}

	return "123"
}

func countDigitsInWordsUsingAGenerator(getNextWord func() string) int {
	outputChannel := make(chan pair)

	go func() {
		for {
			word := getNextWord()
			count := countDigitsInAWord(word)
			outputChannel <- pair{word, count}

			if word == "" {
				break
			}
		}
	}()

	counter := 0

	for {
		pairTmp := <-outputChannel
		if pairTmp.word == "" {
			break
		}
		counter += pairTmp.count
	}

	return counter
}

func countDigitsInWordsUsingChannels(sentence string) int {
	outputChannel := make(chan int)

	words := strings.Fields(sentence)

	go func() {
		for _, word := range words {
			outputChannel <- countDigitsInAWord(word)
		}
	}()

	counter := 0

	for range words {
		counter += <-outputChannel
	}

	return counter
}

func saySomething(id int, phrase string) {
	// splits the string around each whitespace group
	for _, word := range strings.Fields(phrase) {

		fmt.Printf("Worker #%d says: %s... \n", id, word)
		duration := time.Duration(rand.Intn(100)) * time.Millisecond
		time.Sleep(duration)
	}
}

func countDigitsInAWord(word string) int {
	count := 0

	for _, character := range word {
		if unicode.IsDigit(character) {
			count++
		}
	}
	return count
}

/**
 * Create a function countDigitsInWords that takes an input string, splits it into words, and counts the digits in each word using countDigits.
 * Be sure to do the counting for each word in a separate goroutine.
 * We haven't discussed how to modify shared data from different goroutines yet, so there is a ready-to-use variable called syncStats that you can
 * safely access from goroutines.
 */
func countDigitsInWords(sentence string) int {
	var waitGroup sync.WaitGroup
	syncStats := new(sync.Map)
	words := strings.Fields(sentence)

	for _, word := range words {
		waitGroup.Go(func() {
			syncStats.Store(word, countDigitsInAWord(word))
		})
	}

	waitGroup.Wait()

	counter := 0
	syncStats.Range(func(word, count any) bool {
		counter += count.(int)
		return true
	})

	return counter
}

// in the beginning the method used in main.go was:
func getLinesFromTheListener(connection io.ReadCloser) <-chan string {
	outputChannel := make(chan string, 1)

	go func() {
		defer connection.Close()
		defer close(outputChannel)

		lines := ""
		for {

			data := make([]byte, 8)

			numberOfLinesRead, err := connection.Read(data)
			if err != nil {
				break
			}

			data = data[:numberOfLinesRead]
			if i := bytes.IndexByte(data, '\n'); i != -1 {

				lines += string(data[:i])
				data = data[i+1:]
				outputChannel <- lines
				lines = ""
			}

			lines += string(data)
		}

		if len(lines) != 0 {
			outputChannel <- lines
		}
	}()

	return outputChannel
}
