package cache_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	redismock "github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/require"
	cache "laschool.ru/event-booking-service/internal/cache"
)

func TestSetGetDelete(t *testing.T) {
	client, mock := redismock.NewClientMock()
	defer client.Close()
	s := cache.NewService(client, 100*time.Millisecond)
	ctx := context.Background()
	value := map[string]string{"foo": "bar"}
	b, err := json.Marshal(value)
	require.NoError(t, err)

	mock.ExpectSet("key1", b, time.Minute).SetVal("OK")
	require.NoError(t, s.Set(ctx, "key1", value, time.Minute))

	mock.ExpectGet("key1").SetVal(string(b))
	var got map[string]string
	found, err := s.Get(ctx, "key1", &got)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, value, got)

	mock.ExpectDel("key1").SetVal(1)
	require.NoError(t, s.Delete(ctx, "key1"))

	mock.ExpectGet("key1").RedisNil()
	var got2 map[string]string
	found2, err := s.Get(ctx, "key1", &got2)
	require.NoError(t, err)
	require.False(t, found2)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetProtected_CacheNo_Calculates(t *testing.T) {
	lockTTL := time.Second
	client, mock := redismock.NewClientMock()
	defer client.Close()
	s := cache.NewService(client, time.Second)
	ctx := context.Background()

	key := "some:key"
	value := map[string]string{"x": "y"}
	b, err := json.Marshal(value)
	require.NoError(t, err)

	mock.ExpectGet(key).RedisNil()

	mock.ExpectSetNX(key+":lock", "1", lockTTL).SetVal(true)

	mock.ExpectDel(key + ":lock").SetVal(1)

	res, err := s.GetProtected(ctx, key, func() (interface{}, error) { return value, nil }, 2*time.Second)
	require.NoError(t, err)
	require.Equal(t, b, res)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeletePattern(t *testing.T) {
	client, mock := redismock.NewClientMock()
	defer client.Close()
	s := cache.NewService(client, 100*time.Millisecond)
	ctx := context.Background()

	pattern := "test:pattern:*"
	keys := []string{"test:pattern:1", "test:pattern:2"}

	// Scan returns keys and cursor 0
	mock.ExpectScan(0, pattern, 100).SetVal(keys, 0)
	mock.ExpectDel(keys...).SetVal(int64(len(keys)))

	err := s.DeletePattern(ctx, pattern)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeletePattern_MultiBatch(t *testing.T) {
	client, mock := redismock.NewClientMock()
	defer client.Close()
	s := cache.NewService(client, 100*time.Millisecond)
	ctx := context.Background()

	pattern := "test:pattern:*"
	batch1 := []string{"test:pattern:1", "test:pattern:2"}
	batch2 := []string{"test:pattern:3"}

	mock.ExpectScan(0, pattern, 100).SetVal(batch1, 1)
	mock.ExpectScan(1, pattern, 100).SetVal(batch2, 0)

	all := append(batch1, batch2...)
	mock.ExpectDel(all...).SetVal(int64(len(all)))

	err := s.DeletePattern(ctx, pattern)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestWithJitter(t *testing.T) {
	client, _ := redismock.NewClientMock()
	defer client.Close()
	s := cache.NewService(client, time.Second)

	base := 1 * time.Second
	jittered := s.WithJitter(base)
	require.GreaterOrEqual(t, jittered, time.Second,
		"WithJitter should guarantee minimum 1 second")
	require.LessOrEqual(t, jittered, base+base/10,
		"WithJitter should not exceed base + 10%")
}
