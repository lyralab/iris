package cache

import (
	"container/heap"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Cache is a concurrent, LRU cache with per-entry TTL.
// It limits memory by a configurable capacity (number of entries), evicts
// least-recently-used items when over capacity, and optionally runs a single
// janitor goroutine that removes expired items promptly without spawning
// per-entry timers.
type Cache[K comparable, V any] struct {
	mu       sync.Mutex
	items    map[K]*node[K, V]
	capacity int

	// Doubly-linked list for LRU (most-recent at head).
	head *node[K, V]
	tail *node[K, V]

	// Min-heap by expiration time for efficient expiry purging.
	expHeap expiryHeap[K, V]

	// Janitor coordination
	stopCh chan struct{}
	wakeCh chan struct{}
	closed bool

	// Options
	cleanupInterval time.Duration
	janitorEnabled  bool
	Logger          *zap.SugaredLogger
}

// node represents one cache entry.
type node[K comparable, V any] struct {
	key      K
	value    V
	expireAt time.Time // zero means no expiration

	// LRU pointers
	prev *node[K, V]
	next *node[K, V]

	// expiry heap index; -1 when not in heap
	hidx int
}

// expiryHeap is a min-heap ordered by expireAt.
type expiryHeap[K comparable, V any] []*node[K, V]

func (h expiryHeap[K, V]) Len() int           { return len(h) }
func (h expiryHeap[K, V]) Less(i, j int) bool { return h[i].expireAt.Before(h[j].expireAt) }
func (h expiryHeap[K, V]) Swap(i, j int)      { h[i], h[j] = h[j], h[i]; h[i].hidx, h[j].hidx = i, j }
func (h *expiryHeap[K, V]) Push(x any)        { n := x.(*node[K, V]); n.hidx = len(*h); *h = append(*h, n) }
func (h *expiryHeap[K, V]) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	x.hidx = -1
	return x
}
func (h expiryHeap[K, V]) peek() *node[K, V] {
	if len(h) == 0 {
		return nil
	}
	return h[0]
}
func (h *expiryHeap[K, V]) remove(n *node[K, V]) {
	if n.hidx < 0 || n.hidx >= len(*h) {
		return
	}
	heap.Remove(h, n.hidx)
}

// New creates a new Cache with the provided options.
func New[K comparable, V any](logger *zap.SugaredLogger, opts ...Option) *Cache[K, V] {
	o := defaultOptions()
	for _, fn := range opts {
		fn(&o)
	}
	if o.Capacity < 1 {
		o.Capacity = 1
	}
	if o.CleanupInterval <= 0 {
		o.CleanupInterval = time.Minute
	}
	c := &Cache[K, V]{
		items:           make(map[K]*node[K, V], o.Capacity),
		capacity:        o.Capacity,
		expHeap:         expiryHeap[K, V]{},
		stopCh:          make(chan struct{}),
		wakeCh:          make(chan struct{}, 1),
		cleanupInterval: o.CleanupInterval,
		janitorEnabled:  o.EnableJanitor,
		Logger:          logger,
	}
	heap.Init(&c.expHeap)
	if c.janitorEnabled {
		go c.janitor()
	}
	return c
}

// Set inserts or updates the value for key with a TTL.
// ttl <= 0 means the item does not expire.
func (c *Cache[K, V]) Set(key K, value V, ttl time.Duration) error {
	now := time.Now()
	exp := time.Time{}
	if ttl > 0 {
		exp = now.Add(ttl)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	if n, ok := c.items[key]; ok {
		// Update existing
		n.value = value
		// Update expiration
		if n.expireAt.IsZero() && !exp.IsZero() {
			// newly expiring
			n.expireAt = exp
			heap.Push(&c.expHeap, n)
			c.tryWake()
		} else if !n.expireAt.IsZero() && exp.IsZero() {
			// no longer expiring
			c.expHeap.remove(n)
			n.expireAt = time.Time{}
		} else if !n.expireAt.Equal(exp) {
			n.expireAt = exp
			if !exp.IsZero() {
				heap.Fix(&c.expHeap, n.hidx)
				c.tryWake()
			}
		}
		c.moveToFront(n)
		return nil
	}

	// Insert new
	n := &node[K, V]{key: key, value: value, expireAt: exp, hidx: -1}
	c.items[key] = n
	c.insertFront(n)

	if !exp.IsZero() {
		heap.Push(&c.expHeap, n)
		c.tryWake()
	}

	// Evict if over capacity
	for len(c.items) > c.capacity {
		c.evictOldestLocked()
	}
	return nil
}

// Get returns the value for key if present and not expired.
// It marks the entry as most-recently-used on hit.
// The bool is false if not found or expired.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	var zero V

	now := time.Now()

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return zero, false
	}

	n, ok := c.items[key]
	if !ok {
		return zero, false
	}

	// Expired?
	if !n.expireAt.IsZero() && !n.expireAt.After(now) {
		// Remove expired
		c.removeNodeLocked(n)
		delete(c.items, key)
		if n.hidx >= 0 {
			c.expHeap.remove(n)
		}
		return zero, false
	}

	// Move to front (mark as recently used)
	c.moveToFront(n)
	return n.value, true
}

