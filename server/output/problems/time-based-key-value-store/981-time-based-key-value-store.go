package time_based_key_value_store

/*
	every "set" will be of a greater timestamp of the previous invocation
	so our timestamps for a given key will always be in sorted order ...

	This allows us to keep a map between [key] -> [value, timestamp]

	and we can binary search for a given timestamp on the `Get` if it exists ...
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
	this.store[key] = append(this.store[key], pair{value, timestamp})
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
