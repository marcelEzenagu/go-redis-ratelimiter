package limiter


import (
   "context"
   "time"


   "github.com/go-redis/redis/v8"
)


type RateLimiter struct {
   client *redis.Client
}


func NewLimiter(c *redis.Client) *RateLimiter {
   return &RateLimiter{
       client: c,
   }
}


func (rt *RateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, int64, error) {


   pipeline := rt.client.TxPipeline()
   count := pipeline.Incr(ctx, key)
   pipeline.Expire(ctx, key, window)


   _, err := pipeline.Exec(ctx)
   if err != nil {
       return false, 0, err
   }


   remain := int64(limit) - count.Val()
   if remain <= 0 {
       remain = 0
   }


   return count.Val() <= int64(limit), remain, nil


}


func (rt *RateLimiter) GetCount(ctx context.Context, ip string) (int64, error) {
   return rt.client.Get(ctx, ip).Int64()
}








