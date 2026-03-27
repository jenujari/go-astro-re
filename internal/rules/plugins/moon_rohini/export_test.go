package moon_rohini

import "local.io/go-astro-re/internal/domain"

func RuleForTest() rule {
	return rule{}
}

func RuleMetadataForTest() domain.RuleMetadata {
	return rule{}.Metadata()
}
