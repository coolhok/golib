package slab

import (
	"sort"
	"testing"
)

func TestBasics(t *testing.T) {
	s := NewArena(1, 1024, 2, nil)
	if s == nil {
		t.Errorf("expected new slab arena to work")
	}
	a, _ := s.Alloc(1)
	if a == nil {
		t.Errorf("expected alloc to work")
	}
	if len(a) != 1 {
		t.Errorf("expected alloc to give right size buf")
	}
	if cap(a) != 1+SLAB_MEMORY_FOOTER_LEN {
		t.Errorf("expected alloc cap to match algorithm, got: %v vs %v",
			cap(a), 1+SLAB_MEMORY_FOOTER_LEN)
	}
	a[0] = 66
	if s.DecRef(a) != true {
		t.Errorf("expected DecRef to be the last one")
	}
	b, _ := s.Alloc(1)
	if b == nil {
		t.Errorf("expected alloc to work")
	}
	if len(b) != 1 {
		t.Errorf("expected alloc to give right size buf")
	}
	if cap(b) != 1+SLAB_MEMORY_FOOTER_LEN {
		t.Errorf("expected alloc cap to match algorithm, got: %v vs %v",
			cap(b), 1+SLAB_MEMORY_FOOTER_LEN)
	}
	if b[0] != 66 {
		t.Errorf("expected alloc to return last freed buf")
	}
	s.AddRef(b)
	if s.DecRef(b) != false {
		t.Errorf("expected DecRef() to not be the last")
	}
	if s.DecRef(b) != true {
		t.Errorf("expected DecRef() to be the last")
	}
	c, _ := s.Alloc(1)
	if c[0] != 66 {
		t.Errorf("expected alloc to return last freed buf")
	}
}

func TestSlabClassGrowth(t *testing.T) {
	s := NewArena(1, 8, 2, nil)
	expectSlabClasses := func(numSlabClasses int) {
		if len(s.slabClasses) != numSlabClasses {
			t.Errorf("expected %v slab classses, got: %v",
				numSlabClasses, len(s.slabClasses))
		}
	}
	expectSlabClasses(1)
	s.Alloc(1)
	expectSlabClasses(1)
	s.Alloc(1)
	expectSlabClasses(1)
	s.Alloc(2)
	expectSlabClasses(2)
	s.Alloc(1)
	s.Alloc(2)
	expectSlabClasses(2)
	s.Alloc(3)
	s.Alloc(4)
	expectSlabClasses(3)
	s.Alloc(5)
	s.Alloc(8)
	expectSlabClasses(4)
}

func TestDecRef(t *testing.T) {
	s := NewArena(1, 8, 2, nil)
	expectSlabClasses := func(numSlabClasses int) {
		if len(s.slabClasses) != numSlabClasses {
			t.Errorf("expected %v slab classses, got: %v",
				numSlabClasses, len(s.slabClasses))
		}
	}
	a := make([][]byte, 128)
	for j := 0; j < 100; j++ {
		for i := 0; i < len(a); i++ {
			a[i], _ = s.Alloc(i % 8)
		}
		for i := 0; i < len(a); i++ {
			s.DecRef(a[i])
		}
	}
	expectSlabClasses(4)
}

func TestAddRef(t *testing.T) {
	s := NewArena(1, 1, 2, nil)
	if !s.slabClasses[0].chunkFree.isEmpty() {
		t.Errorf("expected no free chunks")
	}
	a, _ := s.Alloc(1)
	a[0] = 123
	if !s.slabClasses[0].chunkFree.isEmpty() {
		t.Errorf("expected no free chunks")
	}
	s.AddRef(a)
	if !s.slabClasses[0].chunkFree.isEmpty() {
		t.Errorf("expected no free chunks")
	}
	s.DecRef(a)
	if !s.slabClasses[0].chunkFree.isEmpty() {
		t.Errorf("expected no free chunks")
	}
	s.DecRef(a)
	if s.slabClasses[0].chunkFree.isEmpty() {
		t.Errorf("expected 1 free chunk")
	}
	b, _ := s.Alloc(1)
	if b[0] != 123 {
		t.Errorf("expected chunk to be reused")
	}
}

func TestLargeAlloc(t *testing.T) {
	s := NewArena(1, 1, 2, nil)
	if a, _ := s.Alloc(2); a != nil {
		t.Errorf("expected alloc larger than slab size to fail")
	}
}

