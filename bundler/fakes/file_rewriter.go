package fakes

import "sync"

type FileRewriter struct {
	RewriteCall struct {
		sync.Mutex
		CallCount int
		Receives  struct {
			Filename string
		}
		Returns struct {
			Error error
		}
		Stub func(string) error
	}
}

func (f *FileRewriter) Rewrite(param1 string) error {
	f.RewriteCall.Lock()
	defer f.RewriteCall.Unlock()
	f.RewriteCall.CallCount++
	f.RewriteCall.Receives.Filename = param1
	if f.RewriteCall.Stub != nil {
		return f.RewriteCall.Stub(param1)
	}
	return f.RewriteCall.Returns.Error
}
