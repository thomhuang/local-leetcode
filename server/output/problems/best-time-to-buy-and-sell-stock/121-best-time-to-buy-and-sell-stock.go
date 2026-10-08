package best_time_to_buy_and_sell_stock

import "math"

func maxProfit(prices []int) int {
	/*
		ez pz, keep track of lowest price
		as we move through prices, and subtract
		from the current price ...

		if profit exceeds most profit so far, then update!
	*/

	minPrice := math.MaxInt32
	res := 0
	for _, price := range prices {
		// if price < minPrice, then there's no reason
		// for us to calculate the profit, as it'd be 0,
		// or negative ...
		if price < minPrice {
			minPrice = price
			continue
		}
		res = max(res, price-minPrice)
	}

	return res
}
