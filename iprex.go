package j8a

import "regexp"

//TODO: rather not use multiple regexes in a chain and replace this with a tested thirdparty framework.
type iprex struct{
	//v6 hex with embedded v4 dec address
	ipv64e *regexp.Regexp

	//v6 simple
	ipv6s *regexp.Regexp

	//v6 extended
	ipv6 *regexp.Regexp

	//v4 addresses
	ipv4 *regexp.Regexp
}

func (ipr *iprex) extractAddr(msg string) string {
	addr := "not found"
	if ipr.ipv6s == nil {
		r2, _ := regexp.Compile("([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}")
		ipr.ipv6s = r2
	}
	if ipr.ipv6 == nil {
		r1, _ := regexp.Compile("(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))")
		ipr.ipv6 = r1
	}
	if ipr.ipv64e == nil {
		r, _ := regexp.Compile("(:{0,2})([0-9a-fA-F]{1,4}:){1,6}(:{0,1})((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])")
		ipr.ipv64e = r
	}
	if ipr.ipv4 == nil {
		r, _ := regexp.Compile("((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])")
		ipr.ipv4 = r
	}
	match := ipr.ipv64e.FindStringSubmatch(msg)
	if len(match)>0 {
		addr=match[0]
	} else {
		match2 := ipr.ipv6s.FindStringSubmatch(msg)
		if len(match2)>0 {
			addr=match2[0]
		} else {
			match3 := ipr.ipv6.FindStringSubmatch(msg)
			if len(match3)>0 {
				addr=match3[0]
			} else {
				match4 := ipr.ipv4.FindStringSubmatch(msg)
				if len(match4)>0 {
					addr=match4[0]
				}
			}
		}
	}
	return addr
}
