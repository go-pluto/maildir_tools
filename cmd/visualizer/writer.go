package main

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

func matplotlibWriter(w io.Writer, results map[string]map[string]int) error {
	if len(results) == 0 {
		return nil
	}

	var times []string
	for timestamp := range results {
		times = append(times, timestamp)
	}

	sort.Slice(times, func(i, j int) bool {
		return times[i] < times[j]
	})

	// Deduplicate users
	usersMap := make(map[string]bool)
	for _, val := range results {
		for user := range val {
			usersMap[user] = true
		}
	}

	var users []string
	for user := range usersMap {
		users = append(users, user)
	}

	sort.Slice(users, func(i, j int) bool {
		return users[i] < users[j]
	})

	fmt.Fprintf(w, "t = [%s]\n", strings.Join(times, ", "))

	for i, user := range users {
		var vals []string
		for _, time := range times {
			if val, ok := results[time][user]; ok {
				vals = append(vals, fmt.Sprintf("%d", val))
			} else {
				vals = append(vals, "None")
			}
		}

		fmt.Fprintf(w, "s%d = [%s]\n", i, strings.Join(vals, ", "))
		fmt.Fprintf(w, "plot.plot(t, s%d)\n", i)
	}

	return nil
}

//func matplotlibLegendWriter(w io.Writer, results []Result) error {
//	labels := []string{}
//	for _, result := range results {
//		labels = append(labels, fmt.Sprintf("'%s'", result.Metric))
//	}
//
//	fmt.Fprintf(w, "plot.legend([%s], loc='upper left')\n", strings.Join(labels, ", "))
//
//	return nil
//}
