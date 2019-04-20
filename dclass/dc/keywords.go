package dc

type KeywordList struct {
	keywords []string
}

func (k *KeywordList) AddKeyword(kw string) {
	k.keywords = append(k.keywords, kw)
}

func (k *KeywordList) HasKeyword(kw string) bool {
	for _, key := range k.keywords {
		if key == kw {
			return true
		}
	}
	return false
}

func (k *KeywordList) HasMatchingKeywords(other KeywordList) bool {
	if len(k.keywords) != len(other.keywords) {
		return false
	}

	for n, key := range k.keywords {
		if other.keywords[n] != key {
			return false
		}
	}

	return true
}

func (k *KeywordList) Copy(other KeywordList) {
	k.keywords = append(other.keywords[:0:0], other.keywords...)
}

func (k *KeywordList) GenerateHash(generator *HashGenerator) {
	keywords := map[string]int{
		"required":  0x0001,
		"broadcast": 0x0002,
		"ownrecv":   0x0004,
		"ram":       0x0008,
		"db":        0x0010,
		"clsend":    0x0020,
		"clrecv":    0x0040,
		"ownsend":   0x0080,
		"airecv":    0x0100,
	}

	flags := 0
	for _, kw := range k.keywords {
		if _, ok := keywords[kw]; ok {
			flags |= keywords[kw]
		} else {
			flags = ^0
		}
	}

	if flags != ^0 {
		generator.AddInt(flags)
	} else {
		generator.AddInt(len(k.keywords))
		for _, kw := range k.keywords {
			generator.AddString(kw)
		}
	}

}
