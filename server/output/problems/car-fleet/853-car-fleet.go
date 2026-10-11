package car_fleet

import (
	"sort"
)

/*
What makes sense in my head at first, is to sort our cars by their positions first ...
as that's the order in which we have to process the cars to calculate the fleets properly

On top of that, notice that, for each car, we calculate the time taken to the target,
we know that a car behind another car joins it to become a fleet, if the time taken to the
target is <= the time taken for the car ahead ...
*/
func carFleet(target int, position []int, speed []int) int {
	n := len(position)
	cars := make([][2]float64, n)
	// keep (car_pos, car_speed) together ...
	for i := range n {
		// we have to take the float as we can't round up/down as that'd merge
		// fleets that shouldn't be merged ...
		timeTaken := float64(target-position[i]) / float64(speed[i])
		cars[i] = [2]float64{float64(position[i]), timeTaken}
	}

	// sort cars by position descending
	sort.Slice(cars, func(i, j int) bool {
		return cars[i][0] > cars[j][0]
	})

	// keeps track of fleets ...
	// this will indicate the time taken to the target
	stack := make([]float64, 0)
	for _, car := range cars {
		if len(stack) == 0 {
			stack = append(stack, car[1])
		} else {
			// peek at latest fleet ...
			fleetTime := stack[len(stack)-1]
			currTime := car[1]

			// if currTime <= fleetTime, then they join up ...
			// if >, then it becomes its own fleet ...
			if currTime > fleetTime {
				stack = append(stack, car[1])
			}
		}
	}

	return len(stack)
}