func TestEmptyChunk(t *testing.T) {
	s := NewArena(1, 1, 2, nil)
	sc := s.slabClasses[0]
	if sc.chunk(emptyChunkLoc) != nil {
		t.Errorf("expected empty chunk to not have a chunk()")
	}
	sc1, c1 := s.chunk(emptyChunkLoc)
	if sc1 != nil || c1 != nil {
		t.Errorf("expected empty chunk to not have a chunk()")
	}
}

func TestEmptyChunkMem(t *testing.T) {
	s := NewArena(1, 1, 2, nil)
	sc := s.slabClasses[0]
	if sc.chunkMem(nil) != nil {
		t.Errorf("expected nil chunk to not have a chunk()")
	}
	if sc.chunkMem(&chunk{self: emptyChunkLoc}) != nil {
		t.Errorf("expected empty chunk to not have a chunk()")
	}
	if s.chunkMem(nil) != nil {
		t.Errorf("expected nil chunk to not have a chunk()")
	}
	if s.chunkMem(&chunk{self: emptyChunkLoc}) != nil {
		t.Errorf("expected empty chunk to not have a chunk()")
	}
}

func TestAddRefOnAlreadyReleasedBuf(t *testing.T) {
	s := NewArena(1, 1, 2, nil)
	a, _ := s.Alloc(1)
	s.DecRef(a)
	var err interface{}
	func() {
		defer func() { err = recover() }()
		s.AddRef(a)
	}()
	if err == nil {
		t.Errorf("expected panic on AddRef on already release buf")
	}
}

func TestDecRefOnAlreadyReleasedBuf(t *testing.T) {
	s := NewArena(1, 1, 2, nil)
	a, _ := s.Alloc(1)
	s.DecRef(a)
	var err interface{}
	func() {
		defer func() { err = recover() }()
		s.DecRef(a)
	}()
	if err == nil {
		t.Errorf("expected panic on DecRef on already release buf")
	}
}

func TestPushFreeChunkOnReferencedChunk(t *testing.T) {
	s := NewArena(1, 1, 2, nil)
	sc := s.slabClasses[0]
	var err interface{}
	func() {
		defer func() { err = recover() }()
		sc.pushFreeChunk(&chunk{refs: 1})
	}()
	if err == nil {
		t.Errorf("expected panic when free'ing a ref-counted chunk")
	}
}

func TestPopFreeChunkOnFreeChunk(t *testing.T) {
	s := NewArena(1, 1, 2, nil)
	sc := s.slabClasses[0]
	sc.chunkFree = emptyChunkLoc
	var err interface{}
	func() {
		defer func() { err = recover() }()
		sc.popFreeChunk()
	}()
	if err == nil {
		t.Errorf("expected panic when popFreeChunk() on free chunk")
	}
}

func TestPopFreeChunkOnReferencedFreeChunk(t *testing.T) {
	s := NewArena(1, 1024, 2, nil)
	s.Alloc(1)
	sc := s.slabClasses[0]
	sc.chunk(sc.chunkFree).refs = 1
	var err interface{}
	func() {
		defer func() { err = recover() }()
		sc.popFreeChunk()
	}()
	if err == nil {
		t.Errorf("expected panic when popFreeChunk() on ref'ed chunk")
	}
}

func TestOwns(t *testing.T) {
	s := NewArena(1, 1024, 2, nil)
	if s.Owns(nil) {
		t.Errorf("expected false when Owns on nil buf")
	}
	if s.Owns(make([]byte, 1)) {
		t.Errorf("expected false when Owns on small buf")
	}
	if s.Owns(make([]byte, 1+SLAB_MEMORY_FOOTER_LEN)) {
		t.Errorf("expected false whens Owns on non-magic buf")
	}
	a, _ := s.Alloc(1)
	if !s.Owns(a) {
		t.Errorf("expected Owns on Alloc'ed buf")
	}
}

func TestAddDecRefOnUnowned(t *testing.T) {
	s := NewArena(1, 1024, 2, nil)
	var err interface{}
	func() {
		defer func() { err = recover() }()
		s.AddRef(make([]byte, 1000))
	}()
	if err == nil {
		t.Errorf("expected panic when AddRef() on unowned buf")
	}
	err = nil
	func() {
		defer func() { err = recover() }()
		s.DecRef(make([]byte, 1000))
	}()
	if err == nil {
		t.Errorf("expected panic when DecRef() on unowned buf")
	}
}

