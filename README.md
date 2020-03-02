#Usage

##FixedExecutor
```go
exec := executor.NewFixedExecutor(5) // create 5 fixed worker
future := exec.Execute(func() interface{} {
  time.Sleep(1 * time.Second)
  return "this is my value"
})


if <-future.Done(); future.Error() != nil {
  fmt.Println(future.Value().(string)) //this is my value
}else {
  //handle error
}

exec.Release()
```
