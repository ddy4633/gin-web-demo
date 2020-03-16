package main

import "fmt"

func main() {
	a := []int{2,5,5,11}
	fmt.Println(twoSum(a,10))
}

func twoSum(nums []int, target int) []int {
	var b []int
	for i:=0;i<len(nums);i++{
		for a:=i+1;a<len(nums);a++ {
			fmt.Println(i,a)
			if (target == nums[a]+nums[i]){
				b = append(b,i)
				b = append(b,a)
				return b
			}
		}
	}
	return b
}
