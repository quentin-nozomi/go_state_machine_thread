package main

import (
	"fmt"
	"sync"
	"time"
)

type State interface {
	GoNext(*Context)
	GoBack(*Context)
	Name() string
}

type Context struct {
	state       State
	stateAccess sync.RWMutex
}

func (c *Context) SetState(state State) {
	c.stateAccess.Lock()
	defer c.stateAccess.Unlock()
	c.state = state
}

func (c *Context) Status() string {
	c.stateAccess.RLock()
	defer c.stateAccess.RUnlock()
	return c.state.Name()
}

type First struct{}

func (First) GoNext(c *Context) {
	c.SetState(Transition{})
	defer c.SetState(Second{})
	fmt.Println("First -> Second")
	time.Sleep(1 * time.Millisecond) // simulating a slow transition
}

func (First) GoBack(*Context) {
	fmt.Println("Cannot go back, already at the beginning!")
}

func (First) Name() string {
	return "First"
}

type Second struct{}

func (Second) GoNext(c *Context) {
	c.SetState(Transition{})
	defer c.SetState(Third{})
	fmt.Println("Second -> Third")
	time.Sleep(3 * time.Millisecond) // simulating a very slow transition
}

func (Second) GoBack(c *Context) {
	c.SetState(Transition{})
	defer c.SetState(First{})
	fmt.Println("Second -> First")
}

func (Second) Name() string {
	return "Second"
}

type Third struct{}

func (Third) GoNext(*Context) {
	fmt.Println("Cannot go forward, already at the end!")
}

func (Third) GoBack(c *Context) {
	c.SetState(Transition{})
	defer c.SetState(Second{})
	fmt.Println("Third -> Second")
}

func (Third) Name() string {
	return "Third"
}

type Transition struct{}

func (Transition) GoNext(*Context) {
	fmt.Println("Busy, cannot GoNext")
}

func (Transition) GoBack(*Context) {
	fmt.Println("Busy, cannot GoNext")
}

func (Transition) Name() string {
	return "Transition"
}

// Testing concurrency, every run will be random
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
	// Ask to go next every 5 ms
	go func() {
		defer waitGroup.Done()
		for i := 0; i < 20; i++ {
			fmt.Println("Please GoNext")
			stateMachine.state.GoNext(&stateMachine)
			time.Sleep(5 * time.Millisecond)
		}
	}()

	waitGroup.Add(1)
	// Ask to go back every 10 ms
	go func() {
		defer waitGroup.Done()
		for i := 0; i < 20; i++ {
			fmt.Println("Please GoBack (10ms)")
			stateMachine.state.GoBack(&stateMachine)
			time.Sleep(10 * time.Millisecond)
		}
	}()

	waitGroup.Add(1)
	// Ask to go back every 25 ms
	go func() {
		defer waitGroup.Done()
		for i := 0; i < 20; i++ {
			fmt.Println("Please GoBack (25ms)")
			stateMachine.state.GoBack(&stateMachine)
			time.Sleep(25 * time.Millisecond)
		}
	}()

	waitGroup.Wait()
}
