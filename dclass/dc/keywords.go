package dc

type KeywordList map[string]struct{}

func (k KeywordList) AddKeyword(kw string) {
	k[kw] = struct{}{}
}

func (k KeywordList) HasKeyword(kw string) bool {
	if _, ok := k[kw]; ok {
		return true
	}
	return false
}

func (k KeywordList) HasMatchingKeywords(other KeywordList) bool {
	if len(k) != len(other) {
		return false
	}

	for key := range k {
		if _, ok := other[key]; !ok {
			return false
		}
	}

	return true
}

func (k KeywordList) Copy(other KeywordList) {
	for key := range k {
		k[key] = struct{}{}
	}
}

func (k KeywordList) GenerateHash(generator HashGenerator) {
	// TODO
}