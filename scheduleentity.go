package crontabparser

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type scheduleEntity struct {
	raw string
	typ scheduleType

	expanded []int
}

func (se *scheduleEntity) Raw() string {
	return se.raw
}

func (se *scheduleEntity) Type() scheduleType {
	return se.typ
}

func (se *scheduleEntity) Expanded() []int {
	return se.expanded
}

func (se *scheduleEntity) Match(num int) bool {
	if se == nil {
		return false
	}
	for _, i := range se.expanded {
		if num == i {
			return true
		}
	}
	return false
}

//go:generate stringer -type=scheduleType -trimprefix Schedule
type scheduleType int

const (
	scheduleMinute scheduleType = iota
	scheduleHour
	scheduleDay
	scheduleMonth
	scheduleDayOfWeek
)

type entityParam struct {
	Range   [2]int
	Aliases []string
}

var entityParams = map[scheduleType]entityParam{
	scheduleMinute: {
		Range: [2]int{0, 59},
	},
	scheduleHour: {
		Range: [2]int{0, 23},
	},
	scheduleDay: {
		Range: [2]int{1, 31},
	},
	scheduleMonth: {
		Range:   [2]int{1, 12},
		Aliases: []string{"", "jan", "feb", "mar", "apr", "may", "jun", "jul", "aug", "sep", "oct", "nov", "dec"},
	},
	scheduleDayOfWeek: {
		Range:   [2]int{0, 7},
		Aliases: []string{"sun", "mon", "tue", "wed", "thu", "fri", "sat"},
	},
}

func newScheduleEntity(raw string, st scheduleType) (*scheduleEntity, error) {
	se := &scheduleEntity{
		raw: raw,
		typ: st,
	}
	err := se.init()
	if err != nil {
		return nil, err
	}
	return se, nil
}

func (se *scheduleEntity) init() error {
	ep, ok := entityParams[se.typ]
	if !ok {
		return fmt.Errorf("no entity param setting for %s", se.typ)
	}
	entity := strings.ToLower(se.raw)
	for i, v := range ep.Aliases {
		if v == "" {
			continue
		}
		entity = strings.Replace(entity, v, fmt.Sprintf("%d", i), -1)
	}
	var expanded []int
	for _, item := range strings.Split(entity, ",") {
		if stuffs := strings.SplitN(item, "/", 2); len(stuffs) == 2 {
			rng, err := parseRange(stuffs[0], ep.Range)
			if err != nil {
				return fmt.Errorf("invalid entity: %s, %s", se.raw, err)
			}
			increments, err := strconv.ParseUint(stuffs[1], 10, 64)
			if err != nil || increments == 0 {
				return fmt.Errorf("invalid increments: %q in %q", stuffs[1], se.raw)
			}
			incr := int(increments)
			incrCounter := 0
			for i := rng[0]; i <= rng[1]; i++ {
				if incrCounter%incr == 0 {
					expanded = append(expanded, i)
				}
				incrCounter++
			}
		} else {
			if n, err := strconv.ParseUint(item, 10, 64); err == nil {
				num := int(n)
				if num < ep.Range[0] || num > ep.Range[1] {
					return fmt.Errorf("invalid entity: %s", se.raw)
				}
				expanded = append(expanded, num)
			} else {
				rng, err := parseRange(item, ep.Range)
				if err != nil {
					return fmt.Errorf("invalid entity: %s, %s", se.raw, err)
				}
				for i := rng[0]; i <= rng[1]; i++ {
					expanded = append(expanded, i)
				}
			}
		}
	}

	if se.typ == scheduleDayOfWeek {
		hasSun := false
		for _, v := range expanded {
			if v == 7 {
				hasSun = true
			}
		}
		if hasSun {
			expanded = append(expanded, 0)
		}
	}

	seen := make(map[int]struct{})
	var uniqness []int
	for _, v := range expanded {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			uniqness = append(uniqness, v)
		}
	}
	sort.Ints(uniqness)
	se.expanded = uniqness
	return nil
}

var rangeReg = regexp.MustCompile(`^(\d{1,2})-(\d{1,2})$`)

func parseRange(item string, rng [2]int) (ret [2]int, err error) {
	if item == "*" {
		return [2]int{rng[0], rng[1]}, nil
	}
	if m := rangeReg.FindStringSubmatch(item); len(m) == 3 {
		mi, _ := strconv.ParseInt(m[1], 10, 64)
		min := int(mi)
		ma, _ := strconv.ParseInt(m[2], 10, 64)
		max := int(ma)
		if min >= max || min < rng[0] || max > rng[1] {
			return ret, fmt.Errorf("invalid range: %s", item)
		}
		return [2]int{min, max}, nil
	}
	return ret, fmt.Errorf("invalid range: %s", item)
}
