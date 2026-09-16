package credentials

import "os"

func osLookupEnv(name string) (string, bool) {
	return os.LookupEnv(name)
}
