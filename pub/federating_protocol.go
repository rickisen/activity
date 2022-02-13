package pub

import (
	"context"
	"net/http"
	"net/url"

	"github.com/go-fed/activity/streams/vocab"
)

// FederatingProtocol contains behaviors an application needs to satisfy for the
// full ActivityPub S2S implementation to be supported by this library.
//
// It is only required if the client application wants to support the server-to-
// server, or federating, protocol.
//
// It is passed to the library as a dependency injection from the client
// application.
type FederatingProtocol interface {
	// Hook callback after parsing the request body for a federated request
	// to the Actor's inbox.
	//
	// Can be used to set contextual information based on the Activity
	// received.
	//
	// Only called if the Federated Protocol is enabled.
	//
	// Warning: Neither authentication nor authorization has taken place at
	// this time. Doing anything beyond setting contextual information is
	// strongly discouraged.
	//
	// If an error is returned, it is passed back to the caller of
	// PostInbox. In this case, the DelegateActor implementation must not
	// write a response to the ResponseWriter as is expected that the caller
	// to PostInbox will do so when handling the error.
	PostInboxRequestBodyHook(c context.Context, r *http.Request, activity Activity) (context.Context, error)
	// AuthenticatePostInbox delegates the authentication of a POST to an
	// inbox.
	//
	// If an error is returned, it is passed back to the caller of
	// PostInbox. In this case, the implementation must not write a
	// response to the ResponseWriter as is expected that the client will
	// do so when handling the error. The 'authenticated' is ignored.
	//
	// If no error is returned, but authentication or authorization fails,
	// then authenticated must be false and error nil. It is expected that
	// the implementation handles writing to the ResponseWriter in this
	// case.
	//
	// Finally, if the authentication and authorization succeeds, then
	// authenticated must be true and error nil. The request will continue
	// to be processed.
	AuthenticatePostInbox(c context.Context, w http.ResponseWriter, r *http.Request) (out context.Context, authenticated bool, err error)
	// Blocked should determine whether to permit a set of actors given by
	// their ids are able to interact with this particular end user due to
	// being blocked or other application-specific logic.
	//
	// If an error is returned, it is passed back to the caller of
	// PostInbox.
	//
	// If no error is returned, but authentication or authorization fails,
	// then blocked must be true and error nil. An http.StatusForbidden
	// will be written in the wresponse.
	//
	// Finally, if the authentication and authorization succeeds, then
	// blocked must be false and error nil. The request will continue
	// to be processed.
	Blocked(c context.Context, actorIRIs []*url.URL) (blocked bool, err error)
	// FederatingCallbacks returns the application logic that handles
	// ActivityStreams received from federating peers.
	//
	// Note that certain types of callbacks will be 'wrapped' with default
	// behaviors supported natively by the library. Other callbacks
	// compatible with streams.TypeResolver can be specified by 'other'.
	//
	// For example, setting the 'Create' field in the
	// FederatingWrappedCallbacks lets an application dependency inject
	// additional behaviors they want to take place, including the default
	// behavior supplied by this library. This is guaranteed to be compliant
	// with the ActivityPub Social protocol.
	//
	// To override the default behavior, instead supply the function in
	// 'other', which does not guarantee the application will be compliant
	// with the ActivityPub Social Protocol.
	//
	// Applications are not expected to handle every single ActivityStreams
	// type and extension. The unhandled ones are passed to DefaultCallback.
	FederatingCallbacks(c context.Context) (wrapped FederatingWrappedCallbacks, other []interface{}, err error)
	// DefaultCallback is called for types that go-fed can deserialize but
	// are not handled by the application's callbacks returned in the
	// Callbacks method.
	//
	// Applications are not expected to handle every single ActivityStreams
	// type and extension, so the unhandled ones are passed to
	// DefaultCallback.
	DefaultCallback(c context.Context, activity Activity) error
	// MaxInboxForwardingRecursionDepth determines how deep to search within
	// an activity to determine if inbox forwarding needs to occur.
	//
	// Zero or negative numbers indicate infinite recursion.
	MaxInboxForwardingRecursionDepth(c context.Context) int
	// MaxDeliveryRecursionDepth determines how deep to search within
	// collections owned by peers when they are targeted to receive a
	// delivery.
	//
	// Zero or negative numbers indicate infinite recursion.
	MaxDeliveryRecursionDepth(c context.Context) int
	// FilterForwarding allows the implementation to apply business logic
	// such as blocks, spam filtering, and so on to a list of potential
	// Collections and OrderedCollections of recipients when inbox
	// forwarding has been triggered.
	//
	// The activity is provided as a reference for more intelligent
	// logic to be used, but the implementation must not modify it.
	FilterForwarding(c context.Context, potentialRecipients []*url.URL, a Activity) (filteredRecipients []*url.URL, err error)
	// GetInbox returns the OrderedCollection inbox of the actor for this
	// context. It is up to the implementation to provide the correct
	// collection for the kind of authorization given in the request.
	//
	// AuthenticateGetInbox will be called prior to this.
	//
	// Always called, regardless whether the Federated Protocol or Social
	// API is enabled.
	GetInbox(c context.Context, r *http.Request) (vocab.ActivityStreamsOrderedCollectionPage, error)
	// ResolveInboxIRIs allows implementations to optionally override the default
	// method of resolving inboxes for receiving actors.
	//
	// It is acceptable to not do anything here and just return the given
	// receiver lists as remaining receivers like so:
	//
	// 	return []*url.URL{}, append(receivers, hiddenReceivers...), nil
	//
	// This function may be used by an implementer for multiple reasons,
	// including to optimize network usage by reducing the amount of delivery
	// targets through delivering to shared inboxes when possible, or maybe to
	// add all known instances' shared inbox to the list of receiers when an
	// activity is addressed to the public.
	//
	// Receivers are given as two groups (normal and hidden) in order to aid in
	// determining the best method of handling shared inboxes, as hidden
	// receivers should not have their inboxes replaced with a shared inbox. This
	// is because both the bto and bcc fields are stripped prior delivery. If a
	// hidden receiver has their inbox replaced with a shared inbox the receiving
	// instance would have no way of knowing who the intended recipient was.
	//
	// Keep in mind that:
	//
	//   * Any receiver IRI may resolve into an Actor, or a Collection of Actors,
	//     maybe even Collections of Collections of Actors etc.
	//
	//   * The library will still dereference and handle any remaining receivers
	//     you return.
	//
	//   * The library only deduplicates the final list of inboxes after this
	//     function is called. So duplicated values in the given receivers are
	//     likely. And returning the same inbox multiple times is acceptable.
	ResolveInboxIRIs(c context.Context, receivers []*url.URL, hiddenReceivers []*url.URL) (inboxes []*url.URL, remainingReceivers []*url.URL, err error)
}
