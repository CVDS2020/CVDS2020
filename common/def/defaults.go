package def

func Defaults[X comparable](v X, defs ...X) X {
	var zero X
	for _, def := range defs {
		if v == zero {
			v = def
		} else {
			return v
		}
	}
	return v
}

func SetDefaults[X comparable](vp *X, defs ...X) {
	var zero X
	for _, def := range defs {
		if *vp == zero {
			*vp = def
		} else {
			return
		}
	}
	return
}
