# Golang Study

#### New keywords and usages in Golang explained by examples:
##### Defer: A defer statement defers the execution of a function until the surrounding function returns.
  ```
  package main

  import "fmt"

  func main() {
    defer fmt.Println("world")
    fmt.Println("Hello")
  }
  // Executing the code above will produce
  // Hello
  // world
  ```
##### Type define and initialize, memory allocation: 
  ```
  package main

  import "fmt"

  type A struct {
    ID string
  }

  func main() {
    // Declare a new object with type A
    a := A{ID:"ABC"}
    fmt.Println(a.ID)
    
    // Declare a new pointer pointed to address of a new empty A object
    aPointer := new(A)
    fmt.Println(aPointer)
    aPointer.ID = "CDF"
    fmt.Println(aPointer)
    fmt.Println(aPointer.ID)
  }
  // Executing the code above will produce
  // ABC
  // &{}
  // &{CDF}
  // CDF
  ```
##### Receivers of methods: if a method has a receiver, then it cannot be called directly.
  ```
  package main

  import (
    "fmt"
    "math"
  )

  type Vertex struct {
    X, Y float64
  }

  func (v Vertex) Abs() float64 {
    return math.Sqrt(v.X*v.X + v.Y*v.Y)
  }

  func (v *Vertex) Scale(f float64) {
    v.X = v.X * f
    v.Y = v.Y * f
  }

  func main() {
    v := Vertex{3, 4}
    fmt.Println(v.Abs())
    v.Scale(10)
    fmt.Println(v.Abs())
  }
  // Executing the code above will produce
  // 5
  // 50
  ```
##### Interface: a special method to support polymorphism
  ```
  package main

  import "fmt"
  
  type I interface {
    M()
  }
  
  type T1 struct{}
  
  func (T1) M() { fmt.Println("T1.M") }
  
  type T2 struct{}
  
  func (T2) M() { fmt.Println("T2.M") }
  
  func f(i I) { i.M() }
  
  func main() {
      f(T1{}) // "T1.M"
      f(T2{}) // "T2.M"
  }
  // Executing the code above will produce
  // T1.M
  // T2.M
  ```
##### Plugin: reference from [here](https://golang.org/pkg/plugin/)
* We can run command ```go build -buildmode=plugin FILENAME.go``` to create a plugin file named ```FILENAME.so```. The plugin file ```FILENAME.so``` is derived from ```FILENAME.go```, which all functions and variables in this file are now built as exported functions and variables.
* If another file needs to use those exported functions or variables from the plugin file ```FILENAME.so```, the file should first ```import "plugin"```, then use methods from the imported library to do that.
