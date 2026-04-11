package identity

import (
	"context"
	"testing"
	"time"

	"github.com/ofa/center/internal/models"
)

// TestIdentityStorePerformance 测试身份存储性能
func TestIdentityStorePerformance(t *testing.T) {
	store := NewMemoryStore()

	ctx := context.Background()

	// 创建身份性能测试
	t.Run("CreatePerformance", func(t *testing.T) {
		start := time.Now()
		for i := 0; i < 1000; i++ {
			identity := &models.PersonalIdentity{
				ID:   fmt.Sprintf("perf_identity_%d", i),
				Name: fmt.Sprintf("用户%d", i),
				Personality: models.Personality{
					Openness:          0.5 + float64(i%50)/100,
					Conscientiousness: 0.5,
					Extraversion:      0.5,
					Agreeableness:     0.5,
					Neuroticism:       0.5,
				},
			}
			err := store.Create(ctx, identity)
			if err != nil {
				t.Errorf("创建身份失败: %v", err)
			}
		}
		duration := time.Since(start)

		opsPerSec := float64(1000) / duration.Seconds()
		t.Logf("Create性能: %d 操作耗时 %v, %.2f ops/sec", 1000, duration, opsPerSec)

		if opsPerSec < 1000 {
			t.Logf("警告: Create性能低于1000 ops/sec")
		}
	})

	// 获取身份性能测试
	t.Run("GetPerformance", func(t *testing.T) {
		start := time.Now()
		for i := 0; i < 1000; i++ {
			id := fmt.Sprintf("perf_identity_%d", i)
			identity, err := store.Get(ctx, id)
			if err != nil {
				t.Errorf("获取身份失败: %v", err)
			}
			if identity == nil {
				t.Errorf("身份 %s 不存在", id)
			}
		}
		duration := time.Since(start)

		opsPerSec := float64(1000) / duration.Seconds()
		t.Logf("Get性能: %d 操作耗时 %v, %.2f ops/sec", 1000, duration, opsPerSec)

		if opsPerSec < 5000 {
			t.Logf("警告: Get性能低于5000 ops/sec")
		}
	})

	// 更新身份性能测试
	t.Run("UpdatePerformance", func(t *testing.T) {
		start := time.Now()
		for i := 0; i < 1000; i++ {
			id := fmt.Sprintf("perf_identity_%d", i)
			identity, _ := store.Get(ctx, id)
			identity.Personality.Openness += 0.01
			err := store.Update(ctx, identity)
			if err != nil {
				t.Errorf("更新身份失败: %v", err)
			}
		}
		duration := time.Since(start)

		opsPerSec := float64(1000) / duration.Seconds()
		t.Logf("Update性能: %d 操作耗时 %v, %.2f ops/sec", 1000, duration, opsPerSec)

		if opsPerSec < 1000 {
			t.Logf("警告: Update性能低于1000 ops/sec")
		}
	})
}

// BenchmarkIdentityCreate 身份创建基准测试
func BenchmarkIdentityCreate(b *testing.B) {
	store := NewMemoryStore()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		identity := &models.PersonalIdentity{
			ID:   fmt.Sprintf("bench_identity_%d", i),
			Name: fmt.Sprintf("用户%d", i),
			Personality: models.Personality{
				Openness:          0.5,
				Conscientiousness: 0.5,
				Extraversion:      0.5,
				Agreeableness:     0.5,
				Neuroticism:       0.5,
			},
		}
		store.Create(ctx, identity)
	}
}

// BenchmarkIdentityGet 身份获取基准测试
func BenchmarkIdentityGet(b *testing.B) {
	store := NewMemoryStore()
	ctx := context.Background()

	// 预填充
	for i := 0; i < b.N; i++ {
		identity := &models.PersonalIdentity{
			ID:   fmt.Sprintf("bench_identity_%d", i),
			Name: fmt.Sprintf("用户%d", i),
		}
		store.Create(ctx, identity)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Get(ctx, fmt.Sprintf("bench_identity_%d", i))
	}
}

// BenchmarkIdentityUpdate 身份更新基准测试
func BenchmarkIdentityUpdate(b *testing.B) {
	store := NewMemoryStore()
	ctx := context.Background()

	// 预填充
	identities := make([]*models.PersonalIdentity, b.N)
	for i := 0; i < b.N; i++ {
		identities[i] = &models.PersonalIdentity{
			ID:   fmt.Sprintf("bench_identity_%d", i),
			Name: fmt.Sprintf("用户%d", i),
		}
		store.Create(ctx, identities[i])
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		identities[i].Personality.Openness += 0.01
		store.Update(ctx, identities[i])
	}
}

// BenchmarkIdentityConcurrent 身份并发操作基准测试
func BenchmarkIdentityConcurrent(b *testing.B) {
	store := NewMemoryStore()
	ctx := context.Background()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			id := fmt.Sprintf("parallel_identity_%d", i%1000)
			switch i % 3 {
			case 0:
				identity := &models.PersonalIdentity{
					ID:   id,
					Name: fmt.Sprintf("用户%d", i),
				}
				store.Create(ctx, identity)
			case 1:
				store.Get(ctx, id)
			case 2:
				identity, _ := store.Get(ctx, id)
				if identity != nil {
					identity.Personality.Openness += 0.01
					store.Update(ctx, identity)
				}
			}
			i++
		}
	})
}

// TestConcurrentIdentitySync 测试并发身份同步
func TestConcurrentIdentitySync(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	var wg sync.WaitGroup
	start := time.Now()

	// 10个并发协程创建身份
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				identity := &models.PersonalIdentity{
					ID:   fmt.Sprintf("concurrent_%d_%d", id, j),
					Name: fmt.Sprintf("用户%d-%d", id, j),
				}
				store.Create(ctx, identity)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	totalOps := 10 * 100
	opsPerSec := float64(totalOps) / duration.Seconds()
	t.Logf("并发创建性能: %d 操作耗时 %v, %.2f ops/sec", totalOps, duration, opsPerSec)
}