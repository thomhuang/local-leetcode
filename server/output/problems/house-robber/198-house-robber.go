package house_robber

func rob(nums []int) int {
	res := 0
	robbed := make([]int, len(nums))

	for house, money := range nums {
		robbed[house] = money

		bestRobbery := 0
		for i := house - 2; i >= 0 && i >= house-3; i-- {
			bestRobbery = max(bestRobbery, robbed[i])
		}

		robbed[house] += bestRobbery
		res = max(res, robbed[house])
	}

	return res
}
