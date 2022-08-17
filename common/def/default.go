package def

func Condition[X any](condition bool, yes X, no X) X {
	if condition {
		return yes
	} else {
		return no
	}
}

func ConditionBySetter[X any](condition bool, yes func() X, no func() X) X {
	if condition {
		return yes()
	} else {
		return no()
	}
}

func Default[X comparable](v X, def X) X {
	var zero X
	if v == zero {
		return def
	}
	return v
}

func SetDefault[X comparable](vp *X, def X) {
	var zero X
	if *vp == zero {
		*vp = def
	}
}

func DefaultBySetter[X comparable](v X, setter func() X) X {
	var zero X
	if v == zero {
		return setter()
	}
	return v
}

func SetDefaultBySetter[X comparable](vp *X, setter func() X) {
	var zero X
	if *vp == zero {
		*vp = setter()
	}
}

func DefaultIf[X any](v X, def X, condition func(v X) bool) X {
	if condition(v) {
		return def
	}
	return v
}

func SetDefaultIf[X any](vp *X, def X, condition func(v X) bool) {
	if condition(*vp) {
		*vp = def
	}
}

func DefaultIfBySetter[X any](v X, setter func() X, condition func(v X) bool) X {
	if condition(v) {
		return setter()
	}
	return v
}

func SetDefaultIfBySetter[X any](vp *X, setter func() X, condition func(v X) bool) {
	if condition(*vp) {
		*vp = setter()
	}
}

func DefaultIfEqual[X comparable](v X, def X, ref X) X {
	if v == ref {
		return def
	}
	return v
}

func SetDefaultIfEqual[X comparable](vp *X, def X, ref X) {
	if *vp == ref {
		*vp = def
	}
}

func DefaultIfEqualBySetter[X comparable](v X, setter func() X, ref X) X {
	if v == ref {
		return setter()
	}
	return v
}

func SetDefaultIfEqualBySetter[X comparable](vp *X, setter func() X, ref X) {
	if *vp == ref {
		*vp = setter()
	}
}

func DefaultIfNotEqual[X comparable](v X, def X, ref X) X {
	if v != ref {
		return def
	}
	return v
}

func SetDefaultIfNotEqual[X comparable](vp *X, def X, ref X) {
	if *vp != ref {
		*vp = def
	}
}

func DefaultIfNotEqualBySetter[X comparable](v X, setter func() X, ref X) X {
	if v != ref {
		return setter()
	}
	return v
}

func SetDefaultIfNotEqualBySetter[X comparable](vp *X, setter func() X, ref X) {
	if *vp != ref {
		*vp = setter()
	}
}
