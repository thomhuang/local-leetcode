package time_based_key_value_store

/*
Pretty straight forward, from the problem statement, we know that our timestamps, as Set() is called,
is strictly increasing ...

So our values per timestamp is always in an order in which the timestamps are ordered ...
so Set() can do the easy append on the array of (value, timestamp) pairs.

Now on the Get(), we could've done an easy approach of just looping through each (value, timestamp) from the back
and find the first timestamp that is <= and return that (if it exists)

but we've been done a favor in that the timestamps are sorted already for us, so we can do a binary search to
perform this search in O(log n) versus the O(n) scan in worst case ...
*/
type pair struct {
	value     string
	timestamp int
}

type TimeMap struct {
	store map[string][]pair
}

func Constructor() TimeMap {
	return TimeMap{
		store: make(map[string][]pair),
	}

}

func (this *TimeMap) Set(key string, value string, timestamp int) {
	this.store[key] = append(this.store[key], pair{value: value, timestamp: timestamp})
}

func (this *TimeMap) Get(key string, timestamp int) string {
	curr, exists := this.store[key]

	if !exists {
		return ""
	}

	var res string
	l, r := 0, len(curr)-1
	for l <= r {
		m := (l + r) / 2
		// if the midpoint is valid, then move our left pointer up, and keep our result
		// in case the left window moves too far for our conditional check ...
		if curr[m].timestamp <= timestamp {
			res = curr[m].value
			l = m + 1
		} else {
			r = m - 1
		}
	}

	return res
}

/**
 * Your TimeMap object will be instantiated and called as such:
 * obj := Constructor();
 * obj.Set(key,value,timestamp);
 * param_2 := obj.Get(key,timestamp);
 */
