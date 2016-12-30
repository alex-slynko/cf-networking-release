// This file was generated by counterfeiter
package fakes

import (
	"net/http"
	"policy-server/uaa_client"
	"sync"
)

type AuthenticatedHandler struct {
	ServeHTTPStub        func(response http.ResponseWriter, request *http.Request, tokenData uaa_client.CheckTokenResponse)
	serveHTTPMutex       sync.RWMutex
	serveHTTPArgsForCall []struct {
		response  http.ResponseWriter
		request   *http.Request
		tokenData uaa_client.CheckTokenResponse
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *AuthenticatedHandler) ServeHTTP(response http.ResponseWriter, request *http.Request, tokenData uaa_client.CheckTokenResponse) {
	fake.serveHTTPMutex.Lock()
	fake.serveHTTPArgsForCall = append(fake.serveHTTPArgsForCall, struct {
		response  http.ResponseWriter
		request   *http.Request
		tokenData uaa_client.CheckTokenResponse
	}{response, request, tokenData})
	fake.recordInvocation("ServeHTTP", []interface{}{response, request, tokenData})
	fake.serveHTTPMutex.Unlock()
	if fake.ServeHTTPStub != nil {
		fake.ServeHTTPStub(response, request, tokenData)
	}
}

func (fake *AuthenticatedHandler) ServeHTTPCallCount() int {
	fake.serveHTTPMutex.RLock()
	defer fake.serveHTTPMutex.RUnlock()
	return len(fake.serveHTTPArgsForCall)
}

func (fake *AuthenticatedHandler) ServeHTTPArgsForCall(i int) (http.ResponseWriter, *http.Request, uaa_client.CheckTokenResponse) {
	fake.serveHTTPMutex.RLock()
	defer fake.serveHTTPMutex.RUnlock()
	return fake.serveHTTPArgsForCall[i].response, fake.serveHTTPArgsForCall[i].request, fake.serveHTTPArgsForCall[i].tokenData
}

func (fake *AuthenticatedHandler) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.serveHTTPMutex.RLock()
	defer fake.serveHTTPMutex.RUnlock()
	return fake.invocations
}

func (fake *AuthenticatedHandler) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}
