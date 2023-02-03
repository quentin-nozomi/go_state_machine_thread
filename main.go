package main

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type State interface {
	goNext(*Context) string
	goBack(*Context) string
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

func (c *Context) GoNext() string {
	c.stateAccess.Lock()
	return c.state.goNext(c)
}

func (c *Context) GoBack() string {
	c.stateAccess.Lock()
	return c.state.goBack(c)
}

type First struct{}

func (First) goNext(c *Context) string {
	c.state = Transition{"First -> Second"}
	c.stateAccess.Unlock()
	defer c.setState(Second{})
	time.Sleep(1 * time.Millisecond) // simulating a slow transition
	return "First -> Second\n"
}

func (First) goBack(c *Context) string {
	c.stateAccess.Unlock()
	return "Cannot go back, already at the beginning!\n"
}

func (First) name() string {
	return "First"
}

type Second struct{}

func (Second) goNext(c *Context) string {
	c.state = Transition{"Second -> Third"}
	c.stateAccess.Unlock()
	defer c.setState(Third{})
	time.Sleep(3 * time.Millisecond) // simulating a very slow transition
	return "Second -> Third\n"
}

func (Second) goBack(c *Context) string {
	c.state = Transition{"Second -> First"}
	c.stateAccess.Unlock()
	defer c.setState(First{})
	return "Second -> First\n"
}

func (Second) name() string {
	return "Second"
}

type Third struct{}

func (Third) goNext(c *Context) string {
	c.stateAccess.Unlock()
	return "Cannot go forward, already at the end!\n"
}

func (Third) goBack(c *Context) string {
	c.state = Transition{"Third -> Second"}
	c.stateAccess.Unlock()
	defer c.setState(Second{})
	return "Third -> Second\n"
}

func (Third) name() string {
	return "Third"
}

type Transition struct {
	CurrentName string
}

func (Transition) goNext(c *Context) string {
	c.stateAccess.Unlock()
	return "Busy, cannot GoNext\n"
}

func (Transition) goBack(c *Context) string {
	c.stateAccess.Unlock()
	return "Busy, cannot GoBack\n"
}

func (t Transition) name() string {
	return t.CurrentName
}

// Testing concurrency, every run will be random:
func main() {
	stateMachine := Context{
		First{}, sync.RWMutex{},
	}

	var sb strings.Builder

	waitGroup := sync.WaitGroup{}

	waitGroup.Add(1)
	// Ask status constantly
	go func() {
		defer waitGroup.Done()
		for i := 0; i < 50; i++ {
			sb.WriteString(fmt.Sprintf("> Status? [%s]\n", stateMachine.Status()))
			time.Sleep(1 * time.Millisecond)
		}
	}()

	waitGroup.Add(1)
	// Ask to go next every 1 ms
	go func() {
		defer waitGroup.Done()
		for i := 0; i < 20; i++ {
			time.Sleep(1 * time.Millisecond)
			sb.WriteString(fmt.Sprintln("Please GoNext"))
			sb.WriteString(stateMachine.GoNext())
		}
	}()

	waitGroup.Add(1)
	// Ask to go back every 10 ms
	go func() {
		defer waitGroup.Done()
		for i := 0; i < 20; i++ {
			time.Sleep(10 * time.Millisecond)
			sb.WriteString(fmt.Sprintln("Please GoBack (10ms)"))
			sb.WriteString(stateMachine.GoBack())
		}
	}()

	waitGroup.Add(1)
	// Ask to go back every 25 ms
	go func() {
		defer waitGroup.Done()
		for i := 0; i < 20; i++ {
			time.Sleep(25 * time.Millisecond)
			sb.WriteString(fmt.Sprintln("Please GoBack (25ms)"))
			sb.WriteString(stateMachine.GoBack())
		}
	}()

	waitGroup.Wait()

	fmt.Println(sb.String())
}
