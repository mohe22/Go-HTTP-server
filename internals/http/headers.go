package http

type Header map[string]string

func NewHeader() *Header {
	h := make(Header)
	return &h
}

func (h *Header) Get(key string) (string, error) {
	if value, ok := (*h)[key]; ok {
		return value, nil
	}
	return "", ErrKeyNotFound
}

func (h *Header) Set(key, value string) error {
	if key == "" {
		return ErrEmptyKey
	}
	(*h)[key] = value
	return nil
}
func (h *Header) Replace(key, value string) error {
	if key == "" {
		return ErrEmptyKey
	}
	if _, exists := (*h)[key]; exists {
		(*h)[key] = value
	}
	return ErrKeyNotFound
}

func (h Header) ForEach(f func(key, value string)) {
	for k, v := range h {
		f(k, v)
	}
}
func (h *Header) Delete(key string) error {
	if key == "" {
		return ErrEmptyKey
	}
	if _, exists := (*h)[key]; !exists {
		return ErrKeyNotFound
	}
	delete(*h, key)
	return nil
}
