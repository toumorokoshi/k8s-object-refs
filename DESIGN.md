# Watch-based object references

The goal of this repository is to build a proof of concept for fast object reference resolution via Watch calls.

Normally, an object that references another object takes a long time to resolve, as it's dependent on the rate of the controller loop to refresh the resource state. This can be exacerbated with additional guards against unneeded work such as exponential backoffs.

## TODO

- controller runtime
  - add support for getting the WorkQueue for a controller in reconcile() call
    - this is also required to not use the incorrect context.
  - find an idiomatic way to get the latest status from an object.
    - [similar symptom, although different need](https://github.com/kubernetes-sigs/controller-runtime/issues/585).

## Implementation

The fundamental approach is to have a manager with a registry of the objects that are dependents, with a mapping of those to their referents.

A controller is spawned per resource, which is for efficiency: hundreds or thousands of objects can share the same controller, even if they are watching different objects.

Once there are no object references, the object references controller for that resource is cancelled, ensuring that there are no additional watch calls on resources that are no longer relevant to the controller.

## Design Considerations

### Sharing watch calls

By leveraging controller-runtime Controllers and their watching mechanism, this approach is able to re-use the existing Informers that are used to manage standard resources. In other words, this choice enables that only one Watch will exist per resource, per process.

### object referrent - referrer mapping

One key element of this architecture is a mapping of object referrents to their referrers. It is important to note that this is not just a mapping of object refferent -> object referrer controller: to completely specify the referrent, the request must also be included in the identifier.

In addition, reconcilers should not store state in between requests: this means that the reconciler cannot store state around which objects it is referring to. Thus, the object reference manager must expose a way to update the full list of object references, which also needs to clear ones previously set. This is done by requiring the controller to pass the full set of subscriptions every time, which requires re-calculating the subscriptions per reocncile run.

To figure out the deltas, one could store a double linked list with all of the elements pertaining to that particular controler-nn. Then you can remove subscriptions with O(1) cost per subscription.

## Notes

### Does controller-runtime re-use watches?

If controller-runtime somehow enabled re-use of watch, then that significantly reduces the management of watch calls in controllers, as they can re-use instantiated watches if they already exist.

- [controller managers delegate watch calls to the controllers themselves](vendor/sigs.k8s.io/controller-runtime/pkg/builder/controller.go#L233).
  - [they in turn delegate to the source](vendor/sigs.k8s.io/controller-runtime/pkg/internal/controller/controller.go#L135).
    - [source.Kind reuses an informer if it exists for the same Type. This requires the cache to be injected](vendor/sigs.k8s.io/controller-runtime/pkg/source/source.go#L114).
      - [cache in injected here](vendor/sigs.k8s.io/controller-runtime/pkg/internal/controller/controller.go#L114).

- Is the cache injected?
  - InjectCache is only called by CacheInto, which is only called by SetFields on the cluster struct.

### How to identify controllers / queues to send events o?

- [controller is a runnable](vendor/sigs.k8s.io/controller-runtime/pkg/manager/internal.go#588). so no way to get a list of controllers from that.

- Controller has Queue, and a reconciler
- Manager has runnables
- Watch calls Handlers with the event and it's own queue?
- src is called against a specific Queue.
