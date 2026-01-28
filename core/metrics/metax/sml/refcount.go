package sml

/*
 * library & refcount
 */

type refcount int32

func (r *refcount) IncOnNoError(err error) {
	if err == nil {
		*r++
	}
}

func (r *refcount) DecOnNoError(err error) {
	if err == nil && *r > 0 {
		*r--
	}
}
