package fakes

import "sync"

type Shimmer struct {
	ShimCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Path    string
			Version string
		}
		Returns struct {
			Error error
		}
		Stub func(string, string) error
	}
}

func (f *Shimmer) Shim(param1 string, param2 string) error {
	f.ShimCall.mutex.Lock()
	defer f.ShimCall.mutex.Unlock()
	f.ShimCall.CallCount++
	f.ShimCall.Receives.Path = param1
	f.ShimCall.Receives.Version = param2
	if f.ShimCall.Stub != nil {
		return f.ShimCall.Stub(param1, param2)
	}
	return f.ShimCall.Returns.Error
}
