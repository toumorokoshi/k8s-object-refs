# Watch-based object references

The goal of this repository is to build a proof of concept for fast object reference resolution via Watch calls.

Normally, an object that references another object takes a long time to resolve, as it's dependent on the rate of the controller loop to refresh the resource state. This can be exacerbated with additional guards against unneeded work such as exponential backoffs.

## TODO

- controller runtime
  - add support for getting the WorkQueue for a controller in reconcile() call
  - enable cancelling a controller
  - find an idiomatic way to get the latest status from an object.
    - [similar symptom, although different need](https://github.com/kubernetes-sigs/controller-runtime/issues/585).

## Design considerations

### Single-resource object references

For simple object references where the resource that is references is pre-determined, the watch call can be re-used for all instances of the referer resource, assuming once can store a map of the referrer-referent pair.


- is it better to watch single instances of an object, vs watching all of a particular resource?

### Implementation

- init
  - start a new watch for the target resource type(s)
- object create
  - in Reconcile(), if the target object is not ready, enqueue the objects NamespacedName into the map to be called when the referrent is ready
- when referrent has a state change
  - when watch is triggered, enqueue an immediate reconcile of the object that is waiting on that resource.

- to re-use the cache, the object being watched has to share the types.

### Additional cases

- a single object referencing multiple referrents
- referring to more than one of the same referrent

### Possible Issues

- Make namespacedname a possible struct in a spec by adding json annotations?
  - vendor/k8s.io/apimachinery/pkg/types/namespacedname.go:27
  - with this, I won't have to create my own types.go to reference the object.
- add context to eventhandler?
-

## Future Work

### Multiple object reference types

If the object reference type is not known beforehand, then that requires watch calls to be constructed ad-hoc as a referrer object is created.

## Questions

### Does controller-runtime re-use watches?

If controller-runtime somehow enabled re-use of watch, then that significantly reduces the management of watch calls in controllers, as they can re-use instantiated watches if they already exist.

- [controller managers delegate watch calls to the controllers themselves](vendor/sigs.k8s.io/controller-runtime/pkg/builder/controller.go:233).
  - [they in turn delegate to the source](vendor/sigs.k8s.io/controller-runtime/pkg/internal/controller/controller.go:135).
    - [source.Kind reuses an informer if it exists for the same Type. This requires the cache to be injected](vendor/sigs.k8s.io/controller-runtime/pkg/source/source.go:114).
      - [cache in injected here](vendor/sigs.k8s.io/controller-runtime/pkg/internal/controller/controller.go:114).

- Is the cache injected?
  - InjectCache is only called by CacheInto, which is only called by SetFields on the cluster struct.

### How to identify controllers / queues to send events o?

- [controller is a runnable](vendor/sigs.k8s.io/controller-runtime/pkg/manager/internal.go:588). so no way to get a list of controllers from that.

- Controller has Queue, and a reconciler
- Manager has runnables
- Watch calls Handlers with the event and it's own queue?
- src is called against a specific Queue.