// Delete removes a key from the cache if present.
func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	n, ok := c.items[key]
	if !ok {
		return
	}
	c.removeNodeLocked(n)
	delete(c.items, key)
	if n.hidx >= 0 {
		c.expHeap.remove(n)
	}
}

// Len returns the number of items currently stored (including any that may be expired but not yet purged).
func (c *Cache[K, V]) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.items)
}

// Close stops the background janitor (if enabled) and makes the cache inert.
func (c *Cache[K, V]) Close() {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return
	}
	c.closed = true
	close(c.stopCh)
	// best-effort wake
	c.tryWake()
	c.mu.Unlock()
}

// evictOldestLocked removes tail (LRU) entry.
func (c *Cache[K, V]) evictOldestLocked() {
	if c.tail == nil {
		return
	}
	n := c.tail
	c.removeNodeLocked(n)
	delete(c.items, n.key)
	if n.hidx >= 0 {
		c.expHeap.remove(n)
	}
}

// moveToFront moves node to head (MRU).
func (c *Cache[K, V]) moveToFront(n *node[K, V]) {
	if n == c.head {
		return
	}
	c.removeNodeLinks(n)
	c.insertFront(n)
}

func (c *Cache[K, V]) insertFront(n *node[K, V]) {
	n.prev = nil
	n.next = c.head
	if c.head != nil {
		c.head.prev = n
	}
	c.head = n
	if c.tail == nil {
		c.tail = n
	}
}

func (c *Cache[K, V]) removeNodeLocked(n *node[K, V]) {
	c.removeNodeLinks(n)
	if n == c.head {
		c.head = n.next
	}
	if n == c.tail {
		c.tail = n.prev
	}
	n.prev, n.next = nil, nil
}

func (c *Cache[K, V]) removeNodeLinks(n *node[K, V]) {
	if n.prev != nil {
		n.prev.next = n.next
	}
	if n.next != nil {
		n.next.prev = n.prev
	}
}

// janitor wakes at the next known expiration and purges expired keys.
// When there are no expirations, it sleeps for CleanupInterval.
// It listens to wakeCh to recompute wake-ups when new expirations are added or changed.
func (c *Cache[K, V]) janitor() {
	timer := time.NewTimer(c.cleanupInterval)
	defer timer.Stop()

	for {
		// Compute next wake deadline
		c.mu.Lock()
		if c.closed {
			c.mu.Unlock()
			return
		}
		next := c.nextExpiryLocked()
		var wait time.Duration
		if next.IsZero() {
			wait = c.cleanupInterval
		} else {
			now := time.Now()
			if !next.After(now) {
				wait = 0
			} else {
				wait = next.Sub(now)
			}
		}
		// Reset timer
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		timer.Reset(wait)
		c.mu.Unlock()

		select {
		case <-timer.C:
			c.purgeExpired()
		case <-c.wakeCh:
			// just recompute
		case <-c.stopCh:
			return
		}
	}
}

// nextExpiryLocked returns the earliest expiration time, or zero time if none.
func (c *Cache[K, V]) nextExpiryLocked() time.Time {
	if len(c.expHeap) == 0 {
		return time.Time{}
	}
	return c.expHeap.peek().expireAt
}

// purgeExpired removes all items whose expiration is <= now.
func (c *Cache[K, V]) purgeExpired() {
	now := time.Now()

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	for len(c.expHeap) > 0 {
		n := c.expHeap.peek()
		if n.expireAt.After(now) || n.expireAt.IsZero() {
			break
		}
		// Pop from heap
		heap.Pop(&c.expHeap)
		// Remove from LRU/map if still present
		if existing, ok := c.items[n.key]; ok && existing == n {
			c.removeNodeLocked(n)
			delete(c.items, n.key)
		}
	}
}

// tryWake notifies janitor to recompute next wake without blocking.
func (c *Cache[K, V]) tryWake() {
	select {
	case c.wakeCh <- struct{}{}:
	default:
	}
}
