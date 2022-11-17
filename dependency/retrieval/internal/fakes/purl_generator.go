package fakes

import "sync"

type PackageURLGenerator struct {
	GenerateCall struct {
		mutex     sync.Mutex
		CallCount int
		Receives  struct {
			Name    string
			Version string
			Sha     string
			Uri     string
		}
		Returns struct {
			String string
		}
		Stub func(string, string, string, string) string
	}
}

func (f *PackageURLGenerator) Generate(param1 string, param2 string, param3 string, param4 string) string {
	f.GenerateCall.mutex.Lock()
	defer f.GenerateCall.mutex.Unlock()
	f.GenerateCall.CallCount++
	f.GenerateCall.Receives.Name = param1
	f.GenerateCall.Receives.Version = param2
	f.GenerateCall.Receives.Sha = param3
	f.GenerateCall.Receives.Uri = param4
	if f.GenerateCall.Stub != nil {
		return f.GenerateCall.Stub(param1, param2, param3, param4)
	}
	return f.GenerateCall.Returns.String
}
