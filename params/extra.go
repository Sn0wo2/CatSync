package params

import (
	"hash/fnv"
	"reflect"
	"sync"
	"unsafe"
)

type Key uint64

func KeyFromString(s string) Key {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))

	return Key(h.Sum64())
}

type entry struct {
	ptr unsafe.Pointer
	tid uint64
}

type ExtraStore struct {
	mu sync.RWMutex
	m  map[Key]entry
}

func (s *ExtraStore) Set(k Key, v any) {
	if s.m == nil {
		s.m = make(map[Key]entry)
	}

	t := reflect.TypeOf(v)
	p := reflect.New(t)
	p.Elem().Set(reflect.ValueOf(v))

	s.mu.Lock()
	//nolint:gosec // G103: unsafe is intentional for non-any universal store
	s.m[k] = entry{ptr: unsafe.Pointer(p.Pointer()), tid: typeIDFromType(t)}
	s.mu.Unlock()
}

func (s *ExtraStore) Get(k Key, out any) bool {
	// out must be a non-nil pointer.
	vOut := reflect.ValueOf(out)
	if vOut.Kind() != reflect.Ptr || vOut.IsNil() {
		return false
	}

	wantT := vOut.Elem().Type()
	st := typeIDFromType(wantT)

	s.mu.RLock()
	e, ok := s.m[k]
	s.mu.RUnlock()

	if !ok || e.tid != st {
		return false
	}

	vOut.Elem().Set(reflect.NewAt(wantT, e.ptr).Elem())

	return true
}

func (s *ExtraStore) Delete(k Key) {
	s.mu.Lock()
	delete(s.m, k)
	s.mu.Unlock()
}

func typeIDFromType(t reflect.Type) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(t.PkgPath()))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(t.String()))

	return h.Sum64()
}
