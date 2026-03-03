package homework01

import (
	"fmt"
	"strconv"
	"strings"
	"sort"
)
// 1. 只出现一次的数字
// 给定一个非空整数数组，除了某个元素只出现一次以外，其余每个元素均出现两次。找出那个只出现了一次的元素。

func SingleNumber(nums []int) int {
	m :=make(map[int]int)
	for i:=0;i<len(nums);i++ {
		// fmt.Println(nums[i])
		if i==0 {
			m[nums[i]] = 1
			continue
		}
		count ,exists := m[nums[i]]
		if exists{
			m[nums[i]] = count + 1
		}else{
			m[nums[i]] = 1
		}
	}
	fmt.Println(m)
	for key,value := range m{
		if value == 1{
			return key
		}
	}

	return 0
}

// 2. 回文数
// 判断一个整数是否是回文数
func IsPalindrome(x int) bool {
	if x < 0 || (x>9 && x%10==0){
		return false
	}else{
		bbb:= strconv.Itoa(x)
		length := len(bbb)
		for i,j := 0,length-1;i<length/2;i,j = i+1,j-1 {
			if bbb[i] != bbb[j]{
				return false
			}
		}
	}
	return true

}

// 3. 有效的括号
// 给定一个只包括 '(', ')', '{', '}', '[', ']' 的字符串，判断字符串是否有效
func IsValid(s string) bool {
	matchMap := map[rune]rune{
		')':'(',
		'}':'{',
		']':'[',
	}
	stack := []rune{}
	for i:=0;i<len(s);i++ {
		if s[i]=='(' || s[i]=='{' || s[i]=='[' {
			stack = append(stack, rune(s[i]))
		} else if s[i]==')' || s[i]=='}' || s[i]==']' {
			if len(stack)==0 {
				return false
			}
			top := stack[len(stack)-1]
			if matchMap[rune(s[i])] != top {
				return false
			}
			stack = stack[:len(stack)-1]
		}
		
	}
	return len(stack)==0
}

// 4. 最长公共前缀
// 查找字符串数组中的最长公共前缀
func LongestCommonPrefix(strs []string) string {
	if len(strs)==0 {
		return ""
	}
	if len(strs)==1 {
		return strs[0]
	}
	prefix := strs[0]
	for i:=1;i<len(strs);i++ {
		for !strings.HasPrefix(strs[i], prefix) {
			prefix = prefix[:len(prefix)-1]
			if prefix == "" {
				return ""
			}
		}
	}
	return prefix

}

// 5. 加一
// 给定一个由整数组成的非空数组所表示的非负整数，在该数的基础上加一
func PlusOne(digits []int) []int {
	// DONE: implement
	n := len(digits)
	for i := n - 1; i >= 0; i-- {
		digits[i]++
		digits[i] %= 10
		if digits[i] != 0 {
			return digits
		}
	}
	newDigits := make([]int, n+1)
	newDigits[0] = 1
	return newDigits
}

// 6. 删除有序数组中的重复项
// 给你一个有序数组 nums ，请你原地删除重复出现的元素，使每个元素只出现一次，返回删除后数组的新长度。
// 不要使用额外的数组空间，你必须在原地修改输入数组并在使用 O(1) 额外空间的条件下完成。
func RemoveDuplicates(nums []int) int {
	// TODO: implement
	// return 0
	if len(nums) == 0 {
		return 0
	}
	slow := 0
	for fast := 1; fast < len(nums); fast++ {
		if nums[fast] != nums[slow] {
			slow++
			// fmt.Println(nums[slow], nums[fast])
			nums[slow] = nums[fast]
			// fmt.Println(nums[slow], nums[fast])
		}
	}
	fmt.Println(nums)
	return slow + 1
}

// 7. 合并区间
// 以数组 intervals 表示若干个区间的集合，其中单个区间为 intervals[i] = [starti, endi] 。
// 请你合并所有重叠的区间，并返回一个不重叠的区间数组，该数组需恰好覆盖输入中的所有区间。
func Merge(intervals [][]int) [][]int {
	// DONE: implement
	if len(intervals) == 0 {
		return nil
	}
	// 按照区间的起始位置进行排序
	sort.Slice(intervals, func(i, j int) bool {
		return intervals[i][0] < intervals[j][0]
	})
	result := [][]int{intervals[0]}
	for i := 1; i < len(intervals); i++ {
		last := result[len(result)-1]
		if intervals[i][0] <= last[1] {
			last[1] = max(last[1], intervals[i][1])
		} else {
			result = append(result, intervals[i])
		}
	}
	return result
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// 8. 两数之和
// 给定一个整数数组 nums 和一个目标值 target，请你在该数组中找出和为目标值的那两个整数
func TwoSum(nums []int, target int) []int {
	// DONE: implement
	for i := 0; i < len(nums); i++ {
		sub := target - nums[i]
		for j := i + 1; j < len(nums); j++ {
			// fmt.Println(nums[i],nums[j])
			if nums[j] == sub {
				return []int{i, j}
			}
		}
	}
	return nil
}
