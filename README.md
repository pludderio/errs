# ERRS Package

This package was put together from different repositories as an attempt to merge and improve the coordination of the libraries. There are benefits to both libraries:

### [go-errors](https://github.com/go-errors/errors)

This package was amazing in that it allowed us Go developers to be able to trace an error easily anywhere in the code base. It's only problem was that it did not inter-operate well with `pkg/errors`. This is a huge drawback when dealing with libraries that do not use go-errors.

### [lbryio-errors](https://github.com/lbryio/lbry.go/tree/v2.7.1/extras/errors)

This package wanted the best of both worlds and attempted to build off of it. Allowing `Error` to be used from the go-errors package but also allowing it to be used more generically so it could play nice with `pkg/errors`


## Merging Both

The [ERRS](https://github.com/pludderio/errs) package is melding the two to actually get the best of both worlds. We can trace errors and we can have the inter operability with `pkg/errors`. This package does a few things specifically:

1. It removes the prior API of go-errors. This is important to prevent misuse of the intentions. It is now only used internally via [internal.go](internal.go). 

2. It creates a new API that can be used everywhere regardless of whether or not a library uses go-errors. This API comes from the work [grin](https://github.com/lyoshenka) did at [lbry](https://github.com/lbryio). 

[I](https://github.com/tiger5226) have used the latter package because of its flexibility for quite some time. It is time to merge it into a new creation. 


## Example

```go
package main

import (
	"fmt"
	"github.com/pludderio/errs"
)

var Crashed = errors.Base("oh dear")
  func main() {
     err := Crash()
     if err != nil {
          if errors.Is(err, Crashed) {
              fmt.Println(errors.FullTrace(err))
          } else {
              panic(err)
          }
      }
  }


func Crash() error {
	return errors.Err(Crashed)
}
```

### API

#### Err

This is the standard API for err wrapping. You just wrap every error in your application that you create or get from an external library. So that any any higher level point you can print the full stack trace. 

#### FullTrace

Use this to get the full stack of an error. When you are using `Err`, you can pull up this information any time you want whereever you want. This is really important for understand where the origination point of the error is. 

#### ErrSkip

I never really found a good use case for this, but I assume as some point I will. This allows you to remove/skip sections of the stack. So if you know where you create it, and you know nothing will go wrong up to a certain point you can skip those levels. Technically this should be useful when your stack goes really deep. 

#### Unwrap

This will unwrap everything to get you the original error. You don't really need to use this for comparison because the `Is` API does it automatically. However, it does not do it more than one level down. So you would need this for comparison deeper or with offset depths.

#### Is

For comparing two wrapped errors. 

#### Prefix

Allows you to prefix an error you already have to add more context as come back out of the stack. 

#### Trace

Returns the stack trace for the error as a `string`.

#### FullTrace

Returns the error type, message, and stack trace for added debug information. Traditionally I just always use this so it goes to the logs when an unexpected error occurs in production. It makes debugging the problem massively more easy than `Trace` alone.

#### Base

Create a stackless error. This is the origin point of errors within your own codebase. You would declare this into a var to be used. Keep in mind the origin point from a library would be that first error handling where you would wrap it with `Err`.

#### HasTrace

This just retuns whether or not the error even has a trace with it or now. 


### CONTRIBUTIONS

They are always welcome. I need to migrate the tests from go-errors repo still. Just create a pull request and I will get it merged in ASAP, and if its needed even sooner than that I can tag it for you as a branch. 

