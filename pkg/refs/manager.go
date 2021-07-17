package refs

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// TODO: handle removing watches when there are no more objects that we're waiting on for that watch.

// TODO: when a queue object is exposed, use that instead.
type QueueContext struct {
	// Context    context.Context
	Req        ctrl.Request
	Reconciler reconcile.Reconciler
}

type SubscriptionNode struct {
	prevInContext  *SubscriptionNode
	nextInContext  *SubscriptionNode
	prevInEventMap *SubscriptionNode
	nextInEventMap *SubscriptionNode
	queueContext   *QueueContext
	subscription   *RefSubscription
}

// TODO: split to types

// RefManager is an auxiliary manager for triggering reconciles sooner
// for objects references, when those referents change.
type RefManager struct {
	// EventMapping
	EventMapping EventMapping
	// SubscriptionsByReferrer is a map of QueueContexts to a LinkedList
	// of subscriptions owned by that QueueContext.
	//
	// This is used internally for efficient cleanup of subscriptions.
	SubscriptionsByReferrer map[QueueContext]*SubscriptionNode
	Manager                 ctrl.Manager
	cancelFuncsByGVK        map[GVK]context.CancelFunc
	// SubscriptionsPerGVK counts the number of subscriptions per
	// GVK, and when the value is zero, removes the controller.
	SubscriptionsPerGVK map[GVK]int

	// watchers stores handlers to the watchers RefManager has
	// started, to enable proper garbage collection as they are no
	// longer used.
	// watchers map[GVK]Watch
}

type EventMapping map[GVK]map[types.NamespacedName]*SubscriptionNode

func NewRefManager() RefManager {
	return RefManager{
		EventMapping:            make(EventMapping),
		cancelFuncsByGVK:        make(map[GVK]context.CancelFunc),
		SubscriptionsByReferrer: make(map[QueueContext]*SubscriptionNode),
		SubscriptionsPerGVK:     make(map[GVK]int),
	}
}

func (r *RefManager) SetupWithManager(mgr ctrl.Manager) error {
	r.Manager = mgr
	return nil
}

// edge case: if the size of the list is only one element (how do you make yourself the null pointer?)
func (r *RefManager) UpdateSubscriptions(qc QueueContext, refs []RefSubscription) error {
	// find all previous referrent for that controller+req mapping
	// if they exist in the new list, keep them
	// if they no longer exist, remove them
	// if new ones exit, add them
	// iterate through existing refs
	foundSubscriptions := make([]bool, len(refs))
	curElement := r.SubscriptionsByReferrer[qc]
	if curElement == nil {
		curElement = &SubscriptionNode{}
		r.SubscriptionsByReferrer[qc] = curElement

	} else {
		// always iterate past the first element, since it's a placeholder
		// TODO: add before removing, to ensure that we don't stop controllers too aggressively.
		curElement = curElement.nextInContext
		for curElement != nil {
			found := false
			for i, newRef := range refs {
				if *curElement.subscription == newRef {
					foundSubscriptions[i] = true
					found = true
					break
				}
			}
			// if it can't be found in the new subscriptions, it should be removed.
			if !found {
				// remove from context
				curElement.prevInContext.nextInContext = curElement.nextInContext
				// remove from eventmapping
				curElement.prevInEventMap.nextInEventMap = curElement.nextInEventMap
				// decrement GVK
				r.decrementGVKSubscriptions(curElement.subscription.Gvk)
			} else {
				curElement = curElement.nextInContext
			}
		}
	}
	// now go through subscriptions again, to find new subscriptions
	for i, newRef := range refs {
		if foundSubscriptions[i] {
			continue
		}
		node := SubscriptionNode{
			queueContext: &qc,
			subscription: &newRef,
		}
		// it's a new node, so add to both EventMapping and context linked
		// lists
		r.EventMapping.insert(&node)
		node.nextInContext = r.SubscriptionsByReferrer[qc].nextInContext
		node.prevInContext = r.SubscriptionsByReferrer[qc]
		if node.nextInContext != nil {
			node.nextInContext.prevInContext = &node
		}
		r.SubscriptionsByReferrer[qc].nextInContext = &node
		r.incrementGVKSubscription(newRef.Gvk)
	}
	// TODO: take deltas for GVKs, start and stop controllers
	return nil
}

func (r *RefManager) incrementGVKSubscription(gvk GVK) {
	r.SubscriptionsPerGVK[gvk] += 1
	if r.SubscriptionsPerGVK[gvk] == 1 {
		r.startController(gvk)
	}
}

func (r *RefManager) decrementGVKSubscriptions(gvk GVK) {
	r.SubscriptionsPerGVK[gvk] -= 1
	if r.SubscriptionsPerGVK[gvk] == 0 {
		r.cancelFuncsByGVK[gvk]()
		r.cancelFuncsByGVK[gvk] = nil
	}
}

func (e *EventMapping) insert(node *SubscriptionNode) {
	gvk := node.subscription.Gvk
	namespacedName := node.subscription.NamespacedName
	subscriptionsByNamespaceName := (*e)[gvk]
	if subscriptionsByNamespaceName == nil {
		subscriptionsByNamespaceName = make(map[types.NamespacedName]*SubscriptionNode)
		(*e)[gvk] = subscriptionsByNamespaceName
	}
	headOfList := subscriptionsByNamespaceName[namespacedName]
	// throw a dummy head in this list, to ensure that we can always remove a node in a linked list.
	if headOfList == nil {
		headOfList = &SubscriptionNode{}
		subscriptionsByNamespaceName[namespacedName] = headOfList
	}
	// TODO: fix this to become O(1) do the insert at the beginning,
	// without iterating
	node.nextInEventMap = headOfList.nextInEventMap
	if node.nextInEventMap != nil {
		node.nextInEventMap.prevInEventMap = node
	}
	node.prevInEventMap = headOfList
	headOfList.nextInEventMap = node
}

// TODO handle errors
func (r *RefManager) startController(gvk GVK) error {
	// if a context already exists, then we already spawned
	// a controller.
	if context := r.cancelFuncsByGVK[gvk]; context != nil {
		return nil
	}
	c, err := controller.NewUnmanaged(fmt.Sprintf("%v", gvk), r.Manager, controller.Options{
		Reconciler: &RefReconciler{
			Client:  r.Manager.GetClient(),
			manager: r,
			gvk:     gvk,
		},
	})
	if err != nil {
		r.Manager.GetLogger().Error(err, "unable to create manager")
	}
	if err := c.Watch(&source.Kind{Type: gvk.ToClientObject()}, &handler.EnqueueRequestForObject{}); err != nil {
		r.Manager.GetLogger().Error(err, "unable to watch")
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	go func() {
		if err := c.Start(ctx); err != nil {
			logger := log.FromContext(ctx)
			logger.Error(err, "")
			// TODO: error handling
		}
	}()
	r.cancelFuncsByGVK[gvk] = cancelFunc
	return nil
}
