package handlers

type Credentials struct {
	user string
	pass string
}

func BasicAuth(user string, pass string) (interface{}, error) {
	return Credentials{user: user, pass: pass}, nil
}
