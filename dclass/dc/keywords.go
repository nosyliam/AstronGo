package dc

type KeywordList struct {
	keywords map[string]struct{}
}

func (k KeywordList) AddKeyword(kw string) {
	k.keywords[kw] = struct{}{}
}

func (k KeywordList) HasKeyword(kw string) bool {
	if _, ok := k.keywords[kw]; ok {
		return true
	}
	return false
}

func (k KeywordList) HasMatchingKeywords(other KeywordList) bool {
	if len(k.keywords) != len(other.keywords) {
		return false
	}

	for key := range k.keywords {
		if _, ok := other.keywords[key]; !ok {
			return false
		}
	}

	return true
}

func (k KeywordList) Copy(other KeywordList) {
	for key := range k.keywords {
		k.keywords[key] = struct{}{}
	}
}

func (k KeywordList) GenerateHash(generator HashGenerator) {
	// TODO
}