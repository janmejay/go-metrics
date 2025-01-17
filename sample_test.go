package metrics

import (
	"math/rand"
	"runtime"
	"testing"
	"time"
)

// Benchmark{Compute,Copy}{1000,1000000} demonstrate that, even for relatively
// expensive computations like Variance, the cost of copying the Sample, as
// approximated by a make and copy, is much greater than the cost of the
// computation for small samples and only slightly less for large samples.
func BenchmarkCompute1000(b *testing.B) {
	s := make([]int64, 1000)
	for i := 0; i < len(s); i++ {
		s[i] = int64(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SampleVariance(s)
	}
}
func BenchmarkCompute1000000(b *testing.B) {
	s := make([]int64, 1000000)
	for i := 0; i < len(s); i++ {
		s[i] = int64(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SampleVariance(s)
	}
}
func BenchmarkCopy1000(b *testing.B) {
	s := make([]int64, 1000)
	for i := 0; i < len(s); i++ {
		s[i] = int64(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sCopy := make([]int64, len(s))
		copy(sCopy, s)
	}
}
func BenchmarkCopy1000000(b *testing.B) {
	s := make([]int64, 1000000)
	for i := 0; i < len(s); i++ {
		s[i] = int64(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sCopy := make([]int64, len(s))
		copy(sCopy, s)
	}
}

func BenchmarkExpDecaySample257(b *testing.B) {
	benchmarkSample(b, NewExpDecaySample(257, 0.015))
}

func BenchmarkExpDecaySample514(b *testing.B) {
	benchmarkSample(b, NewExpDecaySample(514, 0.015))
}

func BenchmarkExpDecaySample1028(b *testing.B) {
	benchmarkSample(b, NewExpDecaySample(1028, 0.015))
}

func BenchmarkAutoSizedExpDecaySample257(b *testing.B) {
	benchmarkSample(b, NewAutoSizedExpDecaySample(257, 0.015))
}

func BenchmarkAutoSizedExpDecaySample514(b *testing.B) {
	benchmarkSample(b, NewAutoSizedExpDecaySample(514, 0.015))
}

func BenchmarkAutoSizedExpDecaySample1028(b *testing.B) {
	benchmarkSample(b, NewAutoSizedExpDecaySample(1028, 0.015))
}

func BenchmarkUniformSample257(b *testing.B) {
	benchmarkSample(b, NewUniformSample(257))
}

func BenchmarkUniformSample514(b *testing.B) {
	benchmarkSample(b, NewUniformSample(514))
}

func BenchmarkUniformSample1028(b *testing.B) {
	benchmarkSample(b, NewUniformSample(1028))
}

func TestExpDecaySample10(t *testing.T) {
	rand.Seed(1)
	fixedSizeSample := NewExpDecaySample(100, 0.99)
	autoSizedSample := NewAutoSizedExpDecaySample(100, 0.99)
	for _, s := range []Sample{fixedSizeSample, autoSizedSample} {
		for i := 0; i < 10; i++ {
			s.Update(int64(i))
		}
		if size := s.Count(); 10 != size {
			t.Errorf("s.Count(): 10 != %v\n", size)
		}
		if size := s.Size(); 10 != size {
			t.Errorf("s.Size(): 10 != %v\n", size)
		}
		if l := len(s.Values()); 10 != l {
			t.Errorf("len(s.Values()): 10 != %v\n", l)
		}
		for _, v := range s.Values() {
			if v > 10 || v < 0 {
				t.Errorf("out of range [0, 10): %v\n", v)
			}
		}
	}
}

func TestExpDecaySample100(t *testing.T) {
	rand.Seed(1)
	fixedSizeSample := NewExpDecaySample(1000, 0.01)
	autoSizedSample := NewAutoSizedExpDecaySample(1000, 0.01)
	for _, s := range []Sample{fixedSizeSample, autoSizedSample} {
		for i := 0; i < 100; i++ {
			s.Update(int64(i))
		}
		if size := s.Count(); 100 != size {
			t.Errorf("s.Count(): 100 != %v\n", size)
		}
		if size := s.Size(); 100 != size {
			t.Errorf("s.Size(): 100 != %v\n", size)
		}
		if l := len(s.Values()); 100 != l {
			t.Errorf("len(s.Values()): 100 != %v\n", l)
		}
		for _, v := range s.Values() {
			if v > 100 || v < 0 {
				t.Errorf("out of range [0, 100): %v\n", v)
			}
		}
	}
}

func TestExpDecaySample1000(t *testing.T) {
	rand.Seed(1)
	fixedSizeSample := NewExpDecaySample(100, 0.99)
	autoSizedSample := NewAutoSizedExpDecaySample(100, 0.99)
	for _, s := range []Sample{fixedSizeSample, autoSizedSample} {
		for i := 0; i < 1000; i++ {
			s.Update(int64(i))
		}
		if size := s.Count(); 1000 != size {
			t.Errorf("s.Count(): 1000 != %v\n", size)
		}
		if size := s.Size(); 100 != size {
			t.Errorf("s.Size(): 100 != %v\n", size)
		}
		if l := len(s.Values()); 100 != l {
			t.Errorf("len(s.Values()): 100 != %v\n", l)
		}
		for _, v := range s.Values() {
			if v > 1000 || v < 0 {
				t.Errorf("out of range [0, 1000): %v\n", v)
			}
		}
	}
}

// This test makes sure that the sample's priority is not amplified by using
// nanosecond duration since start rather than second duration since start.
// The priority becomes +Inf quickly after starting if this is done,
// effectively freezing the set of samples until a rescale step happens.
func TestExpDecaySampleNanosecondRegression(t *testing.T) {
	rand.Seed(1)
	fixedSizeSample := NewExpDecaySample(100, 0.99)
	autoSizedSample := NewAutoSizedExpDecaySample(100, 0.99)
	for _, s := range []Sample{fixedSizeSample, autoSizedSample} {
		for i := 0; i < 100; i++ {
			s.Update(10)
		}
		time.Sleep(1 * time.Millisecond)
		for i := 0; i < 100; i++ {
			s.Update(20)
		}
		v := s.Values()
		avg := float64(0)
		for i := 0; i < len(v); i++ {
			avg += float64(v[i])
		}
		avg /= float64(len(v))
		if avg > 16 || avg < 14 {
			t.Errorf("out of range [14, 16]: %v\n", avg)
		}
	}
}

func TestExpDecaySampleRescale(t *testing.T) {
	fixedSizeSample := NewExpDecaySample(2, 0.001)
	autoSizedSample := NewAutoSizedExpDecaySample(2, 0.001)
	for _, s := range []Sample{fixedSizeSample, autoSizedSample} {
		s := s.(*ExpDecaySample)
		s.update(time.Now(), 1)
		s.update(time.Now().Add(time.Hour+time.Microsecond), 1)
		for _, v := range s.values.Values() {
			if v.k == 0.0 {
				t.Fatal("v.k == 0.0")
			}
		}
	}
}

type autoResizeBehavior struct {
	updates      int
	samples      int
	capacity     int
	snapshotSize int
}

func testExpDecaySampleAutoResizeBehavior(
	t *testing.T,
	behavior []autoResizeBehavior,
	s *ExpDecaySample,
) {
	count := int64(0)
	for si, b := range behavior {
		for i := 0; i < b.updates; i++ {
			s.Update(1)
		}
		count += int64(b.updates)
		snapshot := s.Snapshot()
		if b.snapshotSize == 0 {
			b.snapshotSize = b.samples
		}
		if snapshot.Size() != b.snapshotSize {
			t.Errorf(
				"incorrect snapshot size: %v != %v for snap %d \n",
				snapshot.Size(),
				b.snapshotSize,
				si)
		}
		if len(s.values.Values()) != b.samples {
			t.Errorf(
				"incorrect sample retention: %v != %v for snap %d \n",
				s.values.Size(),
				b.samples,
				si)
		}
		if cap(s.values.Values()) != b.capacity {
			t.Errorf(
				"incorrect reservoir capacity: %v != %v for snap %d \n",
				cap(s.values.Values()),
				b.capacity,
				si)
		}
		if snapshot.Count() != count {
			t.Errorf(
				"incorrect snapshot count: %v != %v for snap %d \n",
				snapshot.Count(),
				count,
				si)
		}
		if s.Count() != count {
			t.Errorf(
				"incorrect count: %v != %v for snap %d \n",
				s.Count(),
				count,
				si)
		}
	}
}

func TestExpDecaySampleAutoResize(t *testing.T) {
	testExpDecaySampleAutoResizeBehavior(
		t,
		[]autoResizeBehavior {
			autoResizeBehavior{updates: 1,        samples: 1,   capacity: 50},
			autoResizeBehavior{updates: 1,        samples: 2,   capacity: 25},
			autoResizeBehavior{updates: 1,        samples: 3,   capacity: 12},
			autoResizeBehavior{updates: 1,        samples: 4,   capacity: 8},
			autoResizeBehavior{updates: 1,        samples: 5,   capacity: 8},
			autoResizeBehavior{updates: 1,        samples: 6,   capacity: 8},
			autoResizeBehavior{updates: 1,        samples: 7,   capacity: 8},
			autoResizeBehavior{updates: 1,        samples: 8,   capacity: 8},
			autoResizeBehavior{updates: 1,        samples: 8,   capacity: 8},
			autoResizeBehavior{updates: 2,        samples: 8,   capacity: 8},
			autoResizeBehavior{updates: 4,        samples: 8,   capacity: 8},
			autoResizeBehavior{updates: 8,        samples: 8,   capacity: 8},
			autoResizeBehavior{updates: 15,       samples: 8,   capacity: 8},
			autoResizeBehavior{updates: 16,       samples: 8,   capacity: 16},
			autoResizeBehavior{updates: 31,       samples: 16,  capacity: 16},
			autoResizeBehavior{updates: 32,       samples: 16,  capacity: 32},
			autoResizeBehavior{updates: 64 + 16,  samples: 32,  capacity: 64},
			autoResizeBehavior{updates: 128 + 32, samples: 64,  capacity: 100},
			autoResizeBehavior{updates: 1000,     samples: 100, capacity: 100},
			autoResizeBehavior{updates: 50,       samples: 100, capacity: 100},
			autoResizeBehavior{
				updates: 49,
				samples: 50,
				capacity: 50,
				snapshotSize: 100}},
		NewAutoSizedExpDecaySample(100, 0.01).(*ExpDecaySample))

	testExpDecaySampleAutoResizeBehavior(
		t,
		[]autoResizeBehavior {
			autoResizeBehavior{updates: 1,    samples: 1,   capacity: 100},
			autoResizeBehavior{updates: 1000, samples: 100, capacity: 100}},
		NewExpDecaySample(100, 0.01).(*ExpDecaySample))
}

func TestExpDecaySampleSnapshot(t *testing.T) {
	now := time.Now()
	fixedSizeSample := NewExpDecaySample(100, 0.99)
	autoSizedSample := NewAutoSizedExpDecaySample(100, 0.99)
	for _, s := range []Sample{fixedSizeSample, autoSizedSample} {
		rand.Seed(1)
		for i := 1; i <= 10000; i++ {
			s.(*ExpDecaySample).update(now.Add(time.Duration(i)), int64(i))
		}
		snapshot := s.Snapshot()
		s.Update(1)
		testExpDecaySampleStatistics(t, snapshot)
	}
}

func TestExpDecaySampleStatistics(t *testing.T) {
	now := time.Now()
	fixedSizeSample := NewExpDecaySample(100, 0.99)
	autoSizedSample := NewAutoSizedExpDecaySample(100, 0.99)
	for _, s := range []Sample{fixedSizeSample, autoSizedSample} {
		rand.Seed(1)
		for i := 1; i <= 10000; i++ {
			s.(*ExpDecaySample).update(now.Add(time.Duration(i)), int64(i))
		}
		testExpDecaySampleStatistics(t, s)
	}
}

func TestUniformSample(t *testing.T) {
	fixedSizeSample := NewUniformSample(100)
	autoSizedSample := NewAutoSizedUniformSample(100)
	for _, s := range []Sample{fixedSizeSample, autoSizedSample} {
		rand.Seed(1)
		for i := 0; i < 1000; i++ {
			s.Update(int64(i))
		}
		if size := s.Count(); 1000 != size {
			t.Errorf("s.Count(): 1000 != %v\n", size)
		}
		if size := s.Size(); 100 != size {
			t.Errorf("s.Size(): 100 != %v\n", size)
		}
		if l := len(s.Values()); 100 != l {
			t.Errorf("len(s.Values()): 100 != %v\n", l)
		}
		for _, v := range s.Values() {
			if v > 1000 || v < 0 {
				t.Errorf("out of range [0, 100): %v\n", v)
			}
		}
	}
}

func TestUniformSampleIncludesTail(t *testing.T) {
	fixedSizeSample := NewUniformSample(100)
	autoSizedSample := NewAutoSizedUniformSample(100)
	for _, s := range []Sample{fixedSizeSample, autoSizedSample} {
		rand.Seed(1)
		max := 100
		for i := 0; i < max; i++ {
			s.Update(int64(i))
		}
		v := s.Values()
		sum := 0
		exp := (max - 1) * max / 2
		for i := 0; i < len(v); i++ {
			sum += int(v[i])
		}
		if exp != sum {
			t.Errorf("sum: %v != %v\n", exp, sum)
		}
	}
}

func testUniformSampleAutoResizeBehavior(
	t *testing.T,
	behavior []autoResizeBehavior,
	s *UniformSample,
) {
	count := int64(0)
	for si, b := range behavior {
		for i := 0; i < b.updates; i++ {
			s.Update(1)
		}
		count += int64(b.updates)
		snapshot := s.Snapshot()
		if b.snapshotSize == 0 {
			b.snapshotSize = b.samples
		}
		if snapshot.Size() != b.snapshotSize {
			t.Errorf(
				"incorrect snapshot size: %v != %v for snap %d \n",
				snapshot.Size(),
				b.snapshotSize,
				si)
		}
		if len(s.values) != b.samples {
			t.Errorf(
				"incorrect sample retention: %v != %v for snap %d \n",
				len(s.values),
				b.samples,
				si)
		}
		if cap(s.values) != b.capacity {
			t.Errorf(
				"incorrect reservoir capacity: %v != %v for snap %d \n",
				cap(s.values),
				b.capacity,
				si)
		}
		if snapshot.Count() != count {
			t.Errorf(
				"incorrect snapshot count: %v != %v for snap %d \n",
				snapshot.Count(),
				count,
				si)
		}
		if s.Count() != count {
			t.Errorf(
				"incorrect count: %v != %v for snap %d \n",
				s.Count(),
				count,
				si)
		}
	}
}

func TestUniformSampleAutoResize(t *testing.T) {
	testUniformSampleAutoResizeBehavior(
		t,
		[]autoResizeBehavior {
			autoResizeBehavior{updates: 1,        samples: 1,   capacity: 50},
			autoResizeBehavior{updates: 1,        samples: 2,   capacity: 25},
			autoResizeBehavior{updates: 1,        samples: 3,   capacity: 12},
			autoResizeBehavior{updates: 1,        samples: 4,   capacity: 8},
			autoResizeBehavior{updates: 1,        samples: 5,   capacity: 8},
			autoResizeBehavior{updates: 1,        samples: 6,   capacity: 8},
			autoResizeBehavior{updates: 1,        samples: 7,   capacity: 8},
			autoResizeBehavior{updates: 1,        samples: 8,   capacity: 8},
			autoResizeBehavior{updates: 1,        samples: 8,   capacity: 8},
			autoResizeBehavior{updates: 2,        samples: 8,   capacity: 8},
			autoResizeBehavior{updates: 4,        samples: 8,   capacity: 8},
			autoResizeBehavior{updates: 8,        samples: 8,   capacity: 8},
			autoResizeBehavior{updates: 15,       samples: 8,   capacity: 8},
			autoResizeBehavior{updates: 16,       samples: 8,   capacity: 16},
			autoResizeBehavior{updates: 31,       samples: 16,  capacity: 16},
			autoResizeBehavior{updates: 32,       samples: 16,  capacity: 32},
			autoResizeBehavior{updates: 64 + 16,  samples: 32,  capacity: 64},
			autoResizeBehavior{updates: 128 + 32, samples: 64,  capacity: 100},
			autoResizeBehavior{updates: 1000,     samples: 100, capacity: 100},
			autoResizeBehavior{updates: 50,       samples: 100, capacity: 100},
			autoResizeBehavior{
				updates: 49,
				samples: 50,
				capacity: 50,
				snapshotSize: 100}},
		NewAutoSizedUniformSample(100).(*UniformSample))

	testUniformSampleAutoResizeBehavior(
		t,
		[]autoResizeBehavior {
			autoResizeBehavior{updates: 1,    samples: 1,   capacity: 100},
			autoResizeBehavior{updates: 1000, samples: 100, capacity: 100}},
		NewUniformSample(100).(*UniformSample))
}

func TestUniformSampleSnapshot(t *testing.T) {
	fixedSizeSample := NewUniformSample(100)
	autoSizedSample := NewAutoSizedUniformSample(100)
	for _, s := range []Sample{fixedSizeSample, autoSizedSample} {
		rand.Seed(1)
		for i := 1; i <= 10000; i++ {
			s.Update(int64(i))
		}
		snapshot := s.Snapshot()
		s.Update(1)
		testUniformSampleStatistics(t, snapshot)
	}
}

func TestUniformSampleStatistics(t *testing.T) {
	fixedSizeSample := NewUniformSample(100)
	autoSizedSample := NewAutoSizedUniformSample(100)
	for _, s := range []Sample{fixedSizeSample, autoSizedSample} {
		rand.Seed(1)
		for i := 1; i <= 10000; i++ {
			s.Update(int64(i))
		}
		testUniformSampleStatistics(t, s)
	}
}

func benchmarkSample(b *testing.B, s Sample) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	pauseTotalNs := memStats.PauseTotalNs
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Update(1)
	}
	b.StopTimer()
	runtime.GC()
	runtime.ReadMemStats(&memStats)
	b.Logf("GC cost: %d ns/op", int(memStats.PauseTotalNs-pauseTotalNs)/b.N)
}

func testExpDecaySampleStatistics(t *testing.T, s Sample) {
	if count := s.Count(); 10000 != count {
		t.Errorf("s.Count(): 10000 != %v\n", count)
	}
	if min := s.Min(); 107 != min {
		t.Errorf("s.Min(): 107 != %v\n", min)
	}
	if max := s.Max(); 10000 != max {
		t.Errorf("s.Max(): 10000 != %v\n", max)
	}
	if mean := s.Mean(); 4965.98 != mean {
		t.Errorf("s.Mean(): 4965.98 != %v\n", mean)
	}
	if stdDev := s.StdDev(); 2959.825156930727 != stdDev {
		t.Errorf("s.StdDev(): 2959.825156930727 != %v\n", stdDev)
	}
	ps := s.Percentiles([]float64{0.5, 0.75, 0.99})
	if 4615 != ps[0] {
		t.Errorf("median: 4615 != %v\n", ps[0])
	}
	if 7672 != ps[1] {
		t.Errorf("75th percentile: 7672 != %v\n", ps[1])
	}
	if 9998.99 != ps[2] {
		t.Errorf("99th percentile: 9998.99 != %v\n", ps[2])
	}
}

func testUniformSampleStatistics(t *testing.T, s Sample) {
	if count := s.Count(); 10000 != count {
		t.Errorf("s.Count(): 10000 != %v\n", count)
	}
	if min := s.Min(); 37 != min {
		t.Errorf("s.Min(): 37 != %v\n", min)
	}
	if max := s.Max(); 9989 != max {
		t.Errorf("s.Max(): 9989 != %v\n", max)
	}
	if mean := s.Mean(); 4748.14 != mean {
		t.Errorf("s.Mean(): 4748.14 != %v\n", mean)
	}
	if stdDev := s.StdDev(); 2826.684117548333 != stdDev {
		t.Errorf("s.StdDev(): 2826.684117548333 != %v\n", stdDev)
	}
	ps := s.Percentiles([]float64{0.5, 0.75, 0.99})
	if 4599 != ps[0] {
		t.Errorf("median: 4599 != %v\n", ps[0])
	}
	if 7380.5 != ps[1] {
		t.Errorf("75th percentile: 7380.5 != %v\n", ps[1])
	}
	if 9986.429999999998 != ps[2] {
		t.Errorf("99th percentile: 9986.429999999998 != %v\n", ps[2])
	}
}

// TestUniformSampleConcurrentUpdateCount would expose data race problems with
// concurrent Update and Count calls on Sample when test is called with -race
// argument
func TestUniformSampleConcurrentUpdateCount(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
	fixedSizeSample := NewUniformSample(100)
	autoSizedSample := NewAutoSizedUniformSample(100)
	for _, s := range []Sample{fixedSizeSample, autoSizedSample} {
		rand.Seed(1)
		for i := 0; i < 100; i++ {
			s.Update(int64(i))
		}
		quit := make(chan struct{})
		go func() {
			t := time.NewTicker(10 * time.Millisecond)
			for {
				select {
				case <-t.C:
					s.Update(rand.Int63())
				case <-quit:
					t.Stop()
					return
				}
			}
		}()
		for i := 0; i < 1000; i++ {
			s.Count()
			time.Sleep(5 * time.Millisecond)
		}
		quit <- struct{}{}
	}
}
