package steamkeygrep

type keyCache map[string]string

func (kc keyCache) handleKeyCache(keys <-chan string, restartCurrentGrep chan<- bool, errs chan<- error) {
	for key := range keys {
		if _, ok := kc[key]; ok {
			for k := range kc {
				delete(kc, k)
			}
			kc[key] = key
			restartCurrentGrep <- true
		}
	}
}