func TestArenaChunk(t *testing.T) {
	s := NewArena(1, 100, 2, nil)
	s.Alloc(1)
	sc := &(s.slabClasses[0])
	c := sc.popFreeChunk()
	if sc.chunk(c.self) != c {
		t.Errorf("expected chunk to be the same")
	}
	sc1, c1 := s.chunk(c.self)
	if sc1 != sc || c1 != c {
		t.Errorf("expected chunk to be the same")
	}
}

func TestArenaChunkMem(t *testing.T) {
	s := NewArena(1, 100, 2, nil)
	s.Alloc(1)
	sc := s.slabClasses[0]
	c := sc.popFreeChunk()
	if sc.chunkMem(c) == nil {
		t.Errorf("expected chunkMem to be non-nil")
	}
	if s.chunkMem(c) == nil {
		t.Errorf("expected chunkMem to be non-nil")
	}
}

func TestMalloc(t *testing.T) {
	mallocWorks := true
	mallocCalls := 0
	malloc := func(sizeNeeded int) []byte {
		mallocCalls++
		if mallocWorks {
			return make([]byte, sizeNeeded)
		}
		return nil
	}
	mustNil := func(aaa []byte) {
		if aaa != nil {
			t.Errorf("expected array to be nil")
		}
	}
	notNil := func(aaa []byte) {
		if aaa == nil {
			t.Errorf("expected array to be not nil")
		}
	}
	s := NewArena(1, 4, 2, malloc)
	if mallocCalls != 0 {
		t.Errorf("expect no mallocs yet")
	}
	a, _ := s.Alloc(1)
	notNil(a)
	if mallocCalls != 1 {
		t.Errorf("expect 1 malloc")
	}
	a, _ = s.Alloc(1)
	notNil(a)
	if mallocCalls != 1 {
		t.Errorf("expect 1 malloc still, since we don't need another slab yet")
	}
	a, _ = s.Alloc(2)
	notNil(a)
	if mallocCalls != 2 {
		t.Errorf("expect 2 mallocs, since we need another slab")
	}
	a, _ = s.Alloc(1)
	notNil(a)
	if mallocCalls != 2 {
		t.Errorf("expect 2 malloc still, since we don't need another slab yet")
	}
	a, _ = s.Alloc(1)
	notNil(a)
	if mallocCalls != 2 {
		t.Errorf("expect 2 malloc still, since we don't need another slab yet")
	}
	a, _ = s.Alloc(1)
	notNil(a)
	if mallocCalls != 3 {
		t.Errorf("expect 3 mallocs, since we need another slab")
	}
	mallocWorks = false // Now we pretend to run out of memory.
	a, _ = s.Alloc(2)
	notNil(a)
	if mallocCalls != 3 {
		t.Errorf("expect 3 mallocs, since don't need another slab yet")
	}
	a, _ = s.Alloc(2)
	mustNil(a)
	if mallocCalls != 4 {
		t.Errorf("expect 4 mallocs, since needed another slab")
	}
	a, _ = s.Alloc(3)
	mustNil(a)
	if mallocCalls != 5 {
		t.Errorf("expect 5 mallocs, since needed another slab")
	}
}

func TestChaining(t *testing.T) {
	testChaining(t, NewArena(1, 1, 2, nil))
	testChaining(t, NewArena(1, 100, 2, nil))
}

