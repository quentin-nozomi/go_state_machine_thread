package main

import (
	"fmt"
	"sync"
	"time"
)

type State interface {
	goNext(*Context)
	goBack(*Context)
	name() string
}

type Context struct {
	state       State
	stateAccess sync.RWMutex
}

func (c *Context) setState(state State) {
	c.stateAccess.Lock()
	defer c.stateAccess.Unlock()
	c.state = state
}

func (c *Context) Status() string {
	c.stateAccess.RLock()
	defer c.stateAccess.RUnlock()
	return c.state.name()
}

func (c *Context) GoNext() {
	c.stateAccess.Lock()
	c.state.goNext(c)
}

func (c *Context) GoBack() {
	c.stateAccess.Lock()
	c.state.goBack(c)
}

type First struct{}

func (First) goNext(c *Context) {
	c.state = Transition{}
	c.stateAccess.Unlock()
	defer c.setState(Second{})
	fmt.Println("First -> Second")
	time.Sleep(1 * time.Millisecond) // simulating a slow transition
}

func (First) goBack(c *Context) {
	c.stateAccess.Unlock()
	fmt.Println("Cannot go back, already at the beginning!")
}

func (First) name() string {
	return "First"
}

type Second struct{}

func (Second) goNext(c *Context) {
	c.state = Transition{}
	c.stateAccess.Unlock()
	defer c.setState(Third{})
	fmt.Println("Second -> Third")
	time.Sleep(3 * time.Millisecond) // simulating a very slow transition
}

func (Second) goBack(c *Context) {
	c.state = Transition{}
	c.stateAccess.Unlock()
	defer c.setState(First{})
	fmt.Println("Second -> First")
}

func (Second) name() string {
	return "Second"
}

type Third struct{}

func (Third) goNext(c *Context) {
	c.stateAccess.Unlock()
	fmt.Println("Cannot go forward, already at the end!")
}

func (Third) goBack(c *Context) {
	c.state = Transition{}
	c.stateAccess.Unlock()
	defer c.setState(Second{})
	fmt.Println("Third -> Second")
}

func (Third) name() string {
	return "Third"
}

type Transition struct{}

func (Transition) goNext(c *Context) {
	c.stateAccess.Unlock()
	fmt.Println("Busy, cannot GoNext")
}

func (Transition) goBack(c *Context) {
	c.stateAccess.Unlock()
	fmt.Println("Busy, cannot GoBack")
}

func (Transition) name() string {
	return "Transition"
}

// Testing concurrency, every run will be random:
func main() {
	stateMachine := Context{
		First{}, sync.RWMutex{},
	}

	waitGroup := sync.WaitGroup{}

	waitGroup.Add(1)
	// Ask status constantly
	go func() {
		defer waitGroup.Done()
		for i := 0; i < 50; i++ {
			fmt.Printf("> Status? [%s]\n", stateMachine.Status())
			time.Sleep(1 * time.Millisecond)
		}
	}()

	waitGroup.Add(1)
	// Ask to go next every 1 ms
	go func() {
		defer waitGroup.Done()
		for i := 0; i < 20; i++ {
			time.Sleep(1 * time.Millisecond)
			fmt.Println("Please GoNext")
			stateMachine.GoNext()
		}
	}()

	waitGroup.Add(1)
	// Ask to go back every 10 ms
	go func() {
		defer waitGroup.Done()
		for i := 0; i < 20; i++ {
			time.Sleep(10 * time.Millisecond)
			fmt.Println("Please GoBack (10ms)")
			stateMachine.GoBack()
		}
	}()

	waitGroup.Add(1)
	// Ask to go back every 25 ms
	go func() {
		defer waitGroup.Done()
		for i := 0; i < 20; i++ {
			time.Sleep(25 * time.Millisecond)
			fmt.Println("Please GoBack (25ms)")
			stateMachine.GoBack()
		}
	}()

	waitGroup.Wait()
}
