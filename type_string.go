// Code generated by "stringer -type=Type -trimprefix Type"; DO NOT EDIT.

package crontab

import "strconv"

const _Type_name = "ErrorJobCommentEmptyEnv"

var _Type_index = [...]uint8{0, 5, 8, 15, 20, 23}

func (i Type) String() string {
	if i < 0 || i >= Type(len(_Type_index)-1) {
		return "Type(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Type_name[_Type_index[i]:_Type_index[i+1]]
}
