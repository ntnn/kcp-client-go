package testing

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	restclient "k8s.io/client-go/rest"
)

type FakeScopedClient interface {
	// Tracker gives access to the ObjectTracker internal to the fake client.
	Tracker() ScopedObjectTracker

	// AddReactor appends a reactor to the end of the chain.
	AddReactor(verb, resource string, reaction ReactionFunc)

	// PrependReactor adds a reactor to the beginning of the chain.
	PrependReactor(verb, resource string, reaction ReactionFunc)

	// AddWatchReactor appends a reactor to the end of the chain.
	AddWatchReactor(resource string, reaction WatchReactionFunc)

	// PrependWatchReactor adds a reactor to the beginning of the chain.
	PrependWatchReactor(resource string, reaction WatchReactionFunc)

	// AddProxyReactor appends a reactor to the end of the chain.
	AddProxyReactor(resource string, reaction ProxyReactionFunc)

	// PrependProxyReactor adds a reactor to the beginning of the chain.
	PrependProxyReactor(resource string, reaction ProxyReactionFunc)

	// Invokes records the provided Action and then invokes the ReactionFunc that
	// handles the action if one exists. defaultReturnObj is expected to be of the
	// same type a normal call would return.
	Invokes(action Action, defaultReturnObj runtime.Object) (runtime.Object, error)

	// InvokesWatch records the provided Action and then invokes the ReactionFunc
	// that handles the action if one exists.
	InvokesWatch(action Action) (watch.Interface, error)

	// InvokesProxy records the provided Action and then invokes the ReactionFunc
	// that handles the action if one exists.
	InvokesProxy(action Action) restclient.ResponseWrapper

	// ClearActions clears the history of actions called on the fake client.
	ClearActions()

	// Actions returns a chronologically ordered slice fake actions called on the
	// fake client.
	Actions() []Action
}
