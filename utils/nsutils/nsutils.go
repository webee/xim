package nsutils

// EncodeNSUser encode ns user.
func EncodeNSUser(ns, u string) (user string) {
	if ns == "" {
		return u
	}
	return ns + ":" + u
}