func testChaining(t *testing.T, s *Arena) {
	a, _ := s.Alloc(1)
	f, _ := s.Alloc(1)
	s.DecRef(f) // The f buf is now freed.
	if s.GetNext(a) != nil {
		t.Errorf("expected nil GetNext()")
	}
	s.SetNext(a, nil)
	if s.GetNext(a) != nil {
		t.Errorf("expected nil GetNext()")
	}
	var err interface{}
	func() {
		defer func() { err = recover() }()
		s.GetNext(nil)
	}()
	if err == nil {
		t.Errorf("expected panic when GetNext(nil)")
	}
	err = nil
	func() {
		defer func() { err = recover() }()
		s.GetNext(make([]byte, 1))
	}()
	if err == nil {
		t.Errorf("expected panic when GetNext(non-arena-buf)")
	}
	err = nil
	func() {
		defer func() { err = recover() }()
		s.GetNext(f)
	}()
	if err == nil {
		t.Errorf("expected panic when GetNext(already-freed-buf)")
	}
	err = nil
	func() {
		defer func() { err = recover() }()
		s.SetNext(nil, make([]byte, 1))
	}()
	if err == nil {
		t.Errorf("expected panic when SetNext(nil)")
	}
	err = nil
	func() {
		defer func() { err = recover() }()
		s.SetNext(a, make([]byte, 1))
	}()
	if err == nil {
		t.Errorf("expected panic when SetNext(non-arena-buf)")
	}
	err = nil
	func() {
		defer func() { err = recover() }()
		s.SetNext(f, nil)
	}()
	if err == nil {
		t.Errorf("expected panic when SetNext(already-freed-buf)")
	}
	b0, _ := s.Alloc(1)
	b1, _ := s.Alloc(1)
	b1[0] = 201
	s.SetNext(b0, b1)
	bx := s.GetNext(b0)
	if bx[0] != 201 {
		t.Errorf("expected chain to work")
	}
	s.DecRef(bx)
	s.DecRef(b1)
	bx = s.GetNext(b0)
	if bx[0] != 201 {
		t.Errorf("expected chain to still work")
	}
	s.DecRef(bx)
	_, b0chunk := s.bufContainer(b0)
	if b0chunk.refs != 1 {
		t.Errorf("expected b0chunk to still be alive")
	}
	_, b1chunk := s.bufContainer(b1)
	if b1chunk == nil {
		t.Errorf("expected b1chunk to still be alive")
	}
	if b1chunk.refs != 1 {
		t.Errorf("expected b1chunk to still be ref'ed")
	}
	if b0chunk.next.isEmpty() {
		t.Errorf("expected b0chunk to not be empty")
	}
	if !b1chunk.next.isEmpty() {
		t.Errorf("expected b1chunk to have no next")
	}
	s.DecRef(b0)
	if b0chunk.refs != 0 {
		t.Errorf("expected b0chunk to not be ref'ed")
	}
	if b1chunk.refs != 0 {
		t.Errorf("expected b1chunk to not be ref'ed")
	}
	alice, _ := s.Alloc(1)
	bob, _ := s.Alloc(1)
	betty, _ := s.Alloc(1)
	_, bobChunk := s.bufContainer(bob)
	_, bettyChunk := s.bufContainer(betty)
	s.SetNext(alice, bob)
	if bobChunk.refs != 2 {
		t.Errorf("expected bob to have 2 refs")
	}
	if bettyChunk.refs != 1 {
		t.Errorf("expected betty to have 1 ref")
	}
	s.DecRef(bob)
	if bobChunk.refs != 1 {
		t.Errorf("expected bob to have 1 ref (from alice)")
	}
	if bettyChunk.refs != 1 {
		t.Errorf("expected betty to have 1 ref")
	}
	s.SetNext(alice, betty)
	if bobChunk.refs != 0 {
		t.Errorf("expected bob to have 0 ref's (alice dropped bob for betty)")
	}
	if bettyChunk.refs != 2 {
		t.Errorf("expected betty to have 2 ref (1 from alice)")
	}
	s.DecRef(betty)
	if bobChunk.refs != 0 {
		t.Errorf("expected bob to have 0 ref's (alice dropped bob for betty)")
	}
	if bettyChunk.refs != 1 {
		t.Errorf("expected betty to have 1 ref (from alice)")
	}
	s.DecRef(alice)
	if bobChunk.refs != 0 {
		t.Errorf("expected bob to have 0 ref's (alice dropped bob for betty)")
	}
	if bettyChunk.refs != 0 {
		t.Errorf("expected betty to have 0 ref (alice dropped betty)")
	}
}

func TestFindSlabClassIndex(t *testing.T) {
	s := NewArena(1, 1024, 2, nil)
	test := func(bufSize, idxExp int) {
		idxAct := s.findSlabClassIndex(bufSize)
		if idxExp != idxAct {
			t.Errorf("expected slab class index: %v, got: %v, bufSize: %v",
				idxExp, idxAct, bufSize)
		}
	}
	test(0, 0)
	test(1, 0)
	test(2, 1)
	test(3, 2)
	test(4, 2)
	test(5, 3)
	test(256, 8)
}

