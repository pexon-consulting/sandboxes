package utils

import (
	"regexp"
)

func ProofPexonMail(s string) bool {
	var validPexonMail = regexp.MustCompile(`\w+\.\w+\@pexon-consulting\.de`)
	return validPexonMail.MatchString(s)
}

func Lease_time_Input(s string) bool {
	var validPexonMail = regexp.MustCompile(`(201[4-9]|202[0-9])-(0[1-9]|1[0-2])-(0[1-9]|1[0-9]|2[0-9]|3[0-1])`)
	return validPexonMail.MatchString(s)
}

func ValidateIsoTime(s string) bool {
	time := regexp.MustCompile(`(201[4-9]|20\d{2})-(0[0-9]|1[0-2])-([02]+[0-9]|3[01])T(00|0\d|1\d|2[0123]):([0-5]+\d):([0-5]+\d)Z`)
	return time.MatchString(s)
}
