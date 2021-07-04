package service

import (
	"github.com/go-redis/redis"
	"strconv"
	"time"
)

/**
基于Redis频率控制的实现
	count（计数型）：
		count 类型直接利用 redis 的 setex() 来计数和设置时间。
		也就是，当第一次请求过来的时候，创建一个过期时间为 time 的 key ，并设置为 1 ，
		在接下去的时间内，每次请求过来通过的话 key 就加 1 ，
		当达到 count的时候，就会禁止访问，直到 key 过期
	缺点：
		这种方法实现起来比较简单，但是这可能出现短时间流量暴增问题。
		比如，某个接口限制是 5 秒只能访问 3 次，在前 1 秒，只访问了一次，在 5 秒快过期的时候，
		突然访问了 2 次，在 6 秒的时候又访问了 3 次，
		相当于在 2 秒内访问了 5 次，流量短时间跟预期的比翻了一倍，
		所以后面添加了漏桶型来解决这个问题。

	leak（漏桶型）：
		漏桶模式是基于漏桶算法，能够平滑网络上的流量，简单的讲就是在过去 time 秒内，
		访问次数不能超过 count 次，解决 count 流量倍增问题。
		漏桶模式可以利用 redis 的 list 数据结构或 zset 数据结构来实现。
		采用List的数据结构：
		1、存储设计和判别条件
			采用 redis 的 list 数据结构，实现一种先进先出的队列。
			队列的每个元素，存储一个时间戳，记录一次访问的时间。
			漏桶大小为 count。
			如果第 count 个元素的时间戳，距离当前时间，小于等于 time ，则说明漏桶有“溢出”。
		2、过期元素清除
			因为 redis 不能设置 list 里元素过期时间，所以需要手动删除，有两种方法：
			a、可以在每次访问后清除队尾多余元素。
			b、可以利用 go 协程进行异步处理，不影响速度。
			可能出现一个key访问一段时间后突然不访问，导致内存浪费，还需要设置大于 time 的过期时间。
		采用Zset的数据结构：
		1、存储设计和判别条件
			采用 redis 的 zset 数据结构，实现一种时间戳有序集合。
			集合的每个元素， member 和 score 都为时间戳(纳秒级别)。
			漏桶大小为 count。
			如果在（当前时间戳 - time）的时间戳内元素的个数超过 count 则说明漏桶有“溢出”。
		2、过期元素清除
			和上面 list 数据结构基本类似，不同的是每次清理是清理 score 小于当前时间戳 - time的时间戳。

*/

type Rule struct {
	rds   *redis.Client
	ttl   int64 // 时间
	count int64 // 次数
}

// 访问控制函数
func (this *Rule) CountBrowse(key string) (bool, error) {
	cnt, err := this.rds.Get(key).Int64()
	if err == redis.Nil {
		return false, nil
	}
	if this.count == 0 || cnt < this.count {
		return false, nil
	}
	return true, nil
}

// 更新
func (this *Rule) CountUpdate(key string) (err error) {
	exp, err := this.rds.PTTL(key).Result()
	if err != nil {
		return
	}
	if exp <= 0 {
		err = this.rds.Set(key, 1, time.Second*time.Duration(this.ttl)).Err()
		return
	}

	err = this.rds.Incr(key).Err()
	return

}

func (this *Rule) LeakListBrowse(key string) (bool, error) {
	lstlen, err := this.rds.LLen(key).Result()
	if err != nil {
		return false, err
	}
	if lstlen == 0 || lstlen < this.count {
		return false, err
	}
	// 利用go协程，删除list队尾超过count长度外的数据
	defer func() {
		go func() {
			for lstlen > this.count+1 {
				this.rds.RPop(key).Result()
				lstlen--
			}
		}()
	}()

	now := time.Now().Unix()
	cts, err := this.rds.LIndex(key, this.count-1).Int64()
	if err != nil {
		return false, err
	}
	if now-cts < this.ttl {
		return true, nil
	}
	return false, nil

}

func (this *Rule) LeakListUpdate(key string) (err error) {
	now := time.Now().Unix()
	// 队列首部入当前时间
	_, err = this.rds.LPush(key, now).Result()
	if err != nil {
		return
	}

	// 设置key超时时间（防止过久没访问占用内存）
	_, err = this.rds.Expire(key, 2*time.Duration(this.ttl)*time.Second).Result()
	return
}

func (this *Rule) LeakZsetBrowse(key string) (bool, error) {
	now := time.Now().UnixNano()
	//转为纳秒级别，防止zset覆盖
	from := now - this.ttl*1e9 + 1e9
	defer func() {
		go func() {
			this.rds.ZRemRangeByScore(key, "0", strconv.FormatInt(from, 10)).Result()
		}()
	}()

	zsetlen, err := this.rds.ZCount(key, strconv.FormatInt(from, 10), strconv.FormatInt(now, 10)).Result()
	if err != nil {
		return false, err
	}
	if zsetlen >= this.count {
		return true, nil
	}
	return false, nil
}

func (this *Rule) LeakZsetUpdate(key string) (err error) {
	now := time.Now().UnixNano()
	z := redis.Z{float64(now), now}
	if _, err = this.rds.ZAdd(key, z).Result(); err != nil {
		return
	}
	_, err = this.rds.Expire(key, 2*time.Duration(this.ttl)*time.Second).Result()
	return
}
