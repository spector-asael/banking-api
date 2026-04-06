package helpers

import (
	"fmt"
)

// Accept a function and run it in the background also recover from any panic
func (a *HelperDependencies) Background(fn func()) {
	a.WG.Add(1) // Use a wait group to ensure all goroutines finish before we exit
	go func() {
		defer a.WG.Done() // signal goroutine is done
		defer func() {
			err := recover()
			if err != nil {
				a.Logger.Error(fmt.Sprintf("%v", err))
			}
		}()
		fn() // Run the actual function
	}()
}
