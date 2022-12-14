package container

import "github.com/CVDS2020/CVDS2020/common/generic"

type Comparable interface {
	Compare(other Comparable) int
}

func OrderedCompare[Ordered generic.Ordered](n1, n2 Ordered) int {
	switch {
	case n1 < n2:
		return -1
	case n1 > n2:
		return 1
	default:
		return 0
	}
}