func TestGrowthFactors(t *testing.T) {
	for gf := 1.1; gf < 16.7; gf = gf + 0.1 {
		s := NewArena(1, 1024, gf, nil)
		a, _ := s.Alloc(1024)
		a[0] = 123
		s.DecRef(a)
		b, _ := s.Alloc(1024)
		if b[0] != 123 {
			t.Errorf("expected re-used alloc mem")
		}
	}
}

func BenchmarkReffing(b *testing.B) {
	a := NewArena(1, 1024, 2, nil)

	data, _ := a.Alloc(1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.AddRef(data)
		a.DecRef(data)
	}
}

func BenchmarkAllocingSize1(b *testing.B) {
	benchmarkAllocingConstant(b, NewArena(1, 1024, 2, nil), 1)
}

func BenchmarkAllocingSize128(b *testing.B) {
	benchmarkAllocingConstant(b, NewArena(1, 1024, 2, nil), 128)
}

func BenchmarkAllocingSize256(b *testing.B) {
	benchmarkAllocingConstant(b, NewArena(1, 1024, 2, nil), 256)
}

func benchmarkAllocingConstant(b *testing.B, a *Arena, allocSize int) {
	stuff := [][]byte{}
	for i := 0; i < 1024; i++ {
		b, _ := a.Alloc(allocSize)
		stuff = append(stuff, b)
	}
	for _, x := range stuff {
		a.DecRef(x)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b, _ := a.Alloc(allocSize)
		a.DecRef(b)
	}
}

func BenchmarkAllocingModSizes(b *testing.B) {
	benchmarkAllocingFunc(b, NewArena(1, 1024, 2, nil),
		func(i int) int { return i % 1024 })
}

func BenchmarkAllocingModSizesGrowthFactor1Dot1(b *testing.B) {
	benchmarkAllocingFunc(b, NewArena(1, 1024, 1.1, nil),
		func(i int) int { return i % 1024 })
}

func benchmarkAllocingFunc(b *testing.B, a *Arena,
	allocSize func(i int) int) {
	stuff := [][]byte{}
	for i := 0; i < 1024; i++ {
		b, _ := a.Alloc(allocSize(i))
		stuff = append(stuff, b)
	}
	for _, x := range stuff {
		a.DecRef(x)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b, _ := a.Alloc(allocSize(i))
		a.DecRef(b)
	}
}

func TestChainingSizes(t *testing.T) {
	testChainingSizes(t, NewArena(1, 100, 2, nil))
	testChainingSizes(t, NewArena(1, 200, 8, nil))
}

func testChainingSizes(t *testing.T, s *Arena) {
	curr, _ := s.Alloc(91)
	for i := 90; i > 0; i-- {
		b, _ := s.Alloc(100)
		next := b[0:i]
		s.SetNext(next, curr)
		curr = next
	}
	i := 1
	for x := curr; x != nil; x = s.GetNext(x) {
		if len(x) != i {
			t.Fatalf("expected len(x): %d, got: %d", i, len(x))
		}
		i++
	}

	b, _ := s.Alloc(5)
	if s.GetNext(b) != nil {
		t.Fatalf("expected nil")
	}
}

func TestStats(t *testing.T) {
	a := NewArena(1, 1024*1024, 2, nil)
	a.Alloc(3)
	b, _ := a.Alloc(17)
	a.DecRef(b)
	a.Alloc(4096)
	stats := a.Stats(map[string]int64{})
	if len(stats) == 0 {
		t.Errorf("expected some stats")
	}
	if stats["numAllocs"] != 3 || stats["numDecRefs"] != 1 {
		t.Errorf("expected stats did not match")
	}
	if stats["slabClass-000002-numChunksInUse"] != 1 ||
		stats["slabClass-000005-numChunksInUse"] != 0 {
		t.Errorf("expected stats did not match InUse")
	}
	if stats["slabClass-000012-chunkSize"] != 4096 ||
		stats["slabClass-000012-numChunks"] != 256 ||
		stats["slabClass-000012-numChunksFree"] != 255 ||
		stats["slabClass-000012-numChunksInUse"] != 1 ||
		stats["slabClass-000012-numSlabs"] != 1 {
		t.Errorf("expected stats did not match slabClass 12")
	}

	mk := []string{}
	for k, _ := range stats {
		mk = append(mk, k)
	}
	sort.Strings(mk)
	for _, k := range mk {
		t.Logf("%s = %d", k, stats[k])
	}
}
